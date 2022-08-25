package crypto

import (
	"bytes"
	"fmt"
)

const ChallengePrefix = "Create Challenge Bytes"
const DomainSeparator = "dom-sep-challenge"

func IsChallengeEqual(challengeOkm []byte, okm []byte) error {
	if bytes.Equal(challengeOkm, okm) {
		return nil
	}
	return fmt.Errorf("anonymous credential bbs+/membership proof verification (challenge) failed")
}

func CombineChanllengeOkm(okm ...[]byte) []byte {
	x := []byte(ChallengePrefix)
	for _, y := range okm {
		x = append(x, []byte(DomainSeparator)...)
		x = append(x, y...)
	}
	return x
}
