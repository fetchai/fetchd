package module

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/fetchai/fetchd/x/blsgroup"
	blsgroupcli "github.com/fetchai/fetchd/x/blsgroup/client/cli"
	"github.com/fetchai/fetchd/x/blsgroup/keeper"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

type AppModule struct {
	AppModuleBasic

	keeper   keeper.Keeper
	registry cdctypes.InterfaceRegistry
}

type AppModuleBasic struct {
	cdc codec.Codec
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper, registry cdctypes.InterfaceRegistry) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         keeper,
		registry:       registry,
	}
}

func (am AppModule) RegisterInvariants(sdk.InvariantRegistry) {}

func (am AppModule) Route() sdk.Route {
	return sdk.Route{}
}

func (am AppModule) QuerierRoute() string {
	return ""
}

func (am AppModule) LegacyQuerierHandler(*codec.LegacyAmino) sdk.Querier {
	return nil
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	blsgroup.RegisterMsgServer(cfg.MsgServer(), am.keeper)
}

func (am AppModule) ConsensusVersion() uint64 {
	return 1
}

func (am AppModule) BeginBlock(sdk.Context, abci.RequestBeginBlock) {}

func (am AppModule) EndBlock(sdk.Context, abci.RequestEndBlock) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

func (am AppModule) InitGenesis(sdk.Context, codec.JSONCodec, json.RawMessage) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}

func (am AppModule) ExportGenesis(sdk.Context, codec.JSONCodec) json.RawMessage {
	return nil
}

func (am AppModule) Name() string {
	return blsgroup.ModuleName
}

func (a AppModuleBasic) Name() string {
	return blsgroup.ModuleName
}

func (a AppModuleBasic) RegisterLegacyAminoCodec(*codec.LegacyAmino) {}

func (a AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	blsgroup.RegisterInterfaces(registry)
}

func (a AppModuleBasic) DefaultGenesis(codec.JSONCodec) json.RawMessage {
	return nil
}

func (a AppModuleBasic) ValidateGenesis(codec.JSONCodec, client.TxEncodingConfig, json.RawMessage) error {
	return nil
}

func (a AppModuleBasic) RegisterRESTRoutes(client.Context, *mux.Router) {}

func (a AppModuleBasic) RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux) {}

func (a AppModuleBasic) GetTxCmd() *cobra.Command {
	return blsgroupcli.TxCmd(a.Name())
}

func (a AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil
}
