package app

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	epochstypes "github.com/cosmos/cosmos-sdk/x/epochs/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	icatypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"

	storetypes "cosmossdk.io/store/types"
	circuittypes "cosmossdk.io/x/circuit/types"
	"cosmossdk.io/x/nft"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/CosmWasm/wasmd/app/upgrades"
	"github.com/CosmWasm/wasmd/app/upgrades/noop"
	v060 "github.com/CosmWasm/wasmd/app/upgrades/v060"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	protocolpooltypes "github.com/cosmos/cosmos-sdk/x/protocolpool/types"
	icacontrollertypes "github.com/cosmos/ibc-go/v10/modules/apps/27-interchain-accounts/controller/types"
)

// ---- Match this to the plan name that is already stored on disk and halted the chain.
const UpgradeNameV053 = "v0.14.0" // <-- CHANGE if your plan name is different.

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

			// TODO: Write custom migration for ICA
			err := cfg.RegisterMigration(icatypes.ModuleName, 1, func(ctx sdk.Context) error {
				// Logic here
				return nil
			})
			if err != nil {
				return nil, err
			}

			// TODO: Write custom migration for transfer
			err = cfg.RegisterMigration(ibctransfertypes.ModuleName, 1, func(ctx sdk.Context) error {
				// Logic here
				return nil
			})
			if err != nil {
				return nil, err
			}

			// Pre-seed legacy x/params for mint
			sdkCtx := sdk.UnwrapSDKContext(ctx)
			if ss, ok := app.ParamsKeeper.GetSubspace(minttypes.ModuleName); ok {
				if !ss.Has(sdkCtx, minttypes.KeyInflationRateChange) {
					p := minttypes.DefaultParams()
					p.MintDenom = "afet"
					// TODO: if your chain had custom values, set them here:
					ss.SetParamSet(sdkCtx, &p)
				}
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
