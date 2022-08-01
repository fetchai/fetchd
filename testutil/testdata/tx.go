package testdata

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fetchai/fetchd/crypto/keys/bls12381"
)

// KeyTestPubAddr generates a new secp256k1 keypair.
func KeyTestPubAddrBls12381() (cryptotypes.PrivKey, cryptotypes.PubKey, sdk.AccAddress) {
	key := bls12381.GenPrivKey()
	pub := key.PubKey()
	addr := sdk.AccAddress(pub.Address())
	return key, pub, addr
}
