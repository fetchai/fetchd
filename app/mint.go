package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

// FixedInflationCalculationFn returns a fixed 3% inflation rate
func FixedInflationCalculationFn(ctx sdk.Context, minter minttypes.Minter, params minttypes.Params, bondedRatio sdk.Dec) sdk.Dec {
	// hardcoded for now, but we could move it to consensus later and user params.InflationMin or params.InflationMax
	// once we migrated the state to undo our changes to the params struct.
	// current:
	//  Params{
	//      InflationRate: 0.03,
	//      BlocksPerYear: xxx
	//  }
	// need to become:
	//  Params{
	//      InflationMin: 0.03,
	//      InflationMax: 0.03,
	//      InflationRateChange: 0,
	//      BlocksPerYear: xxx
	//  }
	return sdk.NewDecWithPrec(3, 2)
}
