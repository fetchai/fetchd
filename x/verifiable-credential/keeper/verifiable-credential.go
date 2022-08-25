package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	didtypes "github.com/fetchai/fetchd/x/did/types"
	"github.com/fetchai/fetchd/x/verifiable-credential/types"
)

// SetVerifiableCredential commit a verifiable credential to the storage
func (q Keeper) SetVerifiableCredential(ctx sdk.Context, key []byte, vc types.VerifiableCredential) (err error) {
	if err = ValidateProof(ctx, q, vc, didtypes.Authentication, didtypes.AssertionMethod); err != nil {
		return
	}
	q.Set(ctx, key, types.VerifiableCredentialKey, vc, q.MarshalVerifiableCredential)
	return
}

// GetVerifiableCredential regurns the verifiable credentials associated with an key
func (q Keeper) GetVerifiableCredential(ctx sdk.Context, key []byte) (types.VerifiableCredential, bool) {
	val, found := q.Get(ctx, key, types.VerifiableCredentialKey, q.UnmarshalVerifiableCredential)
	return val.(types.VerifiableCredential), found
}

func (q Keeper) UnmarshalVerifiableCredential(value []byte) (interface{}, bool) {
	vc := types.VerifiableCredential{}

	err := q.cdc.Unmarshal(value, &vc)
	if err != nil {
		return types.VerifiableCredential{}, false
	}

	if vc.Id == "" {
		return types.VerifiableCredential{}, false
	}

	return vc, true
}

func (q Keeper) MarshalVerifiableCredential(value interface{}) []byte {
	did := value.(types.VerifiableCredential)

	bytes, _ := q.cdc.Marshal(&did)

	return bytes
}

func (q Keeper) GetAllVerifiableCredentialsWithCondition(
	ctx sdk.Context,
	key []byte,
	vcSelector func(votes types.VerifiableCredential) bool,
) (vcs []types.VerifiableCredential) {
	iterator := q.GetAll(ctx, key)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		vc, _ := q.UnmarshalVerifiableCredential(iterator.Value())
		vcTyped := vc.(types.VerifiableCredential)
		if vcSelector(vcTyped) {
			vcs = append(vcs, vcTyped)
		}
	}

	return vcs
}

// GetVerifiableCredentialWithType returns the list of verifiable credential of a certain type
// for an holder (the subject of the credential)
func (q Keeper) GetVerifiableCredentialWithType(ctx sdk.Context, subjectDID, vcType string) (vcs []types.VerifiableCredential) {
	return q.GetAllVerifiableCredentialsWithCondition(ctx, types.VerifiableCredentialKey, func(vc types.VerifiableCredential) bool {
		if vc.GetSubjectDID().String() == subjectDID && vc.HasType(vcType) {
			return true
		}
		return false
	})
}

// GetAllVerifiableCredentialsByIssuer returns all verifiable credentials issued by an issuer
func (q Keeper) GetAllVerifiableCredentialsByIssuer(ctx sdk.Context, issuerDID string) []types.VerifiableCredential {
	return q.GetAllVerifiableCredentialsWithCondition(
		ctx,
		types.VerifiableCredentialKey,
		func(vc types.VerifiableCredential) bool { return issuerDID == vc.Issuer },
	)
}

func (q Keeper) GetAllVerifiableCredentials(ctx sdk.Context) []types.VerifiableCredential {
	return q.GetAllVerifiableCredentialsWithCondition(
		ctx,
		types.VerifiableCredentialKey,
		func(vc types.VerifiableCredential) bool { return true },
	)
}

func (q Keeper) SetVcMetadata(ctx sdk.Context, key []byte, meta types.VcMetadata) {
	q.Set(ctx, key, types.VcMetadataKey, meta, q.MarshalVcMetadata)
}

func (q Keeper) GetVcMetadata(ctx sdk.Context, key []byte) (types.VcMetadata, bool) {
	val, found := q.Get(ctx, key, types.VcMetadataKey, q.UnmarshalVcMetadata)
	return val.(types.VcMetadata), found
}

func (q Keeper) MarshalVcMetadata(value interface{}) []byte {
	meta := value.(types.VcMetadata)
	bytes, _ := q.cdc.Marshal(&meta)

	return bytes
}

func (q Keeper) UnmarshalVcMetadata(value []byte) (interface{}, bool) {
	data := types.VcMetadata{}
	err := q.cdc.Unmarshal(value, &data)
	if err != nil {
		return types.VcMetadata{}, false
	}

	return data, types.IsValidVcMetadata(&data)
}

// ValidateProof validate the proof of a verifiable credential
func ValidateProof(ctx sdk.Context, k Keeper, vc types.VerifiableCredential, verificationRelationships ...string) error {
	// resolve the subject
	_, _, err := k.didKeeper.ResolveDid(ctx, vc.GetSubjectDID())
	if err != nil {
		return sdkerrors.Wrapf(
			err, "subject DID is not resolvable",
		)
	}
	// resolve the issuer
	doc, _, err := k.didKeeper.ResolveDid(ctx, vc.GetIssuerDID())
	if err != nil {
		return sdkerrors.Wrapf(
			err, "issuer DID is not resolvable",
		)
	}
	// verify the signature
	if vc.Proof == nil {
		return sdkerrors.Wrapf(
			types.ErrMessageSigner,
			"proof is nil %v",
			err,
		)
	}
	//check relationships
	authorized := false
	methodRelationships := doc.GetVerificationRelationships(vc.Proof.VerificationMethod)
Outer:
	for _, gotR := range methodRelationships {
		for _, wantR := range verificationRelationships {
			if gotR == wantR {
				authorized = true
				break Outer
			}
		}
	}
	// verify the relationships
	if !authorized {
		return sdkerrors.Wrapf(
			types.ErrMessageSigner,
			"unauthorized, verification method ID not listed in any of the required relationships in the issuer did (want %v, got %v) ", verificationRelationships, methodRelationships,
		)
	}
	// get the address in the verification method
	issuerAddress, err := doc.GetVerificationMethodBlockchainAddress(vc.Proof.VerificationMethod)
	if err != nil {
		return sdkerrors.Wrapf(
			types.ErrMessageSigner,
			"the issuer address cannot be retrieved due to %v",
			err,
		)
	}

	// verify that is the same of the vc
	issuerAccount, err := sdk.AccAddressFromBech32(issuerAddress)
	if err != nil {
		return sdkerrors.Wrapf(
			types.ErrMessageSigner,
			"failed to convert the issuer address to account %v: %v", issuerAddress,
			err,
		)
	}
	// get the public key from the account
	pk, err := k.accountKeeper.GetPubKey(ctx, issuerAccount)
	if err != nil || pk == nil {
		return sdkerrors.Wrapf(
			types.ErrMessageSigner,
			"issuer public key not found %v",
			err,
		)
	}
	err = vc.Validate(pk)
	if err != nil {
		return sdkerrors.Wrapf(
			types.ErrMessageSigner,
			"verification error %v",
			err,
		)
	}
	return nil
}
