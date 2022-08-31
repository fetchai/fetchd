package anonymouscredential

import (
	crand "crypto/rand"
	"testing"

	"github.com/fetchai/fetchd/x/verifiable-credential/crypto"

	accumcrypto "github.com/coinbase/kryptology/pkg/accumulator"
	"github.com/coinbase/kryptology/pkg/core/curves"
	"github.com/coinbase/kryptology/pkg/signatures/common"
	"github.com/fetchai/fetchd/x/verifiable-credential/crypto/accumulator"
	"github.com/fetchai/fetchd/x/verifiable-credential/crypto/bbsplus"
	"github.com/stretchr/testify/require"
)

func TestAnonymousCredential(t *testing.T) {
	// set up scheme parameters
	sk, pp, err := NewAnonymousCredentialSchema(4)
	require.NoError(t, err)

	// create bbs+ credential for alice
	aliceMemberIdSeed := []byte("a member id for alice")
	msg10 := bbsplus.Curve.Scalar.Hash(aliceMemberIdSeed)
	msg11 := bbsplus.Curve.Scalar.Hash([]byte("alice"))
	msg12 := bbsplus.Curve.Scalar.New(25)
	msg13 := bbsplus.Curve.Scalar.Hash([]byte("london"))
	msgs1 := []curves.Scalar{msg10, msg11, msg12, msg13}
	aliceBbsSig, err := sk.BbsPlusKey.Sign(pp.BbsPlusPublicParams, msgs1)
	require.NoError(t, err)

	// create bbs+ credential for bob
	bobMemberIdseed := []byte("a member id for bob")
	msg20 := bbsplus.Curve.Scalar.Hash(bobMemberIdseed)
	msg21 := bbsplus.Curve.Scalar.Hash([]byte("bob"))
	msg22 := bbsplus.Curve.Scalar.New(30)
	msg23 := bbsplus.Curve.Scalar.Hash([]byte("cambridge"))
	msgs2 := []curves.Scalar{msg20, msg21, msg22, msg23}
	bobBbsSig, err := sk.BbsPlusKey.Sign(pp.BbsPlusPublicParams, msgs2)
	_ = bobBbsSig // no more error
	require.NoError(t, err)

	// initialise accumulator with alice and bob
	mem1 := accumulator.Curve.Scalar.Hash(aliceMemberIdSeed)
	mem2 := accumulator.Curve.Scalar.Hash(bobMemberIdseed)
	adds := accumcrypto.ElementSet{Elements: []accumcrypto.Element{mem1, mem2}}
	dels := accumcrypto.ElementSet{}
	pp.AccumulatorPublicParams, _, err = pp.AccumulatorPublicParams.UpdateAccumulatorState(sk.AccumulatorKey, adds, dels)
	require.NoError(t, err)

	// create witness for each member
	alice_wit, err := sk.AccumulatorKey.InitMemberWitness(pp.AccumulatorPublicParams, mem1)
	require.NoError(t, err)
	bob_wit, err := sk.AccumulatorKey.InitMemberWitness(pp.AccumulatorPublicParams, mem2)
	_ = bob_wit // no more error
	require.NoError(t, err)

	// anonymous crendential for alice is (alice_bbs_sig, alice_wit), (bob_bbs_sig, bob_wit)

	// alice creates a zkp using her credential (alice_bbs_sig, alice_wit)
	// nonce should be from the verifier or a certain event tag
	var nonce [32]byte
	cnt, err := crand.Read(nonce[:])
	require.NoError(t, err)
	require.Equal(t, cnt, 32)

	// create external blinding for membership id
	eb, err := new(accumcrypto.ExternalBlinding).New(accumulator.Curve)
	require.NoError(t, err)

	// bbs+ proof messages
	proofMsgs := []common.ProofMessage{
		&common.SharedBlindingMessage{
			Message:  msgs1[0],
			Blinding: eb.GetBlinding(),
		},
		&common.ProofSpecificMessage{
			Message: msgs1[1],
		},
		&common.ProofSpecificMessage{
			Message: msgs1[2],
		},
		&common.ProofSpecificMessage{
			Message: msgs1[3],
		},
	}
	// create bbs+ proof
	pok, bbsOkm, err := bbsplus.CreateProofPre(pp.BbsPlusPublicParams, aliceBbsSig, nonce[:], proofMsgs)
	require.NoError(t, err)
	// create membership proof
	mpc, accumOkm, memProofEntropy, err := accumulator.CreateMembershipProofPre(pp.AccumulatorPublicParams, alice_wit, eb)
	require.NoError(t, err)
	// merge okm to create challenge
	challengeOkm := crypto.CombineChanllengeOkm(bbsOkm, accumOkm)
	// complete bbs+ proof
	bbsProof, err := bbsplus.CreateProofPost(pok, challengeOkm)
	require.NoError(t, err)
	// complete membership proof
	memProof, err := accumulator.CreateMembershipProofPost(mpc, challengeOkm)
	require.NoError(t, err)

	// the final proof is
	proof := &AnonymousCredentialProof{
		nonce[:],
		challengeOkm,
		bbsProof,
		memProofEntropy,
		memProof,
	}

	// verify the proof
	revealedMsgs := map[int]curves.Scalar{}
	okm, err := VerifyProof(pp, revealedMsgs, proof)
	require.NoError(t, err)
	err = crypto.IsChallengeEqual(proof.Challenge, okm)
	require.NoError(t, err)

	// add another two members and remove bob
	charlieMemberIdSeed := []byte("a member id for charlie")
	msg30 := bbsplus.Curve.Scalar.Hash(charlieMemberIdSeed)
	msg31 := bbsplus.Curve.Scalar.Hash([]byte("charlie"))
	msg32 := bbsplus.Curve.Scalar.New(40)
	msg33 := bbsplus.Curve.Scalar.Hash([]byte("edinburgh"))
	msgs3 := []curves.Scalar{msg30, msg31, msg32, msg33}
	charlieBbsSig, err := sk.BbsPlusKey.Sign(pp.BbsPlusPublicParams, msgs3)
	_ = charlieBbsSig // no more error
	require.NoError(t, err)

	mem3 := accumulator.Curve.Scalar.Hash(charlieMemberIdSeed)

	daveMemberIdSeed := []byte("a member id for dave")
	msg40 := bbsplus.Curve.Scalar.Hash(daveMemberIdSeed)
	msg41 := bbsplus.Curve.Scalar.Hash([]byte("dave"))
	msg42 := bbsplus.Curve.Scalar.New(40)
	msg43 := bbsplus.Curve.Scalar.Hash([]byte("edinburgh"))
	msgs4 := []curves.Scalar{msg40, msg41, msg42, msg43}
	daveBbsSig, err := sk.BbsPlusKey.Sign(pp.BbsPlusPublicParams, msgs4)
	_ = daveBbsSig // no more error
	require.NoError(t, err)

	mem4 := accumulator.Curve.Scalar.Hash(daveMemberIdSeed)

	adds = accumcrypto.ElementSet{Elements: []accumcrypto.Element{mem3, mem4}}
	//adds = accumcrypto.ElementSet{}
	dels = accumcrypto.ElementSet{Elements: []accumcrypto.Element{mem2}}
	app := *pp.AccumulatorPublicParams
	pp.AccumulatorPublicParams, _, err = pp.AccumulatorPublicParams.UpdateAccumulatorState(sk.AccumulatorKey, adds, dels)
	require.NoError(t, err)

	// issuer creates witness for charlie and dave
	charlieWit, err := sk.AccumulatorKey.InitMemberWitness(pp.AccumulatorPublicParams, mem3)
	_ = charlieWit // no more error
	require.NoError(t, err)
	daveWit, err := sk.AccumulatorKey.InitMemberWitness(pp.AccumulatorPublicParams, mem4)
	_ = daveWit // no more error
	require.NoError(t, err)

	// alice updates her own witness
	newAliceWit, err := accumulator.UpdateWitness(&app, pp.AccumulatorPublicParams, alice_wit)

	// alice create a new proof
	// create external blinding for membership id
	eb2, err := new(accumcrypto.ExternalBlinding).New(accumulator.Curve)
	require.NoError(t, err)
	// bbs+ proof messages
	proofMsgs2 := []common.ProofMessage{
		&common.SharedBlindingMessage{
			Message:  msgs1[0],
			Blinding: eb2.GetBlinding(),
		},
		&common.ProofSpecificMessage{
			Message: msgs1[1],
		},
		&common.ProofSpecificMessage{
			Message: msgs1[2],
		},
		&common.ProofSpecificMessage{
			Message: msgs1[3],
		},
	}
	nonce2 := []byte("a new nonce")
	// create bbs+ proof
	pok2, bbsOkm2, err := bbsplus.CreateProofPre(pp.BbsPlusPublicParams, aliceBbsSig, nonce2[:], proofMsgs2)
	require.NoError(t, err)
	// create membership proof
	mpc2, accumOkm2, memProofEntropy2, err := accumulator.CreateMembershipProofPre(pp.AccumulatorPublicParams, newAliceWit, eb2)
	require.NoError(t, err)
	// merge okm to create challenge
	challengeOkm2 := crypto.CombineChanllengeOkm(bbsOkm2, accumOkm2)
	// complete bbs+ proof
	bbsProof2, err := bbsplus.CreateProofPost(pok2, challengeOkm2)
	require.NoError(t, err)
	// complete membership proof
	memProof2, err := accumulator.CreateMembershipProofPost(mpc2, challengeOkm2)
	require.NoError(t, err)

	// the final proof is
	proof2 := &AnonymousCredentialProof{
		nonce2,
		challengeOkm2,
		bbsProof2,
		memProofEntropy2,
		memProof2,
	}

	// verify the proof
	revealedMsgs2 := map[int]curves.Scalar{}
	okm2, err := VerifyProof(pp, revealedMsgs2, proof2)
	require.NoError(t, err)
	err = crypto.IsChallengeEqual(proof2.Challenge, okm2)
	require.NoError(t, err)
}

