package bls12381_test

import (
	"encoding/base64"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto"

	"github.com/fetchai/fetchd/crypto/keys/bls12381"
)

func TestSignAndValidateBls12381(t *testing.T) {
	privKey := bls12381.GenPrivKey()
	pubKey := privKey.PubKey()

	msg := crypto.CRandBytes(1000)
	sig, err := privKey.Sign(msg)
	require.Nil(t, err)

	// Test the signature
	assert.True(t, pubKey.VerifySignature(msg, sig))

}

func TestKeyFromSecret(t *testing.T) {
	insecureSeed := []byte("a random number for testing")
	privKey := bls12381.GenPrivKeyFromSecret(insecureSeed)
	pubKey := privKey.PubKey()

	msg := []byte("hello")
	sig, err := privKey.Sign(msg)
	require.Nil(t, err)
	assert.True(t, pubKey.VerifySignature(msg, sig))
}

func TestPubKeyEquals(t *testing.T) {
	bls12381PubKey := bls12381.GenPrivKey().PubKey().(*bls12381.PubKey)

	testCases := []struct {
		msg      string
		pubKey   cryptotypes.PubKey
		other    cryptotypes.PubKey
		expectEq bool
	}{
		{
			"different bytes",
			bls12381PubKey,
			bls12381.GenPrivKey().PubKey(),
			false,
		},
		{
			"equals",
			bls12381PubKey,
			&bls12381.PubKey{
				Key: bls12381PubKey.Key,
			},
			true,
		},
		{
			"different types",
			bls12381PubKey,
			secp256k1.GenPrivKey().PubKey(),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) {
			eq := tc.pubKey.Equals(tc.other)
			require.Equal(t, eq, tc.expectEq)
		})
	}
}

func TestPrivKeyEquals(t *testing.T) {
	bls12381PrivKey := bls12381.GenPrivKey()

	testCases := []struct {
		msg      string
		privKey  cryptotypes.PrivKey
		other    cryptotypes.PrivKey
		expectEq bool
	}{
		{
			"different bytes",
			bls12381PrivKey,
			bls12381.GenPrivKey(),
			false,
		},
		{
			"equals",
			bls12381PrivKey,
			&bls12381.PrivKey{
				Key: bls12381PrivKey.Key,
			},
			true,
		},
		{
			"different types",
			bls12381PrivKey,
			secp256k1.GenPrivKey(),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.msg, func(t *testing.T) {
			eq := tc.privKey.Equals(tc.other)
			require.Equal(t, eq, tc.expectEq)
		})
	}
}

func TestMarshalAmino(t *testing.T) {
	aminoCdc := codec.NewLegacyAmino()
	privKey := bls12381.GenPrivKey()
	pubKey := privKey.PubKey().(*bls12381.PubKey)

	testCases := []struct {
		desc      string
		msg       codec.AminoMarshaler
		typ       interface{}
		expBinary []byte
		expJSON   string
	}{
		{
			"bls12381 private key",
			privKey,
			&bls12381.PrivKey{},
			append([]byte{32}, privKey.Bytes()...), // Length-prefixed.
			"\"" + base64.StdEncoding.EncodeToString(privKey.Bytes()) + "\"",
		},
		{
			"bls12381 public key",
			pubKey,
			&bls12381.PubKey{},
			append([]byte{96}, pubKey.Bytes()...), // Length-prefixed.
			"\"" + base64.StdEncoding.EncodeToString(pubKey.Bytes()) + "\"",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			// Do a round trip of encoding/decoding binary.
			bz, err := aminoCdc.Marshal(tc.msg)
			require.NoError(t, err)
			require.Equal(t, tc.expBinary, bz)

			err = aminoCdc.Unmarshal(bz, tc.typ)
			require.NoError(t, err)

			require.Equal(t, tc.msg, tc.typ)

			// Do a round trip of encoding/decoding JSON.
			bz, err = aminoCdc.MarshalJSON(tc.msg)
			require.NoError(t, err)
			require.Equal(t, tc.expJSON, string(bz))

			err = aminoCdc.UnmarshalJSON(bz, tc.typ)
			require.NoError(t, err)

			require.Equal(t, tc.msg, tc.typ)
		})
	}
}

func TestMarshalJSON(t *testing.T) {
	require := require.New(t)
	privKey := bls12381.GenPrivKey()
	pk := privKey.PubKey()

	registry := types.NewInterfaceRegistry()
	bls12381.RegisterInterfaces(registry)
	cryptocodec.RegisterInterfaces(registry)

	cdc := codec.NewProtoCodec(registry)

	bz, err := cdc.MarshalInterfaceJSON(pk)
	require.NoError(err)

	var pk2 cryptotypes.PubKey
	err = cdc.UnmarshalInterfaceJSON(bz, &pk2)
	require.NoError(err)
	require.True(pk2.Equals(pk))
}
