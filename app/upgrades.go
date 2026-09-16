package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	circuittypes "cosmossdk.io/x/circuit/types"
	"cosmossdk.io/x/nft"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	epochstypes "github.com/cosmos/cosmos-sdk/x/epochs/types"
	"github.com/cosmos/cosmos-sdk/x/group"

	//minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	liquidtypes "github.com/cosmos/gaia/v27/x/liquid/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"
	icatypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
	"github.com/fetchai/fetchd/app/ica_migration"
	"github.com/fetchai/fetchd/app/traces"
	"github.com/strangelove-ventures/tokenfactory/x/tokenfactory/keeper"
	tokenfactorytypes "github.com/strangelove-ventures/tokenfactory/x/tokenfactory/types"
)

// ============================================================================
// UPGRADE PLAN NAMES
// ============================================================================

// UpgradeNameV0_15_0 is the plan name of the PREVIOUS upgrade ("v0.15.0").
// It is already applied on-chain; its handler is kept only so that this
// binary can still correctly execute that upgrade if it is ever pending.
const UpgradeNameV0_15_0 = "v0.15.0"

// UpgradeNameV0_15_1 is the plan name of the CURRENT UPCOMING upgrade.
const UpgradeNameV0_15_1 = "v0.15.1"

// RegisterUpgradeHandlers registers the software upgrade handlers.
//
// The code is split into two clearly separated sections below:
// one for the previous upgrade (v0.15.0) and one for the upcoming
// upgrade (v0.15.1).
func (app *App) RegisterUpgradeHandlers(cfg module.Configurator) {
	app.registerV0_15_0UpgradeHandler(cfg)
	app.registerV0_15_1UpgradeHandler(cfg)
	app.setStoreLoaderForPendingV0_15_0Upgrade()
}

// ============================================================================
// PREVIOUS UPGRADE: v0.15.0  ("v0.15.0")
// ============================================================================
//
// Everything in this section belongs exclusively to the previous v0.15.0
// upgrade. It is kept ONLY so that this binary can still perform that
// upgrade if its plan is pending on some node (or during a re-run of it).
//
// >>> CLEANUP CHECKLIST — for the upgrade AFTER v0.15.1: <<<
// Once v0.15.1 has been applied on-chain, the v0.15.0 plan can never be
// pending again, and this whole section can be deleted. That means, in the
// NEXT release (the one after the v0.15.1 release):
//
//   1. Delete the entire "PREVIOUS UPGRADE: v0.15.0" section below
//      (registerV0_15_0UpgradeHandler, v0_15_0StoreUpgrades,
//       setStoreLoaderForPendingV0_15_0Upgrade,
//       migrateConsensusParamsFromParamsStore, parseI64).
//   2. Delete the registerV0_15_0UpgradeHandler and
//      setStoreLoaderForPendingV0_15_0Upgrade calls in
//      RegisterUpgradeHandlers above.
//   3. Remove the now-unused imports (fmt, json, strconv, strings, time,
//      math, storetypes, circuittypes, nft, tmproto, sdk, module's
//      Configurator stays if still needed, consensusparamtypes,
//      epochstypes, group, paramstypes, protocolpooltypes, stakingtypes,
//      liquidtypes, ica* / ibc* types, ica_migration, traces, tokenfactory
//      imports — let the compiler/gopls prune them).
//   4. At that point, keep the (still registered) empty v0.15.1 handler
//      as the "previous upgrade" handler for one more release, per the
//      convention of retaining the previous upgrade's handler.
// ============================================================================

// v0_15_0StoreUpgrades lists all new/renamed/deleted KV stores at the
// v0.15.0 upgrade height.
var v0_15_0StoreUpgrades = storetypes.StoreUpgrades{
	Added: []string{
		protocolpooltypes.StoreKey,
		circuittypes.StoreKey,
		epochstypes.StoreKey,
		consensusparamtypes.StoreKey,
		group.StoreKey,
		icacontrollertypes.StoreKey,
		nft.StoreKey,
		tokenfactorytypes.StoreKey,
		liquidtypes.StoreKey,
	},
	Renamed: []storetypes.StoreRename{
		// {OldKey: "oldkey", NewKey: "newkey"},
	},
	Deleted: []string{
		// Only if truly removed and fully migrated:
		// "params",
	},
}