func TestAnonymousCredentialWithRevealedMessage(t *testing.T) {
	// set up scheme parameters
	sk, pp, err := NewAnonymousCredentialSchema(4)
	require.NoError(t, err)

	// create bbs+ credential for alice
	aliceMemberIdSeed := []byte("a member id for alice")
	msg10 := bbsplus.Curve.Scalar.Hash(aliceMemberIdSeed)
	msg11 := bbsplus.Curve.Scalar.Hash([]byte("alice"))
	msg12 := bbsplus.Curve.Scalar.Hash([]byte("london"))
	msg13 := bbsplus.Curve.Scalar.New(25)
	msgs1 := []curves.Scalar{msg10, msg11, msg12, msg13}
	aliceBbsSig, err := sk.BbsPlusKey.Sign(pp.BbsPlusPublicParams, msgs1)
	require.NoError(t, err)

	// create bbs+ credential for bob
	bobMemberIdseed := []byte("a member id for bob")
	msg20 := bbsplus.Curve.Scalar.Hash(bobMemberIdseed)
	msg21 := bbsplus.Curve.Scalar.Hash([]byte("bob"))
	msg22 := bbsplus.Curve.Scalar.Hash([]byte("cambridge"))
	msg23 := bbsplus.Curve.Scalar.New(30)
	msgs2 := []curves.Scalar{msg20, msg21, msg22, msg23}
	bobBbsSig, err := sk.BbsPlusKey.Sign(pp.BbsPlusPublicParams, msgs2)
	_ = bobBbsSig // no more error
	require.NoError(t, err)

	// initialise accumulator with alice and bob
	mem1 := accumulator.Curve.Scalar.Hash(aliceMemberIdSeed)
	mem2 := accumulator.Curve.Scalar.Hash(bobMemberIdseed)
	adds := accumcrypto.ElementSet{Elements: []accumcrypto.Element{mem1, mem2}}
	dels := accumcrypto.ElementSet{}
	pp.AccumulatorPublicParams, _, err = pp.AccumulatorPublicParams.UpdateAccumulatorState(sk.AccumulatorKey, adds, dels)
	require.NoError(t, err)

	// create witness for each member
	alice_wit, err := sk.AccumulatorKey.InitMemberWitness(pp.AccumulatorPublicParams, mem1)
	require.NoError(t, err)
	bob_wit, err := sk.AccumulatorKey.InitMemberWitness(pp.AccumulatorPublicParams, mem2)
	_ = bob_wit // no more error
	require.NoError(t, err)

	// anonymous crendential for alice is (alice_bbs_sig, alice_wit), (bob_bbs_sig, bob_wit)

	// alice creates a zkp using her credential (alice_bbs_sig, alice_wit)
	// nonce should be from the verifier or a certain event tag
	var nonce [32]byte
	cnt, err := crand.Read(nonce[:])
	require.NoError(t, err)
	require.Equal(t, cnt, 32)

	// create external blinding for membership id
	eb, err := new(accumcrypto.ExternalBlinding).New(accumulator.Curve)
	require.NoError(t, err)

	// bbs+ proof messages
	proofMsgs := []common.ProofMessage{
		&common.SharedBlindingMessage{
			Message:  msgs1[0],
			Blinding: eb.GetBlinding(),
		},
		&common.RevealedMessage{
			Message: msgs1[1],
		},
		&common.ProofSpecificMessage{
			Message: msgs1[2],
		},
		&common.ProofSpecificMessage{
			Message: msgs1[3],
		},
	}
	// create bbs+ proof
	pok, bbsOkm, err := bbsplus.CreateProofPre(pp.BbsPlusPublicParams, aliceBbsSig, nonce[:], proofMsgs)
	require.NoError(t, err)
	// create membership proof
	mpc, accumOkm, memProofEntropy, err := accumulator.CreateMembershipProofPre(pp.AccumulatorPublicParams, alice_wit, eb)
	require.NoError(t, err)
	// merge okm to create challenge
	challengeOkm := crypto.CombineChanllengeOkm(bbsOkm, accumOkm)
	// complete bbs+ proof
	bbsProof, err := bbsplus.CreateProofPost(pok, challengeOkm)
	require.NoError(t, err)
	// complete membership proof
	memProof, err := accumulator.CreateMembershipProofPost(mpc, challengeOkm)
	require.NoError(t, err)

	// the final proof is
	proof := &AnonymousCredentialProof{
		nonce[:],
		challengeOkm,
		bbsProof,
		memProofEntropy,
		memProof,
	}

	// verify the proof
	revealedMsgs := map[int]curves.Scalar{1: msgs1[1]}
	okm, err := VerifyProof(pp, revealedMsgs, proof)
	require.NoError(t, err)
	err = crypto.IsChallengeEqual(proof.Challenge, okm)
	require.NoError(t, err)
}

