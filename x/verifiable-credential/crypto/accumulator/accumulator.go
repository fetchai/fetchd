package accumulator

import (
	crand "crypto/rand"
	"errors"
	"fmt"

	"github.com/fetchai/fetchd/x/verifiable-credential/crypto"

	accumcrypto "github.com/coinbase/kryptology/pkg/accumulator"
	"github.com/coinbase/kryptology/pkg/core/curves"
)

var Curve = curves.BLS12381(&curves.PointBls12381G1{})

func NewAccumulatorSchema() (*PrivateKey, *PublicParameters, error) {
	var ikm [32]byte
	cnt, err := crand.Read(ikm[:])
	if err != nil {
		return nil, nil, err
	}
	if cnt != 32 {
		return nil, nil, fmt.Errorf("unable to read sufficient random data")
	}

	sk, err := new(accumcrypto.SecretKey).New(Curve, ikm[:])
	if err != nil {
		return nil, nil, err
	}

	pk, err := sk.GetPublicKey(Curve)
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

	return &PrivateKey{Value: skBytes}, &PublicParameters{PublicKey: pkBytes}, nil
}

func (pp *PublicParameters) InitAccumulator(ask *PrivateKey, members accumcrypto.ElementSet) (*PublicParameters, error) {
	sk := new(accumcrypto.SecretKey)
	err := sk.UnmarshalBinary(ask.Value)
	if err != nil {
		return nil, err
	}
	accum, err := new(accumcrypto.Accumulator).WithElements(Curve, sk, members.Elements)
	if err != nil {
		return nil, err
	}
	acc, err := accum.MarshalBinary()
	if err != nil {
		return nil, err
	}
	pp.States = append(pp.States, &State{AccValue: acc, Update: nil})

	return pp, err
}

func (ask *PrivateKey) InitMemberWitness(pp *PublicParameters, member accumcrypto.Element) (wit []byte, err error) {
	sk := new(accumcrypto.SecretKey)
	err = sk.UnmarshalBinary(ask.Value)
	if err != nil {
		return nil, err
	}

	accum := new(accumcrypto.Accumulator)
	num := len(pp.States)
	if num < 1 {
		fmt.Errorf("need to initiate accumulator first")
	}
	err = accum.UnmarshalBinary(pp.States[num-1].AccValue)
	if err != nil {
		return nil, err
	}
	witness, err := new(accumcrypto.MembershipWitness).New(member, accum, sk)
	if err != nil {
		return nil, err
	}
	return witness.MarshalBinary()
}

func CreateMembershipProofPre(pp *PublicParameters, wit []byte, eb *accumcrypto.ExternalBlinding) (mpc *accumcrypto.MembershipProofCommitting, accumOkm []byte, proofEntropy []byte, err error) {
	pk := new(accumcrypto.PublicKey)
	err = pk.UnmarshalBinary(pp.PublicKey)
	if err != nil {
		return nil, nil, nil, err
	}

	var entropy [32]byte
	cnt, err := crand.Read(entropy[:])
	if err != nil {
		return nil, nil, nil, err
	}
	if cnt != 32 {
		return nil, nil, nil, errors.New("unable to read sufficient random data")
	}

	params, err := new(accumcrypto.ProofParams).New(Curve, pk, entropy[:])

	witness := new(accumcrypto.MembershipWitness)
	err = witness.UnmarshalBinary(wit)
	if err != nil {
		return nil, nil, nil, err
	}

	accum := new(accumcrypto.Accumulator)
	num := len(pp.States)
	err = accum.UnmarshalBinary(pp.States[num-1].AccValue)
	if err != nil {
		return nil, nil, nil, err
	}

	mpc, err = new(accumcrypto.MembershipProofCommitting).New(witness, accum, params, pk, eb)
	if err != nil {
		return nil, nil, nil, err
	}
	accumOkm = mpc.GetChallengeBytes()

	return mpc, accumOkm, entropy[:], err
}

func GetChallenge(okm []byte) curves.Scalar {
	prefix := []byte(crypto.ChallengePrefix)
	c := append(prefix, okm...)
	return Curve.Scalar.Hash(c)
}

func CreateMembershipProofPost(mpc *accumcrypto.MembershipProofCommitting, challengeOkm []byte) (proof []byte, err error) {
	// generate the final membership proof
	challenge := GetChallenge(challengeOkm)
	memProof := mpc.GenProof(challenge)
	return memProof.MarshalBinary()
}