// registerV0_15_0UpgradeHandler registers the v0.15.0 upgrade handler,
// which runs the per-module Migrators and the custom state migrations of
// the v0.15.0 upgrade.
func (app *App) registerV0_15_0UpgradeHandler(cfg module.Configurator) {
	app.UpgradeKeeper.SetUpgradeHandler(
		UpgradeNameV0_15_0,
		func(ctx context.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			sdkCtx := sdk.UnwrapSDKContext(ctx)

			// Migrate ICA controller
			ICAmigrator := ica_migration.NewICAMigrator(sdkCtx, sdkCtx.KVStore(app.GetKey(icacontrollertypes.StoreKey)), sdkCtx.KVStore(app.GetKey(ibcexported.StoreKey)), sdkCtx.KVStore(app.GetMemKey(CapabilityMemStoreKey)), app.GetKey(CapabilityStoreKey))
			err := cfg.RegisterMigration(icatypes.ModuleName, 1, ICAmigrator.AssertChannelCapabilityMigrations)
			if err != nil {
				return nil, err
			}

			// Migrate transfer traces
			m := traces.NewDenomTracesMigrator(sdkCtx.KVStore(app.GetKey(ibctransfertypes.StoreKey)))
			err = cfg.RegisterMigration(ibctransfertypes.ModuleName, 1, m.MigrateTraces)
			if err != nil {
				return nil, err
			}

			err = migrateConsensusParamsFromParamsStore(app, sdkCtx)
			if err != nil {
				return nil, err
			}

			res, err := app.mm.RunMigrations(ctx, cfg, fromVM)
			if err != nil {
				return res, err
			}

			// Bootstrap liquid staking
			err = app.StakingKeeper.IterateValidators(ctx, func(_ int64, v stakingtypes.ValidatorI) (stop bool) {
				lv := liquidtypes.LiquidValidator{
					OperatorAddress: v.GetOperator(),
					LiquidShares:    math.LegacyZeroDec(),
				}
				err := app.LiquidKeeper.SetLiquidValidator(ctx, lv)
				if err != nil {
					return false
				}
				return false
			})
			if err != nil {
				return nil, err
			}

			bondDenom, err := app.StakingKeeper.BondDenom(sdkCtx)
			if err != nil {
				return nil, err
			}

			type DenomAdmin struct {
				Denom   string
				Address string
			}

			type ChainConfig struct {
				Admins []DenomAdmin
				Params tokenfactorytypes.Params
			}

			defaultDenomCreationGasConsume := uint64(2000000)
			defaultDenomCreationFee := sdk.NewCoins(sdk.NewCoin(bondDenom, math.NewIntWithDecimal(1, 18)))

			defaultParams := tokenfactorytypes.Params{
				DenomCreationFee:        defaultDenomCreationFee,
				DenomCreationGasConsume: defaultDenomCreationGasConsume,
			}

			var chainConfig ChainConfig

			switch sdkCtx.ChainID() {
			case "fetchhub-4":
				chainConfig = ChainConfig{
					Admins: []DenomAdmin{
						// Mainnet bridge contract
						{Denom: bondDenom, Address: "fetch1qxxlalvsdjd07p07y3rc5fu6ll8k4tmetpha8n"},
					},
					Params: defaultParams,
				}
			case "dorado-1":
				chainConfig = ChainConfig{
					Admins: []DenomAdmin{
						// Dorado testnet bridge contract
						{Denom: bondDenom, Address: "fetch182q50y030ctp39dkjhv4pn95h9vxg29s67djtr0560fuwprtks0sfrtyz0"},
					},
					Params: defaultParams,
				}
			default:
				feeAmount := math.NewIntWithDecimal(1, 9)

				if metadata, ok := app.BankKeeper.GetDenomMetaData(ctx, bondDenom); ok {
					var maxExp uint32
					var found bool

					for _, du := range metadata.DenomUnits {
						if !found || du.Exponent > maxExp {
							maxExp = du.Exponent
							found = true
						}
					}

					if found {
						feeAmount = math.NewIntWithDecimal(1, int(maxExp))
					}
				}

				blockMaxGas := sdkCtx.ConsensusParams().Block.MaxGas

				denomCreationGasConsume := uint64(0)
				// sanity check
				if blockMaxGas > 0 {
					denomCreationGasConsume = uint64(blockMaxGas) / 3
				}

				chainConfig = ChainConfig{
					Params: tokenfactorytypes.Params{
						DenomCreationFee:        sdk.NewCoins(sdk.NewCoin(bondDenom, feeAmount)),
						DenomCreationGasConsume: denomCreationGasConsume,
					},
				}
			}

			app.TokenFactoryKeeper.SetParams(sdkCtx, chainConfig.Params)

			udc := keeper.NewUnboundDenomCreator(app.TokenFactoryKeeper)
			for _, denomAdmin := range chainConfig.Admins {
				err = udc.CreateDenom(sdkCtx, denomAdmin.Address, denomAdmin.Denom)
				if err != nil {
					return nil, err
				}
			}

			return res, err
		},
	)
}