func TestRevocationOnly(t *testing.T) {
	// set up scheme parameters
	sk, pp, err := NewAnonymousCredentialSchema(4)
	require.NoError(t, err)

	// create bbs+ credential for alice
	aliceMemberIdSeed := []byte("a member id for alice")
	msg10 := bbsplus.Curve.Scalar.Hash(aliceMemberIdSeed)
	msg11 := bbsplus.Curve.Scalar.Hash([]byte("alice"))
	msg12 := bbsplus.Curve.Scalar.New(25)
	msg13 := bbsplus.Curve.Scalar.Hash([]byte("london"))
	msgs1 := []curves.Scalar{msg10, msg11, msg12, msg13}
	aliceBbsSig, err := sk.BbsPlusKey.Sign(pp.BbsPlusPublicParams, msgs1)
	require.NoError(t, err)

	// create bbs+ credential for bob
	bobMemberIdseed := []byte("a member id for bob")
	msg20 := bbsplus.Curve.Scalar.Hash(bobMemberIdseed)
	msg21 := bbsplus.Curve.Scalar.Hash([]byte("bob"))
	msg22 := bbsplus.Curve.Scalar.New(30)
	msg23 := bbsplus.Curve.Scalar.Hash([]byte("cambridge"))
	msgs2 := []curves.Scalar{msg20, msg21, msg22, msg23}
	bobBbsSig, err := sk.BbsPlusKey.Sign(pp.BbsPlusPublicParams, msgs2)
	_ = bobBbsSig // no more error
	require.NoError(t, err)

	// initialise accumulator with alice and bob
	mem1 := accumulator.Curve.Scalar.Hash(aliceMemberIdSeed)
	mem2 := accumulator.Curve.Scalar.Hash(bobMemberIdseed)
	adds := accumcrypto.ElementSet{Elements: []accumcrypto.Element{mem1, mem2}}
	dels := accumcrypto.ElementSet{}
	pp.AccumulatorPublicParams, _, err = pp.AccumulatorPublicParams.UpdateAccumulatorState(sk.AccumulatorKey, adds, dels)
	require.NoError(t, err)

	// create witness for each member
	alice_wit, err := sk.AccumulatorKey.InitMemberWitness(pp.AccumulatorPublicParams, mem1)
	require.NoError(t, err)
	bob_wit, err := sk.AccumulatorKey.InitMemberWitness(pp.AccumulatorPublicParams, mem2)
	_ = bob_wit // no more error
	require.NoError(t, err)

	// anonymous crendential for alice is (alice_bbs_sig, alice_wit), (bob_bbs_sig, bob_wit)

	// remove bob
	adds = accumcrypto.ElementSet{}
	dels = accumcrypto.ElementSet{Elements: []accumcrypto.Element{mem2}}
	app := *pp.AccumulatorPublicParams
	pp.AccumulatorPublicParams, _, err = pp.AccumulatorPublicParams.UpdateAccumulatorState(sk.AccumulatorKey, adds, dels)
	require.NoError(t, err)

	// alice updates her own witness
	newAliceWit, err := accumulator.UpdateWitness(&app, pp.AccumulatorPublicParams, alice_wit)
	require.NoError(t, err)

	// alice create a new proof
	// create external blinding for membership id
	eb, err := new(accumcrypto.ExternalBlinding).New(accumulator.Curve)
	require.NoError(t, err)
	// bbs+ proof messages
	proofMsgs := []common.ProofMessage{
		&common.SharedBlindingMessage{
			Message:  msgs1[0],
			Blinding: eb.GetBlinding(),
		},
		&common.ProofSpecificMessage{
			Message: msgs1[1],
		},
		&common.ProofSpecificMessage{
			Message: msgs1[2],
		},
		&common.ProofSpecificMessage{
			Message: msgs1[3],
		},
	}
	nonce := []byte("a new nonce")
	// create bbs+ proof
	pok, bbsOkm, err := bbsplus.CreateProofPre(pp.BbsPlusPublicParams, aliceBbsSig, nonce[:], proofMsgs)
	require.NoError(t, err)
	// create membership proof
	mpc, accumOkm, memProofEntropy, err := accumulator.CreateMembershipProofPre(pp.AccumulatorPublicParams, newAliceWit, eb)
	require.NoError(t, err)
	// merge okm to create challenge
	challengeOkm := crypto.CombineChanllengeOkm(bbsOkm, accumOkm)
	// complete bbs+ proof
	bbsProof, err := bbsplus.CreateProofPost(pok, challengeOkm)
	require.NoError(t, err)
	// complete membership proof
	memProof, err := accumulator.CreateMembershipProofPost(mpc, challengeOkm)
	require.NoError(t, err)

	// the final proof is
	proof := &AnonymousCredentialProof{
		nonce,
		challengeOkm,
		bbsProof,
		memProofEntropy,
		memProof,
	}

	// verify the proof
	revealedMsgs := map[int]curves.Scalar{}
	okm, err := VerifyProof(pp, revealedMsgs, proof)
	require.NoError(t, err)
	err = crypto.IsChallengeEqual(proof.Challenge, okm)
	require.NoError(t, err)
}

