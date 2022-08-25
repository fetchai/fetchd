package keeper

import (
	"context"

	didtypes "github.com/fetchai/fetchd/x/did/types"
	"github.com/fetchai/fetchd/x/verifiable-credential/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// IssueRegistrationCredential issuers a registration credential for a business
func (k msgServer) IssueRegistrationCredential(goCtx context.Context, msg *types.MsgIssueRegistrationCredential) (*types.MsgIssueRegistrationCredentialResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Info("issue registration request", "address", msg.Owner, "credential", msg.Credential)

	_, found := k.Keeper.GetVerifiableCredential(ctx, []byte(msg.Credential.Id))
	if found {
		err := sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "vc %s already exist", msg.Credential.Id)
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	if _, ok := msg.Credential.GetCredentialSubject().(*types.VerifiableCredential_RegistrationCred); !ok {
		err := sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "not a registration credential type")
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	// check if message signer is authorised by the did
	if err := k.didKeeper.VerifyDidWithRelationships(ctx, []string{didtypes.Authentication}, msg.Credential.Issuer, msg.Owner); err != nil {
		return nil, err
	}
	// store the credentials
	if vcErr := k.Keeper.SetVerifiableCredential(ctx, []byte(msg.Credential.Id), *msg.Credential); vcErr != nil {
		err := sdkerrors.Wrapf(vcErr, "credential proof cannot be verified")
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	// now create and persist the metadata
	vcM := types.NewVcMetadata(ctx.TxBytes(), ctx.BlockTime())
	k.Keeper.SetVcMetadata(ctx, []byte(msg.Credential.Id), vcM)

	k.Logger(ctx).Info("issue registration request successful", "did", msg.Credential.Issuer, "address", msg.Owner)

	ctx.EventManager().EmitEvent(
		types.NewCredentialCreatedEvent(msg.Owner, msg.Credential.Id),
	)

	return &types.MsgIssueRegistrationCredentialResponse{}, nil
}

// IssueUserCredential issues user credential
func (k msgServer) IssueUserCredential(
	goCtx context.Context,
	msg *types.MsgIssueUserCredential,
) (*types.MsgIssueUserCredentialResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Info("issue user credential request", "credential", msg.Credential, "address", msg.Owner)

	_, found := k.Keeper.GetVerifiableCredential(ctx, []byte(msg.Credential.Id))
	if found {
		err := sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "vc %s already exist", msg.Credential.Id)
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	if _, ok := msg.Credential.GetCredentialSubject().(*types.VerifiableCredential_UserCred); !ok {
		err := sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "not a user credential type")
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	// check if message signer is authorised by the did
	if err := k.didKeeper.VerifyDidWithRelationships(ctx, []string{didtypes.Authentication}, msg.Credential.Issuer, msg.Owner); err != nil {
		return nil, err
	}

	// store the credentials
	if vcErr := k.Keeper.SetVerifiableCredential(ctx, []byte(msg.Credential.Id), *msg.Credential); vcErr != nil {
		err := sdkerrors.Wrapf(vcErr, "credential proof cannot be verified")
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	// now create and persist the metadata
	vcM := types.NewVcMetadata(ctx.TxBytes(), ctx.BlockTime())
	k.Keeper.SetVcMetadata(ctx, []byte(msg.Credential.Id), vcM)

	k.Logger(ctx).Info("issue user credential request successful", "credentialID", msg.Credential.Id)

	ctx.EventManager().EmitEvent(
		types.NewCredentialCreatedEvent(msg.Owner, msg.Credential.Id),
	)

	return &types.MsgIssueUserCredentialResponse{}, nil
}

// IssueAnonymousCredentialSchema issues an anonymous credential schema
func (k msgServer) IssueAnonymousCredentialSchema(
	goCtx context.Context,
	msg *types.MsgIssueAnonymousCredentialSchema,
) (*types.MsgIssueAnonymousCredentialSchemaResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Info("issue anonymous credential schema request", "credential", msg.Credential, "address", msg.Owner)

	_, found := k.Keeper.GetVerifiableCredential(ctx, []byte(msg.Credential.Id))
	if found {
		err := sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "vc %s already exist", msg.Credential.Id)
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	if _, ok := msg.Credential.GetCredentialSubject().(*types.VerifiableCredential_AnonCredSchema); !ok {
		err := sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "not an anonymous credential schema type")
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	// check if message signer is authorised by the did
	if err := k.didKeeper.VerifyDidWithRelationships(ctx, []string{didtypes.Authentication}, msg.Credential.Issuer, msg.Owner); err != nil {
		return nil, err
	}

	// store the credentials
	if vcErr := k.Keeper.SetVerifiableCredential(ctx, []byte(msg.Credential.Id), *msg.Credential); vcErr != nil {
		err := sdkerrors.Wrapf(vcErr, "credential proof cannot be verified")
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	// now create and persist the metadata
	vcM := types.NewVcMetadata(ctx.TxBytes(), ctx.BlockTime())
	k.Keeper.SetVcMetadata(ctx, []byte(msg.Credential.Id), vcM)

	k.Logger(ctx).Info("issue anonymous credential schema request successful", "credentialID", msg.Credential.Id)

	ctx.EventManager().EmitEvent(
		types.NewCredentialCreatedEvent(msg.Owner, msg.Credential.Id),
	)

	return &types.MsgIssueAnonymousCredentialSchemaResponse{}, nil
}

// UpdateAccumulatorState update an existing anonymous credential schema
func (k msgServer) UpdateAccumulatorState(
	goCtx context.Context,
	msg *types.MsgUpdateAccumulatorState,
) (*types.MsgUpdateAccumulatorStateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Info("update anonymous credential schema request", "credential id", msg.CredentialId, "address", msg.Owner)

	vc, found := k.Keeper.GetVerifiableCredential(ctx, []byte(msg.CredentialId))
	if !found {
		err := sdkerrors.Wrapf(sdkerrors.ErrNotFound, "vc %s not found", msg.CredentialId)
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	_, found = k.Keeper.GetVcMetadata(ctx, []byte(msg.CredentialId))
	if !found {
		err := sdkerrors.Wrapf(sdkerrors.ErrNotFound, "vc %s meta data not found", msg.CredentialId)
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	// check if message signer is authorised by the did
	if didErr := k.didKeeper.VerifyDidWithRelationships(ctx, []string{didtypes.Authentication}, vc.Issuer, msg.Owner); didErr != nil {
		return nil, didErr
	}

	// update credential
	vc.IssuanceDate = msg.IssuanceDate
	vc, err := vc.UpdateAccumulatorState(msg.State)
	if err != nil {
		return nil, err
	}
	vc.Proof = msg.Proof

	// store the credentials
	if vcErr := k.Keeper.SetVerifiableCredential(ctx, []byte(msg.CredentialId), vc); vcErr != nil {
		err := sdkerrors.Wrapf(vcErr, "credential proof cannot be verified")
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	// update the Metadata
	if metaErr := updateVcMetadata(&k.Keeper, ctx, msg.CredentialId, false); metaErr != nil {
		err := sdkerrors.Wrapf(metaErr, "vc %s", msg.CredentialId)
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	k.Logger(ctx).Info("update accumulator state request successful", "credentialID", msg.CredentialId)

	return &types.MsgUpdateAccumulatorStateResponse{}, nil
}

// UpdateAnonymousCredentialSchema update an existing anonymous credential schema
func (k msgServer) UpdateVerifiableCredential(
	goCtx context.Context,
	msg *types.MsgUpdateVerifiableCredential,
) (*types.MsgUpdateVerifiableCredentialResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k.Logger(ctx).Info("update verifiable credential request", "credential id", msg.Credential.Id, "address", msg.Owner)

	vc, found := k.Keeper.GetVerifiableCredential(ctx, []byte(msg.Credential.Id))
	if !found {
		err := sdkerrors.Wrapf(sdkerrors.ErrNotFound, "vc %s not found", msg.Credential.Id)
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	_, found = k.Keeper.GetVcMetadata(ctx, []byte(msg.Credential.Id))
	if !found {
		err := sdkerrors.Wrapf(sdkerrors.ErrNotFound, "vc %s meta data not found", msg.Credential.Id)
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	/*
		if vcMeta.Deactivated {
			err := sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "vc %s already deactived", msg.Credential.Id)
			k.Logger(ctx).Error(err.Error())
			return nil, err
		}
	*/

	// check if message signer is authorised by the did
	if didErr := k.didKeeper.VerifyDidWithRelationships(ctx, []string{didtypes.Authentication}, msg.Credential.Issuer, msg.Owner); didErr != nil {
		return nil, didErr
	}

	// issuer address of old and new vc are not necessarily the same
	if oldDidErr := k.didKeeper.VerifyDidWithRelationships(ctx, []string{didtypes.Authentication}, vc.Issuer, msg.Owner); oldDidErr != nil {
		return nil, oldDidErr
	}

	// store the credentials
	if vcErr := k.Keeper.SetVerifiableCredential(ctx, []byte(msg.Credential.Id), *msg.Credential); vcErr != nil {
		err := sdkerrors.Wrapf(vcErr, "credential proof cannot be verified")
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	// update the Metadata
	if metaErr := updateVcMetadata(&k.Keeper, ctx, msg.Credential.Id, false); metaErr != nil {
		err := sdkerrors.Wrapf(metaErr, "vc %s", msg.Credential.Id)
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	k.Logger(ctx).Info("update verifiable credential request successful", "credentialID", msg.Credential.Id)

	return &types.MsgUpdateVerifiableCredentialResponse{}, nil
}

// RevokeCredential revoke a credential
func (k msgServer) RevokeCredential(goCtx context.Context, msg *types.MsgRevokeCredential) (*types.MsgRevokeCredentialResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	k.Logger(ctx).Info("revoke credential request", "credential", msg.CredentialId, "address", msg.Owner)

	vc, found := k.GetVerifiableCredential(ctx, []byte(msg.CredentialId))
	if !found {
		err := sdkerrors.Wrapf(sdkerrors.ErrNotFound, "vc %s not found", msg.CredentialId)
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	// check if message signer is authorised by the did
	if didErr := k.didKeeper.VerifyDidWithRelationships(ctx, []string{didtypes.Authentication}, vc.Issuer, msg.Owner); didErr != nil {
		return nil, didErr
	}

	// for now revoking credential is done by marking it as deactived in meta data
	// update the Metadata
	if metaErr := updateVcMetadata(&k.Keeper, ctx, msg.CredentialId, true); metaErr != nil {
		err := sdkerrors.Wrapf(metaErr, "vc %s", msg.CredentialId)
		k.Logger(ctx).Error(err.Error())
		return nil, err
	}

	k.Logger(ctx).Info("revoke license request successful", "credential", msg.CredentialId, "address", msg.Owner)

	ctx.EventManager().EmitEvent(
		types.NewCredentialDeletedEvent(msg.Owner, msg.CredentialId),
	)

	return &types.MsgRevokeCredentialResponse{}, nil
}

// helper function to update the vc metadata
func updateVcMetadata(keeper *Keeper, ctx sdk.Context, vcId string, deactived bool) (err error) {
	vcMeta, found := keeper.GetVcMetadata(ctx, []byte(vcId))
	if found {
		types.UpdateVcMetadata(&vcMeta, ctx.TxBytes(), ctx.BlockTime(), deactived)
		keeper.SetVcMetadata(ctx, []byte(vcId), vcMeta)
	} else {
		err = sdkerrors.Wrapf(sdkerrors.ErrNotFound, "(warning) vc metadata not found")
	}
	return
}
