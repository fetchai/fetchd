package bls12381

import (
	"encoding/base64"
	"fmt"

	blst "github.com/supranational/blst/bindings/go"
)

func aggregatePublicKey(pks []*PubKey) (*blst.P1Affine, error) {
	pubkeys := make([]*blst.P1Affine, len(pks))
	for i, pk := range pks {
		pubkeys[i] = new(blst.P1Affine).Deserialize(pk.Key)
		if pubkeys[i] == nil {
			return nil, fmt.Errorf("failed to deserialize public key")
		}
	}

	aggregator := new(blst.P1Aggregate)
	b := aggregator.Aggregate(pubkeys, false)
	if !b {
		return nil, fmt.Errorf("failed to aggregate public keys")
	}
	apk := aggregator.ToAffine()

	return apk, nil
}

// AggregateSignature combines a set of verified signatures into a single bls signature
func AggregateSignature(sigs [][]byte) ([]byte, error) {
	sigmas := make([]*blst.P2Affine, len(sigs))
	for i, sig := range sigs {
		sigmas[i] = new(blst.P2Affine).Uncompress(sig)
		if sigmas[i] == nil {
			return nil, fmt.Errorf("failed to deserialize the %d-th signature", i)
		}
	}

	aggregator := new(blst.P2Aggregate)
	b := aggregator.Aggregate(sigmas, false)
	if !b {
		return nil, fmt.Errorf("failed to aggregate signatures")
	}
	aggSigBytes := aggregator.ToAffine().Compress()
	return aggSigBytes, nil
}

// VerifyMultiSignature assumes public key is already validated
func VerifyMultiSignature(msg []byte, sig []byte, pks []*PubKey) error {
	return VerifyAggregateSignature([][]byte{msg}, false, sig, [][]*PubKey{pks})
}

func Unique(msgs [][]byte) bool {
	if len(msgs) <= 1 {
		return true
	}
	msgMap := make(map[string]bool, len(msgs))
	for _, msg := range msgs {
		s := base64.StdEncoding.EncodeToString(msg)
		if _, ok := msgMap[s]; ok {
			return false
		}
		msgMap[s] = true
	}
	return true
}

func VerifyAggregateSignature(msgs [][]byte, msgCheck bool, sig []byte, pkss [][]*PubKey) error {
	n := len(msgs)
	if n == 0 {
		return fmt.Errorf("messages cannot be empty")
	}

	for i, msg := range msgs {
		if len(msg) == 0 {
			return fmt.Errorf("%d-th message is empty", i)
		}
	}

	if len(pkss) != n {
		return fmt.Errorf("the number of messages and public key sets must match")
	}

	for i, pks := range pkss {
		if len(pks) == 0 {
			return fmt.Errorf("%d-th public key set is empty", i)
		}
	}

	if msgCheck {
		if !Unique(msgs) {
			return fmt.Errorf("messages must be pairwise distinct")
		}
	}

	apks := make([]*blst.P1Affine, len(pkss))
	for i, pks := range pkss {
		apk, err := aggregatePublicKey(pks)
		if err != nil {
			return fmt.Errorf("cannot aggregate public keys: %s", err.Error())
		}
		apks[i] = apk
	}

	sigma := new(blst.P2Affine).Uncompress(sig)
	if sigma == nil {
		return fmt.Errorf("failed to deserialize signature")
	}

	dst := []byte("BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_POP_")
	if !sigma.AggregateVerify(true, apks, false, msgs, dst) {
		return fmt.Errorf("failed to verify signature")
	}

	return nil
}
