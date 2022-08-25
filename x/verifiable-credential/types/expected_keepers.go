package types

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fetchai/fetchd/x/did/types"
)

// DidKeeper defines the expected did keeper functions
type DidKeeper interface {
	GetDidDocument(ctx sdk.Context, key []byte) (types.DidDocument, bool)
	ResolveDid(ctx sdk.Context, did types.DID) (doc types.DidDocument, meta types.DidMetadata, err error)
	VerifyDidWithRelationships(ctx sdk.Context, constraints []string, did, signer string) (err error)
}

// AccountKeeper defines the functions from the account keeper
type AccountKeeper interface {
	GetPubKey(ctx sdk.Context, addr sdk.AccAddress) (cryptotypes.PubKey, error)
}

// VcKeeper defines the expected verfiable credentials keeper functions
type VcKeeper interface {
	GetVerifiableCredential(ctx sdk.Context, key []byte) (VerifiableCredential, bool)
	SetVerifiableCredential(ctx sdk.Context, key []byte, vc VerifiableCredential) error
	GetVerifiableCredentialWithType(ctx sdk.Context, subjectDID, vcType string) []VerifiableCredential
}
