package bls12381

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlsMultiSig(t *testing.T) {
	total := 5
	pks := make([]*PubKey, total)
	sigs := make([][]byte, total)
	msg := []byte("hello world")
	for i := 0; i < total; i++ {
		sk := GenPrivKey()
		pk, ok := sk.PubKey().(*PubKey)
		require.True(t, ok)

		sig, err := sk.Sign(msg)
		require.Nil(t, err)

		pks[i] = pk
		sigs[i] = sig
	}

	aggSig, err := AggregateSignature(sigs)
	require.Nil(t, err)

	assert.Nil(t, VerifyMultiSignature(msg, aggSig, pks))

}

func TestBlsAggSig(t *testing.T) {
	total := 5
	pks := make([][]*PubKey, total)
	sigs := make([][]byte, total)
	msgs := make([][]byte, total)
	for i := 0; i < total; i++ {
		msgs[i] = []byte(fmt.Sprintf("message %d", i))
		sk := GenPrivKey()
		pk, ok := sk.PubKey().(*PubKey)
		require.True(t, ok)

		sig, err := sk.Sign(msgs[i])
		require.Nil(t, err)

		pks[i] = []*PubKey{pk}
		sigs[i] = sig
	}

	aggSig, err := AggregateSignature(sigs)
	require.Nil(t, err)

	assert.Nil(t, VerifyAggregateSignature(msgs, true, aggSig, pks))
}

func benchmarkBlsVerifyMulti(total int, b *testing.B) {
	pks := make([]*PubKey, total)
	sigs := make([][]byte, total)
	msg := []byte("hello world")
	for i := 0; i < total; i++ {
		sk := GenPrivKey()
		pk, ok := sk.PubKey().(*PubKey)
		require.True(b, ok)

		sig, err := sk.Sign(msg)
		require.Nil(b, err)

		pks[i] = pk
		sigs[i] = sig
	}

	aggSig, err := AggregateSignature(sigs)
	require.Nil(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		require.NoError(b, VerifyMultiSignature(msg, aggSig, pks))
	}
}

func BenchmarkBlsVerifyMulti8(b *testing.B)   { benchmarkBlsVerifyMulti(8, b) }
func BenchmarkBlsVerifyMulti16(b *testing.B)  { benchmarkBlsVerifyMulti(16, b) }
func BenchmarkBlsVerifyMulti32(b *testing.B)  { benchmarkBlsVerifyMulti(32, b) }
func BenchmarkBlsVerifyMulti64(b *testing.B)  { benchmarkBlsVerifyMulti(64, b) }
func BenchmarkBlsVerifyMulti128(b *testing.B) { benchmarkBlsVerifyMulti(128, b) }

func benchmarkBlsVerifyAgg(total int, b *testing.B) {
	pks := make([][]*PubKey, total)
	sigs := make([][]byte, total)
	msgs := make([][]byte, total)
	for i := 0; i < total; i++ {
		msgs[i] = []byte(fmt.Sprintf("message %d", i))
		sk := GenPrivKey()
		pk, ok := sk.PubKey().(*PubKey)
		require.True(b, ok)

		sig, err := sk.Sign(msgs[i])
		require.Nil(b, err)

		pks[i] = []*PubKey{pk}
		sigs[i] = sig
	}

	aggSig, err := AggregateSignature(sigs)
	require.Nil(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		require.NoError(b, VerifyAggregateSignature(msgs, false, aggSig, pks))
	}
}

func BenchmarkBlsVerifyAgg8(b *testing.B)   { benchmarkBlsVerifyAgg(8, b) }
func BenchmarkBlsVerifyAgg16(b *testing.B)  { benchmarkBlsVerifyAgg(16, b) }
func BenchmarkBlsVerifyAgg32(b *testing.B)  { benchmarkBlsVerifyAgg(32, b) }
func BenchmarkBlsVerifyAgg64(b *testing.B)  { benchmarkBlsVerifyAgg(64, b) }
func BenchmarkBlsVerifyAgg128(b *testing.B) { benchmarkBlsVerifyAgg(128, b) }