func VerifyMembershipProof(pp *PublicParameters, proofEntropy []byte, challengeOkm []byte, proof []byte) (accOkm []byte, err error) {
	pk := new(accumcrypto.PublicKey)
	err = pk.UnmarshalBinary(pp.PublicKey)
	if err != nil {
		return nil, err
	}

	accum := new(accumcrypto.Accumulator)
	// obtain latest state
	num := len(pp.States)
	err = accum.UnmarshalBinary(pp.States[num-1].AccValue)
	if err != nil {
		return nil, err
	}

	// recreate generators to make sure they are not co-related
	params, err := new(accumcrypto.ProofParams).New(Curve, pk, proofEntropy[:])
	if err != nil {
		return nil, err
	}

	challenge := GetChallenge(challengeOkm)

	memProof := new(accumcrypto.MembershipProof)
	err = memProof.UnmarshalBinary(proof)
	if err != nil {
		return nil, err
	}
	finalProof, err := memProof.Finalize(accum, params, pk, challenge)
	if err != nil {
		return nil, err
	}
	accOkm = finalProof.GetChallengeBytes(Curve)

	return accOkm, err
}

func (pp *PublicParameters) UpdateAccumulatorState(ask *PrivateKey, adds accumcrypto.ElementSet, dels accumcrypto.ElementSet) (*PublicParameters, *State, error) {
	if adds.Elements == nil && dels.Elements == nil {
		return nil, nil, fmt.Errorf("addition and deletion are both empty: nothing to update")
	}

	sk := new(accumcrypto.SecretKey)
	err := sk.UnmarshalBinary(ask.Value)
	if err != nil {
		return nil, nil, err
	}

	accum := new(accumcrypto.Accumulator)
	num := len(pp.States)
	if num < 1 {
		fmt.Errorf("need to initiate accumulator first")
	}
	accum.UnmarshalBinary(pp.States[num-1].AccValue)
	newAccum, coeffs, err := accum.Update(sk, adds.Elements, dels.Elements)
	if err != nil {
		return nil, nil, err
	}
	newAcc, err := newAccum.MarshalBinary()

	additions, err := adds.MarshalBinary()
	if err != nil {
		return nil, nil, err
	}

	deletions, err := dels.MarshalBinary()
	if err != nil {
		return nil, nil, err
	}

	coffSet := accumcrypto.CoefficientSet{coeffs}
	coefficients, err := coffSet.MarshalBinary()

	batchUpdate := BatchUpdate{additions, deletions, coefficients}
	newState := &State{newAcc, &batchUpdate}
	pp.States = append(pp.States, newState)

	return pp, newState, nil
}

func UpdateWitness(oldPp, pp *PublicParameters, wit []byte) (newWit []byte, err error) {
	pk := new(accumcrypto.PublicKey)
	err = pk.UnmarshalBinary(pp.PublicKey)
	if err != nil {
		return nil, err
	}

	n1 := len(oldPp.States)
	n2 := len(pp.States)
	if n1 >= n2 {
		return nil, fmt.Errorf("no update for accumulator states")
	}

	witness := new(accumcrypto.MembershipWitness)
	err = witness.UnmarshalBinary(wit)
	if err != nil {
		return nil, err
	}

	accum := new(accumcrypto.Accumulator)
	accum.UnmarshalBinary(pp.States[n2-1].AccValue)

	var A, D [][]accumcrypto.Element
	var C [][]accumcrypto.Coefficient
	for i := n1; i < n2; i++ {
		adds := new(accumcrypto.ElementSet)
		err = adds.UnmarshalBinary(pp.States[i].Update.Additions)
		if err != nil {
			return nil, err
		}

		dels := new(accumcrypto.ElementSet)
		err = dels.UnmarshalBinary(pp.States[i].Update.Deletions)
		if err != nil {
			return nil, err
		}

		coeffs := new(accumcrypto.CoefficientSet)
		err = coeffs.UnmarshalBinary(pp.States[i].Update.Coefficients)
		if err != nil {
			return nil, err
		}

		A = append(A, adds.Elements)
		D = append(D, dels.Elements)
		C = append(C, coeffs.Coefficients)
	}

	newWitness, err := witness.MultiBatchUpdate(A, D, C)
	if err != nil {
		return nil, err
	}

	err = newWitness.Verify(pk, accum)
	if err != nil {
		return nil, err
	}

	return newWitness.MarshalBinary()
}

func GetPublicBlinding(proof []byte) ([]byte, error) {
	memProof := new(accumcrypto.MembershipProof)
	err := memProof.UnmarshalBinary(proof)
	if err != nil {
		return nil, err
	}

	s := memProof.GetPublicBlinding()
	return s.Bytes(), nil
}
