package server

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"math"
)

func addUint64Overflow(a, b uint64) (uint64, bool) {
	if math.MaxUint64-a < b {
		return 0, true
	}

	return a + b, false
}

func mulUint64Overflow(a, b uint64) (uint64, bool) {
	if math.MaxUint64/a < b {
		return 0, true
	}
	return a*b, false
}

func subUint64Overflow(a, b uint64) (uint64, bool) {
	if a < b {
		return 0, true
	}

	return a - b, false
}

// DefaultSigVerificationGasConsumer computes and consumes gas cost using the formula:
// gas := base + pairingCost * (numMsg - 1) + additionCost * (numPk - numMsg)
func DefaultAggSigVerifyGasConsumer(meter sdk.GasMeter, numPk uint64, numMsg uint64, params authtypes.Params) error {
	base := params.SigVerifyCostBls12381
	pairingCost := base*10/33
	additionCost := base/1215

	sub, overflow := subUint64Overflow(numMsg, 1)
	if overflow {
		return fmt.Errorf("subtraction between uint64 overflow")
	}

	mul, overflow := mulUint64Overflow(pairingCost, sub)
	if overflow {
		return fmt.Errorf("multiplication between uint64 overflow")
	}

	sub2, overflow := subUint64Overflow(numPk, numMsg)
	if overflow {
		return fmt.Errorf("subtraction between uint64 overflow")
	}

	mul2, overflow := mulUint64Overflow(additionCost, sub2)
	if overflow {
		return fmt.Errorf("multiplication between uint64 overflow")
	}

	sum, overflow := addUint64Overflow(base, mul)
	if overflow {
		return fmt.Errorf("addition between uint64 overflow")
	}

	gas, overflow := addUint64Overflow(sum, mul2)
	if overflow {
		return fmt.Errorf("addition between uint64 overflow")
	}

	meter.ConsumeGas(gas, "verify aggregated signature")

	return nil
}