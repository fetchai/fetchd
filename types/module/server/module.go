package server

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdkmodule "github.com/cosmos/cosmos-sdk/types/module"

	"github.com/fetchai/fetchd/types/module"
)

// Module is the module type that all server modules must satisfy
type Module interface {
	module.TypeModule

	RegisterServices(Configurator)
}

type Configurator interface {
	sdkmodule.Configurator

	ModuleKey() RootModuleKey
	Marshaler() codec.Marshaler
	RequireServer(interface{})
	RegisterInvariantsHandler(registry RegisterInvariantsHandler)
	RegisterGenesisHandlers(module.InitGenesisHandler, module.ExportGenesisHandler)
	RegisterWeightedOperationsHandler(WeightedOperationsHandler)
}
