package traces

import (
	"fmt"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
)

func NewDenomTracesMigrator(ibcStore storetypes.KVStore) Migrator {
	return Migrator{ibcStore: ibcStore}
}

// DenomTraces is a struct for handling in-place store migrations.
type Migrator struct {
	ibcStore storetypes.KVStore
}

func equalTraces(dtA, dtB DenomTrace) bool {
	return dtA.BaseDenom == dtB.BaseDenom && dtA.Path == dtB.Path
}

// IterateDenomTraces iterates over the denomination traces in the store
// and performs a callback function.
func (m Migrator) IterateDenomTraces(ctx sdk.Context, cdc codec.Codec, cb func(denomTrace DenomTrace) bool) {

	iterator := storetypes.KVStorePrefixIterator(m.ibcStore, ibctransfertypes.DenomTraceKey)

	defer sdk.LogDeferred(ctx.Logger(), func() error { return iterator.Close() })
	for ; iterator.Valid(); iterator.Next() {
		denomTrace := MustUnmarshalDenomTrace(iterator.Value(), cdc)
		if cb(denomTrace) {
			break
		}
	}
}

// MustUnmarshalDenomTrace attempts to decode and return an DenomTrace object from
// raw encoded bytes. It panics on error.
func MustUnmarshalDenomTrace(bz []byte, cdc codec.Codec) DenomTrace {
	var denomTrace DenomTrace

	cdc.MustUnmarshal(bz, &denomTrace)
	return denomTrace
}

// MustMarshalDenomTrace attempts to encode an DenomTrace object and returns the
// raw encoded bytes. It panics on error.
func MustMarshalDenomTrace(denomTrace DenomTrace, cdc codec.Codec) []byte {
	return cdc.MustMarshal(&denomTrace)
}

// SetDenomTrace sets a new {trace hash -> denom trace} pair to the store.
func (m Migrator) SetDenomTrace(ctx sdk.Context, cdc codec.Codec, denomTrace DenomTrace) {
	bz := MustMarshalDenomTrace(denomTrace, cdc)
	m.ibcStore.Set(denomTrace.Hash(), bz)
}

// MigrateTraces migrates the DenomTraces to the correct format, accounting for slashes in the BaseDenom.
func (m Migrator) MigrateTraces(ctx sdk.Context) error {
	ir := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(ir)

	// list of traces that must replace the old traces in store
	var newTraces []DenomTrace
	m.IterateDenomTraces(ctx, cdc,
		func(dt DenomTrace) (stop bool) {
			// check if the new way of splitting FullDenom
			// is the same as the current DenomTrace.
			// If it isn't then store the new DenomTrace in the list of new traces.
			newTrace := ParseDenomTrace(dt.GetFullDenomPath())
			err := newTrace.Validate()
			if err != nil {
				panic(err)
			}

			if dt.IBCDenom() != newTrace.IBCDenom() {
				// The new form of parsing will result in a token denomination change.
				// A bank migration is required. A panic should occur to prevent the
				// chain from using corrupted state.
				panic(fmt.Errorf("migration will result in corrupted state. Previous IBC token (%s) requires a bank migration. Expected denom trace (%s)", dt, newTrace))
			}

			if !equalTraces(newTrace, dt) {
				newTraces = append(newTraces, newTrace)
			}

			return false
		})

	// replace the outdated traces with the new trace information
	for _, nt := range newTraces {
		m.SetDenomTrace(ctx, cdc, nt)
	}
	return nil
}
