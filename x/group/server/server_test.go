package server_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/suite"

	"github.com/fetchai/fetchd/types/module"
	"github.com/fetchai/fetchd/types/module/server"
	group "github.com/fetchai/fetchd/x/group/module"
	"github.com/fetchai/fetchd/x/group/server/testsuite"
)

func TestServer(t *testing.T) {
	ff := server.NewFixtureFactory(t, 6)
	cdc := ff.Codec()
	// Setting up bank keeper
	banktypes.RegisterInterfaces(cdc.InterfaceRegistry())
	authtypes.RegisterInterfaces(cdc.InterfaceRegistry())

	paramsKey := sdk.NewKVStoreKey(paramstypes.StoreKey)
	authKey := sdk.NewKVStoreKey(authtypes.StoreKey)
	bankKey := sdk.NewKVStoreKey(banktypes.StoreKey)
	mintKey := sdk.NewKVStoreKey(minttypes.StoreKey)
	stakingKey := sdk.NewKVStoreKey(stakingtypes.StoreKey)
	tkey := sdk.NewTransientStoreKey(paramstypes.TStoreKey)
	amino := codec.NewLegacyAmino()

	authSubspace := paramstypes.NewSubspace(cdc, amino, paramsKey, tkey, authtypes.ModuleName)
	bankSubspace := paramstypes.NewSubspace(cdc, amino, paramsKey, tkey, banktypes.ModuleName)
	stakingSubspace := paramstypes.NewSubspace(cdc, amino, paramsKey, tkey, stakingtypes.ModuleName)
	mintSubspace := paramstypes.NewSubspace(cdc, amino, paramsKey, tkey, minttypes.ModuleName)

	maccPerms := map[string][]string{
		authtypes.FeeCollectorName:     nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
	}

	accountKeeper := authkeeper.NewAccountKeeper(
		cdc, authKey, authSubspace, authtypes.ProtoBaseAccount, maccPerms,
	)

	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc, bankKey, accountKeeper, bankSubspace, modAccAddrs,
	)

	stakingKeeper := stakingkeeper.NewKeeper(
		cdc, stakingKey, accountKeeper, bankKeeper, stakingSubspace,
	)

	mintKeeper := mintkeeper.NewKeeper(
		cdc, mintKey, mintSubspace, stakingKeeper, accountKeeper, bankKeeper, authtypes.FeeCollectorName,
	)

	baseApp := ff.BaseApp()
	baseApp.Router().AddRoute(sdk.NewRoute(banktypes.ModuleName, bank.NewHandler(bankKeeper))) // TODO: remove once sdk v0.44 is landed
	baseApp.MsgServiceRouter().SetInterfaceRegistry(cdc.InterfaceRegistry())
	banktypes.RegisterMsgServer(baseApp.MsgServiceRouter(), bankkeeper.NewMsgServerImpl(bankKeeper))
	baseApp.MountStore(tkey, sdk.StoreTypeTransient)
	baseApp.MountStore(paramsKey, sdk.StoreTypeIAVL)
	baseApp.MountStore(authKey, sdk.StoreTypeIAVL)
	baseApp.MountStore(bankKey, sdk.StoreTypeIAVL)
	baseApp.MountStore(stakingKey, sdk.StoreTypeIAVL)
	baseApp.MountStore(mintKey, sdk.StoreTypeIAVL)

	ff.SetModules([]module.Module{
		group.Module{
			AccountKeeper: accountKeeper,
		},
	})

	s := testsuite.NewIntegrationTestSuite(ff, accountKeeper, bankKeeper, mintKeeper)

	suite.Run(t, s)
}
