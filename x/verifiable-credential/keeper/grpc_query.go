package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/fetchai/fetchd/x/verifiable-credential/types"
)

var _ types.QueryServer = Keeper{}

// VerifiableCredentials implements the VerifiableCredentials gRPC method
func (q Keeper) VerifiableCredentials(
	c context.Context,
	req *types.QueryVerifiableCredentialsRequest,
) (*types.QueryVerifiableCredentialsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	vcs := q.GetAllVerifiableCredentials(ctx)

	return &types.QueryVerifiableCredentialsResponse{
		Vcs: vcs,
	}, nil
}

// VerifiableCredential queries verifiable credentials info for given verifiable credentials id
func (q Keeper) VerifiableCredential(
	c context.Context,
	req *types.QueryVerifiableCredentialRequest,
) (*types.QueryVerifiableCredentialResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.VerifiableCredentialId == "" {
		return nil, status.Error(codes.InvalidArgument, "verifiable credential id cannot be empty")
	}

	ctx := sdk.UnwrapSDKContext(c)
	vc, found := q.GetVerifiableCredential(ctx, []byte(req.VerifiableCredentialId))
	if !found {
		return nil, status.Errorf(codes.NotFound, "vc %s not found", req.VerifiableCredentialId)
	}

	vcMeta, found := q.GetVcMetadata(ctx, []byte(req.VerifiableCredentialId))
	if !found {
		return nil, status.Errorf(codes.NotFound, "vc %s meta data not found", req.VerifiableCredentialId)
	}

	return &types.QueryVerifiableCredentialResponse{VerifiableCredential: vc, VcMetadata: vcMeta}, nil
}
