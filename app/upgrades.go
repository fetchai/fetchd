package app

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

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
	//minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
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

			// If you must pin any module "from" versions, adjust fromVM here.
			return app.mm.RunMigrations(ctx, cfg, fromVM)
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