// setStoreLoaderForPendingV0_15_0Upgrade installs the custom store loader
// needed by the v0.15.0 upgrade (it mounts new KV stores).
//
// This must run during app construction — before LoadLatestVersion — and
// therefore CANNOT live inside the upgrade handler. It is a no-op unless
// the v0.15.0 plan is currently pending on disk (no upgrade-info.json file
// or a different plan name => default store loader is used).
//
// NOTE: this reads the pending plan from disk on every startup, but only
// actually overrides the store loader for the v0.15.0 plan. It belongs to
// the v0.15.0 section and is part of the cleanup checklist above.
func (app *App) setStoreLoaderForPendingV0_15_0Upgrade() {
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
	}
	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	// The v0.15.0 plan needs the new stores mounted.
	if upgradeInfo.Name == UpgradeNameV0_15_0 {
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &v0_15_0StoreUpgrades))
	}
}

// migrateConsensusParamsFromParamsStore migrates the consensus params from
// the legacy x/params store into the x/consensus module (part of the
// v0.15.0 upgrade).
func migrateConsensusParamsFromParamsStore(app *App, ctx sdk.Context) error {
	paramsStore := ctx.KVStore(app.GetKey(paramstypes.StoreKey))

	// BlockParams
	var blockParams tmproto.BlockParams
	if bz := paramsStore.Get([]byte("baseapp/BlockParams")); len(bz) > 0 {
		// these are plain numbers; codec JSON is fine
		if err := app.AppCodec().UnmarshalJSON(bz, &blockParams); err != nil {
			return err
		}
	}

	// --- EvidenceParams (strings; duration in ns) ---
	var epRaw struct {
		MaxAgeNumBlocks string `json:"max_age_num_blocks"`
		MaxAgeDuration  string `json:"max_age_duration"` // ns as string
		MaxBytes        string `json:"max_bytes"`
	}
	if err := json.Unmarshal(paramsStore.Get([]byte("baseapp/EvidenceParams")), &epRaw); err != nil {
		return err
	}
	blocks, err := parseI64(epRaw.MaxAgeNumBlocks)
	if err != nil {
		return err
	}
	durNs, err := parseI64(epRaw.MaxAgeDuration)
	if err != nil {
		return err
	}
	evMaxBytes, err := parseI64(epRaw.MaxBytes)
	if err != nil {
		return err
	}
	evidenceParams := tmproto.EvidenceParams{
		MaxAgeNumBlocks: blocks,
		MaxAgeDuration:  time.Duration(durNs), // ns → time.Duration
		MaxBytes:        evMaxBytes,
	}

	// ValidatorParams
	var validatorParams tmproto.ValidatorParams
	if bz := paramsStore.Get([]byte("baseapp/ValidatorParams")); len(bz) > 0 {
		if err := app.AppCodec().UnmarshalJSON(bz, &validatorParams); err != nil {
			return err
		}
	}

	cp := tmproto.ConsensusParams{
		Block:     &blockParams,
		Evidence:  &evidenceParams,
		Validator: &validatorParams,
		Version:   &tmproto.VersionParams{App: 0},
	}

	// Write to x/consensus so it’s persisted in app state
	if err := app.ConsensusParamsKeeper.ParamsStore.Set(ctx, cp); err != nil {
		return err
	}
	// Also write to BaseApp so CometBFT uses it immediately
	if err := app.BaseApp.StoreConsensusParams(ctx, cp); err != nil {
		return err
	}
	return nil
}

// parseI64 is a helper of migrateConsensusParamsFromParamsStore (v0.15.0).
func parseI64(s string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(s), 10, 64)
}

// ============================================================================
// CURRENT UPCOMING UPGRADE: v0.15.1  ("v0.15.1")
// ============================================================================
//
// The v0.15.1 upgrade requires no state migrations and no store shape
// changes (no added/renamed/deleted KV stores), hence:
//   - the handler below is empty (no-op), and
//   - no custom store loader is set up for it (the default one is used).
//
// Once v0.15.1 has been applied on-chain, keep this empty handler
// registered for one more release (as the "previous upgrade" handler),
// then it can be removed alongside the v0.15.0 cleanup described above.
// ============================================================================

// registerV0_15_1UpgradeHandler registers the empty (no-op) v0.15.1
// upgrade handler.
func (app *App) registerV0_15_1UpgradeHandler(cfg module.Configurator) {
	app.UpgradeKeeper.SetUpgradeHandler(
		UpgradeNameV0_15_1,
		func(ctx context.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
			// No state migrations or store changes needed for this upgrade.
			return fromVM, nil
		},
	)
}
