package blsgroup

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/fetchai/fetchd/crypto/keys/bls12381"
)

func VerifyAggregateSignature(msgs [][]byte, msgCheck bool, sig []byte, pks []cryptotypes.PubKey) error {
	pkssBls := make([][]*bls12381.PubKey, len(pks))
	for i, pk := range pks {
		pkBls, ok := pk.(*bls12381.PubKey)
		if !ok {
			return sdkerrors.Wrap(sdkerrors.ErrInvalidPubKey, "only support bls public key")
		}
		pkssBls[i] = append(pkssBls[i], pkBls)
	}

	return bls12381.VerifyAggregateSignature(msgs, msgCheck, sig, pkssBls)
}
