package blsgroup

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/group"
)

type GroupKeeper interface {
	Proposal(goCtx context.Context, request *group.QueryProposalRequest) (*group.QueryProposalResponse, error)
	GroupInfo(goCtx context.Context, request *group.QueryGroupInfoRequest) (*group.QueryGroupInfoResponse, error)
	GroupPolicyInfo(goCtx context.Context, request *group.QueryGroupPolicyInfoRequest) (*group.QueryGroupPolicyInfoResponse, error)
	GroupMembers(goCtx context.Context, request *group.QueryGroupMembersRequest) (*group.QueryGroupMembersResponse, error)

	Vote(goCtx context.Context, req *group.MsgVote) (*group.MsgVoteResponse, error)
	Exec(goCtx context.Context, req *group.MsgExec) (*group.MsgExecResponse, error)
}

type AccountKeeper interface {
	GetAccount(sdk.Context, sdk.AccAddress) authtypes.AccountI
}
