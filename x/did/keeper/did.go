package keeper

import (
	"strings"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/fetchai/fetchd/x/did/types"
)

func (k Keeper) SetDidDocument(ctx sdk.Context, key []byte, document types.DidDocument) {
	k.Set(ctx, key, types.DidDocumentKey, document, k.Marshal)
}

func (k Keeper) GetDidDocument(ctx sdk.Context, key []byte) (types.DidDocument, bool) {
	val, found := k.Get(ctx, key, types.DidDocumentKey, k.UnmarshalDidDocument)
	return val.(types.DidDocument), found
}

// UnmarshalDidDocument unmarshall a did document= and check if it is empty
// ad DID document is empty if contains no context
func (k Keeper) UnmarshalDidDocument(value []byte) (interface{}, bool) {
	data := types.DidDocument{}
	k.Unmarshal(value, &data)
	return data, types.IsValidDIDDocument(&data)
}

func (k Keeper) SetDidMetadata(ctx sdk.Context, key []byte, meta types.DidMetadata) {
	k.Set(ctx, key, types.DidMetadataKey, meta, k.Marshal)
}

func (k Keeper) GetDidMetadata(ctx sdk.Context, key []byte) (types.DidMetadata, bool) {
	val, found := k.Get(ctx, key, types.DidMetadataKey, k.UnmarshalDidMetadata)
	return val.(types.DidMetadata), found
}

func (k Keeper) UnmarshalDidMetadata(value []byte) (interface{}, bool) {
	data := types.DidMetadata{}
	k.Unmarshal(value, &data)
	return data, types.IsValidDIDMetadata(&data)
}

// ResolveDid returning the did document and associated metadata
func (k Keeper) ResolveDid(ctx sdk.Context, did types.DID) (doc types.DidDocument, meta types.DidMetadata, err error) {
	if strings.HasPrefix(did.String(), types.DidKeyPrefix) {
		doc, meta, err = types.ResolveAccountDID(did.String(), ctx.ChainID())
		return
	}
	doc, found := k.GetDidDocument(ctx, []byte(did.String()))
	if !found {
		err = types.ErrDidDocumentNotFound
		return
	}
	meta, _ = k.GetDidMetadata(ctx, []byte(did.String()))
	return
}

func (k Keeper) Marshal(value interface{}) (bytes []byte) {
	switch value := value.(type) {
	case types.DidDocument:
		bytes = k.cdc.MustMarshal(&value)
	case types.DidMetadata:
		bytes = k.cdc.MustMarshal(&value)
	}
	return
}

// Unmarshal unmarshal a byte slice to a struct, return false in case of errors
func (k Keeper) Unmarshal(data []byte, val codec.ProtoMarshaler) bool {
	if len(data) == 0 {
		return false
	}
	if err := k.cdc.Unmarshal(data, val); err != nil {
		return false
	}
	return true
}

// GetAllDidDocumentsWithCondition retrieve a list of
// did document by some arbitrary criteria. The selector filter has access
// to both the did and its metadata
func (k Keeper) GetAllDidDocumentsWithCondition(
	ctx sdk.Context,
	key []byte,
	didSelector func(did types.DidDocument) bool,
) (didDocs []types.DidDocument) {
	iterator := k.GetAll(ctx, key)

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		did, _ := k.UnmarshalDidDocument(iterator.Value())
		didTyped := did.(types.DidDocument)

		if didSelector(didTyped) {
			didDocs = append(didDocs, didTyped)
		}
	}

	return didDocs
}

// GetAllDidDocuments returns all the DidDocuments
func (k Keeper) GetAllDidDocuments(ctx sdk.Context) []types.DidDocument {
	return k.GetAllDidDocumentsWithCondition(
		ctx,
		types.DidDocumentKey,
		func(did types.DidDocument) bool { return true },
	)
}

// GetDidDocumentsByPubKey retrieve a did document using a pubkey associated to the DID
// TODO: this function is used only in the issuer module ante handler !
func (k Keeper) GetDidDocumentsByPubKey(ctx sdk.Context, pubkey cryptotypes.PubKey) (dids []types.DidDocument) {

	dids = k.GetAllDidDocumentsWithCondition(
		ctx,
		types.DidDocumentKey,
		func(did types.DidDocument) bool {
			return did.HasPublicKey(pubkey)
		},
	)
	// compute the key did

	// generate the address
	addr, err := sdk.Bech32ifyAddressBytes(sdk.GetConfig().GetBech32AccountAddrPrefix(), pubkey.Address())
	if err != nil {
		return
	}
	doc, _, err := types.ResolveAccountDID(types.NewKeyDID(addr).String(), ctx.ChainID())
	if err != nil {
		return
	}
	dids = append(dids, doc)
	return
}

func (k Keeper) VerifyDidWithRelationships(ctx sdk.Context, constraints []string, did, signer string) (err error) {
	k.Logger(ctx).Info("verify a did document", "target did", did)
	// Check to see if the provided did is in the store
	doc, _, err := k.ResolveDid(ctx, types.DID(did))
	if err != nil {
		return
	}

	// Check to see if the msg signer has a verification relationship in the did document
	if !doc.HasRelationship(types.NewBlockchainAccountID(ctx.ChainID(), signer), constraints...) {
		err = sdkerrors.Wrapf(
			types.ErrUnauthorized,
			"signer account %s not authorized to update the target did document at %s",
			signer, did,
		)
		k.Logger(ctx).Error(err.Error())
		return
	}

	k.Logger(ctx).Info("Verified relationship from did document for", "did", did, "controller", signer)

	return
}
