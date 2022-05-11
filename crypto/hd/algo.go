package hd

import (
	cosmoshd "github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/go-bip39"
	"github.com/fetchai/fetchd/crypto/keys/bls12381"
)

const (
	// Bls12381Type represents the Bls12381Type signature system.
	Bls12381Type = cosmoshd.PubKeyType("bls12381")
)

var (
	Bls12381 = bls12381Algo{}
)

type bls12381Algo struct {
}

func (s bls12381Algo) Name() cosmoshd.PubKeyType {
	return Bls12381Type
}

// todo: replace bitcoin private key generation
// Derive derives and returns the bls12381 private key for the given seed and HD path.
func (s bls12381Algo) Derive() cosmoshd.DeriveFn {
	return func(mnemonic string, bip39Passphrase, hdPath string) ([]byte, error) {
		seed, err := bip39.NewSeedWithErrorChecking(mnemonic, bip39Passphrase)
		if err != nil {
			return nil, err
		}

		masterPriv, ch := cosmoshd.ComputeMastersFromSeed(seed)
		if len(hdPath) == 0 {
			return masterPriv[:], nil
		}
		derivedKey, err := cosmoshd.DerivePrivateKeyForPath(masterPriv, ch, hdPath)

		return derivedKey, err
	}
}

// Generate generates a bls12381 private key from the given bytes.
func (s bls12381Algo) Generate() cosmoshd.GenerateFn {
	return func(bz []byte) types.PrivKey {
		var bzArr = make([]byte, bls12381.SeedSize)
		copy(bzArr, bz)
		sk := bls12381.GenPrivKeyFromSecret(bzArr)

		return sk
	}
}
