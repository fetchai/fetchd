package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/fetchai/fetchd/crypto/keys/bls12381"
)

const (
	// DefaultSigVerifyCostBls12381 defines the default cost of a Bls signature verification
	// NOTE: if for some reasons we want this as part of the state, same as for ED25519 or Secp256k1 keys,
	// we'd need to have a custom module defining this parameter and create our own CustomSignatureVerificationGasConsumer
	// given the complexity of this, we'll keep the hardcoded value until there's some need of having this
	// as a gov-tweakable parameter.
	DefaultSigVerifyCostBls12381 uint64 = 6300
)

// BlsDefaultSigVerificationGasConsumer extends the default ante.DefaultSigVerificationGasConsumer to add support fo bls pubkeys.
func BlsDefaultSigVerificationGasConsumer(meter sdk.GasMeter, sig signing.SignatureV2, params types.Params) error {
	switch sig.PubKey.(type) {
	case *bls12381.PubKey:
		meter.ConsumeGas(DefaultSigVerifyCostBls12381, "ante verify: bls12381")
		return nil
	default:
		return ante.DefaultSigVerificationGasConsumer(meter, sig, params)
	}
}
