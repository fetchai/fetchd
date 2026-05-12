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
	"github.com/CosmWasm/wasmd/app/upgrades"
	"github.com/CosmWasm/wasmd/app/upgrades/noop"
	v060 "github.com/CosmWasm/wasmd/app/upgrades/v060"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	epochstypes "github.com/cosmos/cosmos-sdk/x/epochs/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/strangelove-ventures/tokenfactory/x/tokenfactory/keeper"

	//minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	liquidtypes "github.com/cosmos/gaia/v25/x/liquid/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"
	icatypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v10/modules/core/exported"
	"github.com/fetchai/fetchd/app/ica_migration"
	"github.com/fetchai/fetchd/app/traces"
	tokenfactorytypes "github.com/strangelove-ventures/tokenfactory/x/tokenfactory/types"
)

// ---- Match this to the plan name that is already stored on disk and halted the chain.
const UpgradeNameV053 = "v0.15.0-rc1-gemini"

// List ALL new/renamed/deleted KV stores at this upgrade height.
var v053StoreUpgrades = storetypes.StoreUpgrades{
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

// Keep your existing wasmd upgrades
var Upgrades = []upgrades.Upgrade{v060.Upgrade}

func (app *App) RegisterUpgradeHandlers(cfg module.Configurator) {
	if len(Upgrades) == 0 {
		Upgrades = append(Upgrades, noop.NewUpgrade(app.Version()))
	}

	keepers := upgrades.AppKeepers{
		AccountKeeper:         &app.AccountKeeper,
		ConsensusParamsKeeper: &app.ConsensusParamsKeeper,
		IBCKeeper:             app.IBCKeeper,
		Codec:                 app.appCodec,
		GetStoreKey:           app.GetKey,
	}
	app.GetStoreKeys()

	// 1) Existing wasmd handlers
	for _, upgrade := range Upgrades {
		app.UpgradeKeeper.SetUpgradeHandler(
			upgrade.UpgradeName,
			upgrade.CreateUpgradeHandler(app.mm, cfg, &keepers),
		)
	}

	// 2) The v0.53 handler that runs per-module Migrators
	app.UpgradeKeeper.SetUpgradeHandler(
		UpgradeNameV053,
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

			type ChainConfig struct {
				DenomAdmins map[string]string
				Params      tokenfactorytypes.Params
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
					DenomAdmins: map[string]string{
						// Mainnet bridge contract
						bondDenom: "fetch1qxxlalvsdjd07p07y3rc5fu6ll8k4tmetpha8n",
					},
					Params: defaultParams,
				}
			case "dorado-1":
				chainConfig = ChainConfig{
					DenomAdmins: map[string]string{
						// TODO(pb): Deploy bridge (or dummy) contract on `dorado-1` chain and PROVIDE here the address.
						//           Make sure that following points are ensured:
						//           * contract super-admin is set to Fetch Foundation Multi-Sig account,
						//           * contract bitecode is either *real* bridge contract or it is dummy contract with
						//             *NO* API (= contract is not callable)
						//           * is the contract is real bridge contract, make sure that the Fetch Foundation
						//             Multi-Sig account is the only set in admin role.
						bondDenom: "fetch182q50y030ctp39dkjhv4pn95h9vxg29s67djtr0560fuwprtks0sfrtyz0",
					},
					Params: defaultParams,
				}

				// TODO(pb): Ensure that new bridge contract has set the 'label' to the "token-bridge-contract".
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
			for denom, adminAddr := range chainConfig.DenomAdmins {
				err = udc.CreateDenom(sdkCtx, adminAddr, denom)
				if err != nil {
					return nil, err
				}
			}

			return res, err
		},
	)

	// 3) Load the correct store shape at the upgrade height
	upgradeInfo, err := app.UpgradeKeeper.ReadUpgradeInfoFromDisk()
	if err != nil {
		panic(fmt.Sprintf("failed to read upgrade info from disk %s", err))
	}
	if app.UpgradeKeeper.IsSkipHeight(upgradeInfo.Height) {
		return
	}

	// Try wasmd-defined upgrades first
	for _, u := range Upgrades {
		if upgradeInfo.Name == u.UpgradeName {
			app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &u.StoreUpgrades))
			return
		}
	}

	// Then our v0.53 plan
	if upgradeInfo.Name == UpgradeNameV053 {
		app.SetStoreLoader(upgradetypes.UpgradeStoreLoader(upgradeInfo.Height, &v053StoreUpgrades))
	}
}

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

func parseI64(s string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(s), 10, 64)
}
