package anonymouscredential

import (
	"bytes"
	"errors"
	"fmt"

	accumcrypto "github.com/coinbase/kryptology/pkg/accumulator"
	"github.com/fetchai/fetchd/x/verifiable-credential/crypto"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/fetchai/fetchd/x/verifiable-credential/crypto/accumulator"
	"github.com/fetchai/fetchd/x/verifiable-credential/crypto/bbsplus"
)

func NewAnonymousCredentialSchema(msgLen int) (*PrivateKey, *PublicParameters, error) {
	if msgLen < 0 {
		return nil, nil, errors.New("message length should be non-negative")
	}

	// the first message is reserved for linking identifier between bbs+ and accumulator
	bbsPrivateKey, bbsPubParams, err := bbsplus.NewBbsPlusSchema(msgLen)
	if err != nil {
		return nil, nil, err
	}

	accPrivateKey, accPubParams, err := accumulator.NewAccumulatorSchema()
	if err != nil {
		return nil, nil, err
	}

	accPubParams, err = accPubParams.InitAccumulator(accPrivateKey, accumcrypto.ElementSet{})
	if err != nil {
		return nil, nil, err
	}

	privateKey := PrivateKey{bbsPrivateKey, accPrivateKey}
	pubParams := PublicParameters{bbsPubParams, accPubParams}
	return &privateKey, &pubParams, nil
}

func VerifyProof(pp *PublicParameters, revealedMsgs map[int]curves.Scalar, proof *AnonymousCredentialProof) (okm []byte, err error) {
	// verify bbs+ proof
	bbsOkm, err := bbsplus.VerifyProof(pp.BbsPlusPublicParams, revealedMsgs, proof.Nonce, proof.Challenge, proof.BbsPlusProof)
	if err != nil {
		return nil, err
	}
	// verify membership proof
	accOkm, err := accumulator.VerifyMembershipProof(pp.AccumulatorPublicParams, proof.AccumulatorEntropy, proof.Challenge, proof.AccumulatorProof)
	if err != nil {
		return nil, err
	}
	okm = crypto.CombineChanllengeOkm(bbsOkm, accOkm)

	// verify linked external blinding
	eb, err := bbsplus.GetPublicBlindingForMessage(proof.BbsPlusProof, 0)
	if err != nil {
		return nil, err
	}
	eb2, err := accumulator.GetPublicBlinding(proof.AccumulatorProof)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(eb, eb2) {
		return nil, fmt.Errorf("linked public blindings are not equal")
	}

	return okm, nil
}
