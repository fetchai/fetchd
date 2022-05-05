package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/fetchai/fetchd/crypto/keys/bls12381"
)

// BlsDefaultSigVerificationGasConsumer extends the default ante.DefaultSigVerificationGasConsumer to add support fo bls pubkeys.
func BlsDefaultSigVerificationGasConsumer(meter sdk.GasMeter, sig signing.SignatureV2, params types.Params) error {
	switch sig.PubKey.(type) {
	case *bls12381.PubKey:
		meter.ConsumeGas(bls12381.DefaultSigVerifyCostBls12381, "ante verify: bls12381")
		return nil
	default:
		return ante.DefaultSigVerificationGasConsumer(meter, sig, params)
	}
}
