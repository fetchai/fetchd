package server

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/fetchai/fetchd/x/group"
	"github.com/fetchai/fetchd/x/group/simulation"
)

// WeightedOperations returns all the group module operations with their respective weights.
func (s serverImpl) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {

	interfaceRegistry := types.NewInterfaceRegistry()
	queryClient := group.NewQueryClient(s.key)
	return simulation.WeightedOperations(
		simState.AppParams, simState.Cdc,
		s.accKeeper, s.bankKeeper, queryClient, codec.NewProtoCodec(interfaceRegistry),
	)
}
