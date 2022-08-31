package bbsplus

import (
	crand "crypto/rand"
	"testing"

	"github.com/fetchai/fetchd/x/verifiable-credential/crypto"

	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/signatures/common"
	"github.com/stretchr/testify/require"
)

func TestBbsPlus(t *testing.T) {
	sk, pp, err := NewBbsPlusSchema(3)
	require.NoError(t, err)

	msg1 := Curve.Scalar.Hash([]byte("3"))
	msg2 := Curve.Scalar.Hash([]byte("4"))
	msg3 := Curve.Scalar.Hash([]byte("5"))
	msgs := []curves.Scalar{msg1, msg2, msg3}

	sig, err := sk.Sign(pp, msgs)
	require.NoError(t, err)

	var nonce [32]byte
	cnt, err := crand.Read(nonce[:])
	require.NoError(t, err)
	require.Equal(t, 32, cnt)

	proofMsgs := []common.ProofMessage{
		&common.RevealedMessage{
			Message: msgs[0],
		},
		&common.ProofSpecificMessage{
			Message: msgs[1],
		},
		&common.ProofSpecificMessage{
			Message: msgs[2],
		},
	}
	pok, okm, err := CreateProofPre(pp, sig, nonce[:], proofMsgs)
	require.NoError(t, err)

	proof, err := CreateProofPost(pok, okm)
	require.NoError(t, err)

	revealedMsgs := map[int]curves.Scalar{0: msgs[0]}
	okm2, err := VerifyProof(pp, revealedMsgs, nonce[:], okm, proof)
	require.NoError(t, err)

	err = crypto.IsChallengeEqual(okm, okm2)
	require.NoError(t, err)
}
