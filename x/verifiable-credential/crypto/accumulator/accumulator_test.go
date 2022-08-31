package accumulator

import (
	"testing"

	"github.com/fetchai/fetchd/x/verifiable-credential/crypto"

	accumcrypto "github.com/coinbase/kryptology/pkg/accumulator"
	"github.com/stretchr/testify/require"
)

func TestAccumulatorGeneration(t *testing.T) {
	sk, pp, err := NewAccumulatorSchema()
	require.NoError(t, err)

	element1 := Curve.Scalar.Hash([]byte("3"))
	element2 := Curve.Scalar.Hash([]byte("4"))
	element3 := Curve.Scalar.Hash([]byte("5"))
	element4 := Curve.Scalar.Hash([]byte("6"))
	element5 := Curve.Scalar.Hash([]byte("7"))
	members := accumcrypto.ElementSet{[]accumcrypto.Element{element1, element2, element3, element4, element5}}

	pp, err = pp.InitAccumulator(sk, members)
	require.NoError(t, err)

	witBytes, err := sk.InitMemberWitness(pp, members.Elements[2])
	require.NoError(t, err)

	eb, err := new(accumcrypto.ExternalBlinding).New(Curve)
	require.NoError(t, err)
	mpc, accumOkm, proofEntropy, err := CreateMembershipProofPre(pp, witBytes, eb)
	require.NoError(t, err)

	// when using accumulator independently, there is no need to combine okm with external okm
	proofBytes, err := CreateMembershipProofPost(mpc, accumOkm)
	require.NoError(t, err)

	okm, err := VerifyMembershipProof(pp, proofEntropy, accumOkm, proofBytes)
	require.NoError(t, err)

	err = crypto.IsChallengeEqual(accumOkm, okm)
	require.NoError(t, err)
}

func TestAccumulatorUpdate(t *testing.T) {
	sk, pp, err := NewAccumulatorSchema()
	require.NoError(t, err)

	element1 := Curve.Scalar.Hash([]byte("1"))
	element2 := Curve.Scalar.Hash([]byte("2"))
	element3 := Curve.Scalar.Hash([]byte("3"))
	element4 := Curve.Scalar.Hash([]byte("4"))
	element5 := Curve.Scalar.Hash([]byte("5"))
	members := accumcrypto.ElementSet{[]accumcrypto.Element{element1, element2, element3, element4, element5}}

	pp, err = pp.InitAccumulator(sk, members)
	require.NoError(t, err)

	element6 := Curve.Scalar.Hash([]byte("6"))
	element7 := Curve.Scalar.Hash([]byte("7"))
	element8 := Curve.Scalar.Hash([]byte("8"))

	wit, err := sk.InitMemberWitness(pp, members.Elements[2])
	require.NoError(t, err)

	oldPp := *pp

	// update members and witness
	adds := accumcrypto.ElementSet{Elements: []accumcrypto.Element{element6, element7, element8}}
	dels := accumcrypto.ElementSet{Elements: []accumcrypto.Element{element4, element5}}
	pp, _, err = pp.UpdateAccumulatorState(sk, adds, dels)
	require.NoError(t, err)

	newWit, err := UpdateWitness(&oldPp, pp, wit)
	require.NoError(t, err)

	eb, err := new(accumcrypto.ExternalBlinding).New(Curve)
	require.NoError(t, err)
	// create new membership proof
	mpc, accumOkm, proofEntropy, err := CreateMembershipProofPre(pp, newWit, eb)
	require.NoError(t, err)
	// when using accumulator independently, there is no need to combine okm with external okm
	proofBytes, err := CreateMembershipProofPost(mpc, accumOkm)
	require.NoError(t, err)

	okm, err := VerifyMembershipProof(pp, proofEntropy, accumOkm, proofBytes)
	require.NoError(t, err)

	err = crypto.IsChallengeEqual(accumOkm, okm)
	require.NoError(t, err)
}
