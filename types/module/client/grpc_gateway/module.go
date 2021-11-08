package grpc_gateway

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/fetchai/fetchd/types/module"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

// Module is an interface that modules should implement to register grpc-gateway routes.
type Module interface {
	module.Module

	RegisterGRPCGatewayRoutes(client.Context, *runtime.ServeMux)
}
