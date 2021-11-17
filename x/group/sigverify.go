package group

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/bls12381"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func VerifyAggregateSignature(msgs [][]byte, msgCheck bool, sig []byte, pkss [][]cryptotypes.PubKey) error {
	pkssBls := make([][]*bls12381.PubKey, len(pkss))
	for i, pks := range pkss {
		for _, pk := range pks {
			pkBls, ok := pk.(*bls12381.PubKey)
			if !ok {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidPubKey, "only support bls public key")
			}
			pkssBls[i] = append(pkssBls[i], pkBls)
		}
	}

	return bls12381.VerifyAggregateSignature(msgs, msgCheck, sig, pkssBls)
}