func TestMultipleUpdateWitness(t *testing.T) {
	// set up scheme parameters
	sk, pp, err := NewAnonymousCredentialSchema(4)
	require.NoError(t, err)

	// create bbs+ credential for alice
	aliceMemberIdSeed := []byte("a member id for alice")
	msg10 := bbsplus.Curve.Scalar.Hash(aliceMemberIdSeed)
	msg11 := bbsplus.Curve.Scalar.Hash([]byte("alice"))
	msg12 := bbsplus.Curve.Scalar.New(25)
	msg13 := bbsplus.Curve.Scalar.Hash([]byte("london"))
	msgs1 := []curves.Scalar{msg10, msg11, msg12, msg13}
	aliceBbsSig, err := sk.BbsPlusKey.Sign(pp.BbsPlusPublicParams, msgs1)
	require.NoError(t, err)

	// create bbs+ credential for bob
	bobMemberIdseed := []byte("a member id for bob")
	msg20 := bbsplus.Curve.Scalar.Hash(bobMemberIdseed)
	msg21 := bbsplus.Curve.Scalar.Hash([]byte("bob"))
	msg22 := bbsplus.Curve.Scalar.New(30)
	msg23 := bbsplus.Curve.Scalar.Hash([]byte("cambridge"))
	msgs2 := []curves.Scalar{msg20, msg21, msg22, msg23}
	bobBbsSig, err := sk.BbsPlusKey.Sign(pp.BbsPlusPublicParams, msgs2)
	_ = bobBbsSig // no more error
	require.NoError(t, err)

	// initialise accumulator with alice and bob
	mem1 := accumulator.Curve.Scalar.Hash(aliceMemberIdSeed)
	mem2 := accumulator.Curve.Scalar.Hash(bobMemberIdseed)
	adds := accumcrypto.ElementSet{Elements: []accumcrypto.Element{mem1, mem2}}
	dels := accumcrypto.ElementSet{}
	pp.AccumulatorPublicParams, _, err = pp.AccumulatorPublicParams.UpdateAccumulatorState(sk.AccumulatorKey, adds, dels)
	require.NoError(t, err)

	// create witness for each member
	alice_wit, err := sk.AccumulatorKey.InitMemberWitness(pp.AccumulatorPublicParams, mem1)
	require.NoError(t, err)
	bob_wit, err := sk.AccumulatorKey.InitMemberWitness(pp.AccumulatorPublicParams, mem2)
	_ = bob_wit // no more error
	require.NoError(t, err)

	// anonymous crendential for alice is (alice_bbs_sig, alice_wit), (bob_bbs_sig, bob_wit)

	// keep a copy of the old parameters
	app := *pp.AccumulatorPublicParams

	// remove bob
	adds = accumcrypto.ElementSet{}
	dels = accumcrypto.ElementSet{Elements: []accumcrypto.Element{mem2}}
	pp.AccumulatorPublicParams, _, err = pp.AccumulatorPublicParams.UpdateAccumulatorState(sk.AccumulatorKey, adds, dels)
	require.NoError(t, err)

	// add another two members
	charlieMemberIdSeed := []byte("a member id for charlie")
	msg30 := bbsplus.Curve.Scalar.Hash(charlieMemberIdSeed)
	msg31 := bbsplus.Curve.Scalar.Hash([]byte("charlie"))
	msg32 := bbsplus.Curve.Scalar.New(40)
	msg33 := bbsplus.Curve.Scalar.Hash([]byte("edinburgh"))
	msgs3 := []curves.Scalar{msg30, msg31, msg32, msg33}
	charlieBbsSig, err := sk.BbsPlusKey.Sign(pp.BbsPlusPublicParams, msgs3)
	_ = charlieBbsSig // no more error
	require.NoError(t, err)

	mem3 := accumulator.Curve.Scalar.Hash(charlieMemberIdSeed)

	daveMemberIdSeed := []byte("a member id for dave")
	msg40 := bbsplus.Curve.Scalar.Hash(daveMemberIdSeed)
	msg41 := bbsplus.Curve.Scalar.Hash([]byte("dave"))
	msg42 := bbsplus.Curve.Scalar.New(40)
	msg43 := bbsplus.Curve.Scalar.Hash([]byte("edinburgh"))
	msgs4 := []curves.Scalar{msg40, msg41, msg42, msg43}
	daveBbsSig, err := sk.BbsPlusKey.Sign(pp.BbsPlusPublicParams, msgs4)
	_ = daveBbsSig // no more error
	require.NoError(t, err)

	mem4 := accumulator.Curve.Scalar.Hash(daveMemberIdSeed)

	adds = accumcrypto.ElementSet{Elements: []accumcrypto.Element{mem3, mem4}}
	//adds = accumcrypto.ElementSet{}
	dels = accumcrypto.ElementSet{}
	pp.AccumulatorPublicParams, _, err = pp.AccumulatorPublicParams.UpdateAccumulatorState(sk.AccumulatorKey, adds, dels)
	require.NoError(t, err)

	// alice updates her witness
	newAliceWit, err := accumulator.UpdateWitness(&app, pp.AccumulatorPublicParams, alice_wit)

	// alice create a new proof
	// create external blinding for membership id
	eb, err := new(accumcrypto.ExternalBlinding).New(accumulator.Curve)
	require.NoError(t, err)
	// bbs+ proof messages
	proofMsgs := []common.ProofMessage{
		&common.SharedBlindingMessage{
			Message:  msgs1[0],
			Blinding: eb.GetBlinding(),
		},
		&common.ProofSpecificMessage{
			Message: msgs1[1],
		},
		&common.ProofSpecificMessage{
			Message: msgs1[2],
		},
		&common.ProofSpecificMessage{
			Message: msgs1[3],
		},
	}
	nonce := []byte("a new nonce")
	// create bbs+ proof
	pok, bbsOkm, err := bbsplus.CreateProofPre(pp.BbsPlusPublicParams, aliceBbsSig, nonce[:], proofMsgs)
	require.NoError(t, err)
	// create membership proof
	mpc, accumOkm, memProofEntropy, err := accumulator.CreateMembershipProofPre(pp.AccumulatorPublicParams, newAliceWit, eb)
	require.NoError(t, err)
	// merge okm to create challenge
	challengeOkm := crypto.CombineChanllengeOkm(bbsOkm, accumOkm)
	// complete bbs+ proof
	bbsProof, err := bbsplus.CreateProofPost(pok, challengeOkm)
	require.NoError(t, err)
	// complete membership proof
	memProof, err := accumulator.CreateMembershipProofPost(mpc, challengeOkm)
	require.NoError(t, err)

	// the final proof is
	proof := &AnonymousCredentialProof{
		nonce,
		challengeOkm,
		bbsProof,
		memProofEntropy,
		memProof,
	}

	// verify the proof
	revealedMsgs := map[int]curves.Scalar{}
	okm, err := VerifyProof(pp, revealedMsgs, proof)
	require.NoError(t, err)
	err = crypto.IsChallengeEqual(proof.Challenge, okm)
	require.NoError(t, err)
}
