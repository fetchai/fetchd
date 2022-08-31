package bbsplus

import (
	crand "crypto/rand"
	"errors"
	"fmt"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/signatures/bbs"
	"github.com/coinbase/kryptology/pkg/signatures/common"
	"github.com/fetchai/fetchd/x/verifiable-credential/crypto"
	"github.com/gtank/merlin"
)

var Curve = curves.BLS12381(&curves.PointBls12381G2{})

const TransLabel = "PokSignatureProof"

func NewBbsPlusSchema(msgLen int) (*PrivateKey, *PublicParameters, error) {
	if msgLen < 0 {
		return nil, nil, errors.New("message length should be non-negative")
	}
	pk, sk, err := bbs.NewKeys(Curve)
	if err != nil {
		return nil, nil, err
	}
	skBytes, err := sk.MarshalBinary()
	if err != nil {
		return nil, nil, err
	}
	pkBytes, err := pk.MarshalBinary()
	if err != nil {
		return nil, nil, err
	}
	privateKey := PrivateKey{Key: skBytes}
	publicParams := PublicParameters{
		int32(msgLen),
		pkBytes,
	}
	return &privateKey, &publicParams, nil
}

func (bsk *PrivateKey) Sign(pp *PublicParameters, msgs []curves.Scalar) (signature []byte, err error) {
	sk := new(bbs.SecretKey).Init(Curve)
	err = sk.UnmarshalBinary(bsk.Key)
	if err != nil {
		return nil, err
	}

	pk := new(bbs.PublicKey).Init(Curve)
	err = pk.UnmarshalBinary(pp.PublicKey)
	if err != nil {
		return nil, err
	}
	generators, err := new(bbs.MessageGenerators).Init(pk, int(pp.MsgLength))
	if err != nil {
		return nil, err
	}

	sig, err := sk.Sign(generators, msgs)
	if err != nil {
		return nil, err
	}

	err = pk.Verify(sig, generators, msgs)
	if err != nil {
		return nil, err
	}

	signature, err = sig.MarshalBinary()
	if err != nil {
		return nil, err
	}

	return signature, nil
}

func CreateProofPre(pp *PublicParameters, signature []byte, nonce []byte, proofMsgs []common.ProofMessage) (pok *bbs.PokSignature, bbsOkm []byte, err error) {
	sig := new(bbs.Signature).Init(Curve)
	err = sig.UnmarshalBinary(signature)
	if err != nil {
		return nil, nil, err
	}

	pk := new(bbs.PublicKey).Init(Curve)
	err = pk.UnmarshalBinary(pp.PublicKey)
	if err != nil {
		return nil, nil, err
	}
	generators, err := new(bbs.MessageGenerators).Init(pk, int(pp.MsgLength))
	if err != nil {
		return nil, nil, err
	}

	pok, err = bbs.NewPokSignature(sig, generators, proofMsgs, crand.Reader)
	if err != nil {
		return nil, nil, err
	}

	transcript := merlin.NewTranscript(TransLabel)
	pok.GetChallengeContribution(transcript)
	transcript.AppendMessage([]byte("nonce"), nonce)
	bbsOkm = transcript.ExtractBytes([]byte("signature proof of knowledge"), 64)

	return pok, bbsOkm, err
}

func GetChallenge(okm []byte) curves.Scalar {
	prefix := []byte(crypto.ChallengePrefix)
	c := append(prefix, okm...)
	return Curve.Scalar.Hash(c)
}

func CreateProofPost(pok *bbs.PokSignature, challengeOkm []byte) (proof []byte, err error) {
	challenge := GetChallenge(challengeOkm)
	pokSig, err := pok.GenerateProof(challenge)
	if err != nil {
		return nil, err
	}
	proof, err = pokSig.MarshalBinary()

	return proof, err
}

func VerifyProof(pp *PublicParameters, revealedMsgs map[int]curves.Scalar, nonce []byte, challengeOkm []byte, proof []byte) (bbsOkm []byte, err error) {
	transcript := merlin.NewTranscript(TransLabel)

	pokSig := new(bbs.PokSignatureProof).Init(Curve)
	err = pokSig.UnmarshalBinary(proof)
	if err != nil {
		return nil, err
	}

	pk := new(bbs.PublicKey).Init(Curve)
	err = pk.UnmarshalBinary(pp.PublicKey)
	if err != nil {
		return nil, err
	}
	generators, err := new(bbs.MessageGenerators).Init(pk, int(pp.MsgLength))
	if err != nil {
		return nil, err
	}

	challenge := GetChallenge(challengeOkm)
	pokSig.GetChallengeContribution(generators, revealedMsgs, challenge, transcript)
	if !pokSig.VerifySigPok(pk) {
		return nil, fmt.Errorf("BBS+ proof verification (sig) failed")
	}
	transcript.AppendMessage([]byte("nonce"), nonce)
	bbsOkm = transcript.ExtractBytes([]byte("signature proof of knowledge"), 64)

	return bbsOkm, err
}

func GetPublicBlindingForMessage(proof []byte, index int) ([]byte, error) {
	pokSig := new(bbs.PokSignatureProof).Init(Curve)
	err := pokSig.UnmarshalBinary(proof)
	if err != nil {
		return nil, err
	}

	sMess, err := pokSig.GetPublicBlindingForMessage(index)
	if err != nil {
		return nil, err
	}
	return sMess.Bytes(), nil
}
