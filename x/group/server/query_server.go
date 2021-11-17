package server

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/fetchai/fetchd/orm"
	"github.com/fetchai/fetchd/types"
	"github.com/fetchai/fetchd/x/group"
)

func (s serverImpl) GroupInfo(goCtx context.Context, request *group.QueryGroupInfoRequest) (*group.QueryGroupInfoResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	groupID := request.GroupId
	groupInfo, err := s.getGroupInfo(ctx, groupID)
	if err != nil {
		return nil, err
	}

	return &group.QueryGroupInfoResponse{Info: &groupInfo}, nil
}

func (s serverImpl) getGroupInfo(goCtx context.Context, id uint64) (group.GroupInfo, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	var obj group.GroupInfo
	_, err := s.groupTable.GetOne(ctx, id, &obj)
	return obj, err
}

func (s serverImpl) GroupAccountInfo(goCtx context.Context, request *group.QueryGroupAccountInfoRequest) (*group.QueryGroupAccountInfoResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Address)
	if err != nil {
		return nil, err
	}
	groupAccountInfo, err := s.getGroupAccountInfo(ctx, addr)
	if err != nil {
		return nil, err
	}

	return &group.QueryGroupAccountInfoResponse{Info: &groupAccountInfo}, nil
}

func (s serverImpl) getGroupAccountInfo(goCtx context.Context, accountAddress sdk.AccAddress) (group.GroupAccountInfo, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	var obj group.GroupAccountInfo
	return obj, s.groupAccountTable.GetOne(ctx, orm.AddLengthPrefix(accountAddress.Bytes()), &obj)
}

func (s serverImpl) GroupMembers(goCtx context.Context, request *group.QueryGroupMembersRequest) (*group.QueryGroupMembersResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	groupID := request.GroupId
	it, err := s.getGroupMembers(ctx, groupID, request.Pagination)
	if err != nil {
		return nil, err
	}

	var members []*group.GroupMember
	pageRes, err := orm.Paginate(it, request.Pagination, &members)
	if err != nil {
		return nil, err
	}

	return &group.QueryGroupMembersResponse{
		Members:    members,
		Pagination: pageRes,
	}, nil
}

func (s serverImpl) getGroupMembers(goCtx context.Context, id uint64, pageRequest *query.PageRequest) (orm.Iterator, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	return s.groupMemberByGroupIndex.GetPaginated(ctx, id, pageRequest)
}

func (s serverImpl) GroupsByAdmin(goCtx context.Context, request *group.QueryGroupsByAdminRequest) (*group.QueryGroupsByAdminResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Admin)
	if err != nil {
		return nil, err
	}
	it, err := s.getGroupsByAdmin(ctx, addr, request.Pagination)
	if err != nil {
		return nil, err
	}

	var groups []*group.GroupInfo
	pageRes, err := orm.Paginate(it, request.Pagination, &groups)
	if err != nil {
		return nil, err
	}

	return &group.QueryGroupsByAdminResponse{
		Groups:     groups,
		Pagination: pageRes,
	}, nil
}

func (s serverImpl) getGroupsByAdmin(goCtx context.Context, admin sdk.AccAddress, pageRequest *query.PageRequest) (orm.Iterator, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	return s.groupByAdminIndex.GetPaginated(ctx, admin.Bytes(), pageRequest)
}

func (s serverImpl) GroupAccountsByGroup(goCtx context.Context, request *group.QueryGroupAccountsByGroupRequest) (*group.QueryGroupAccountsByGroupResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	groupID := request.GroupId
	it, err := s.getGroupAccountsByGroup(ctx, groupID, request.Pagination)
	if err != nil {
		return nil, err
	}

	var accounts []*group.GroupAccountInfo
	pageRes, err := orm.Paginate(it, request.Pagination, &accounts)
	if err != nil {
		return nil, err
	}

	return &group.QueryGroupAccountsByGroupResponse{
		GroupAccounts: accounts,
		Pagination:    pageRes,
	}, nil
}

func (s serverImpl) getGroupAccountsByGroup(ctx types.Context, id uint64, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return s.groupAccountByGroupIndex.GetPaginated(ctx, id, pageRequest)
}

func (s serverImpl) GroupAccountsByAdmin(goCtx context.Context, request *group.QueryGroupAccountsByAdminRequest) (*group.QueryGroupAccountsByAdminResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Admin)
	if err != nil {
		return nil, err
	}
	it, err := s.getGroupAccountsByAdmin(ctx, addr, request.Pagination)
	if err != nil {
		return nil, err
	}

	var accounts []*group.GroupAccountInfo
	pageRes, err := orm.Paginate(it, request.Pagination, &accounts)
	if err != nil {
		return nil, err
	}

	return &group.QueryGroupAccountsByAdminResponse{
		GroupAccounts: accounts,
		Pagination:    pageRes,
	}, nil
}

func (s serverImpl) getGroupAccountsByAdmin(ctx types.Context, admin sdk.AccAddress, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return s.groupAccountByAdminIndex.GetPaginated(ctx, admin.Bytes(), pageRequest)
}

func (s serverImpl) Proposal(goCtx context.Context, request *group.QueryProposalRequest) (*group.QueryProposalResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	proposalID := request.ProposalId
	proposal, err := s.getProposal(ctx, proposalID)
	if err != nil {
		return nil, err
	}

	return &group.QueryProposalResponse{Proposal: &proposal}, nil
}

func (s serverImpl) ProposalsByGroupAccount(goCtx context.Context, request *group.QueryProposalsByGroupAccountRequest) (*group.QueryProposalsByGroupAccountResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Address)
	if err != nil {
		return nil, err
	}
	it, err := s.getProposalsByGroupAccount(ctx, addr, request.Pagination)
	if err != nil {
		return nil, err
	}

	var proposals []*group.Proposal
	pageRes, err := orm.Paginate(it, request.Pagination, &proposals)
	if err != nil {
		return nil, err
	}

	return &group.QueryProposalsByGroupAccountResponse{
		Proposals:  proposals,
		Pagination: pageRes,
	}, nil
}

func (s serverImpl) getProposalsByGroupAccount(ctx types.Context, account sdk.AccAddress, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return s.proposalByGroupAccountIndex.GetPaginated(ctx, account.Bytes(), pageRequest)
}

func (s serverImpl) getProposal(ctx types.Context, proposalID uint64) (group.Proposal, error) {
	var p group.Proposal
	if _, err := s.proposalTable.GetOne(ctx, proposalID, &p); err != nil {
		return group.Proposal{}, sdkerrors.Wrap(err, "load proposal")
	}
	return p, nil
}

func (s serverImpl) VoteByProposalVoter(goCtx context.Context, request *group.QueryVoteByProposalVoterRequest) (*group.QueryVoteByProposalVoterResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Voter)
	if err != nil {
		return nil, err
	}
	proposalID := request.ProposalId
	vote, err := s.getVote(ctx, proposalID, addr)
	if err != nil {
		return nil, err
	}
	return &group.QueryVoteByProposalVoterResponse{
		Vote: &vote,
	}, nil
}

func (s serverImpl) VotesByProposal(goCtx context.Context, request *group.QueryVotesByProposalRequest) (*group.QueryVotesByProposalResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	proposalID := request.ProposalId
	it, err := s.getVotesByProposal(ctx, proposalID, request.Pagination)
	if err != nil {
		return nil, err
	}

	var votes []*group.Vote
	pageRes, err := orm.Paginate(it, request.Pagination, &votes)
	if err != nil {
		return nil, err
	}

	return &group.QueryVotesByProposalResponse{
		Votes:      votes,
		Pagination: pageRes,
	}, nil
}

func (s serverImpl) VotesByVoter(goCtx context.Context, request *group.QueryVotesByVoterRequest) (*group.QueryVotesByVoterResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Voter)
	if err != nil {
		return nil, err
	}
	it, err := s.getVotesByVoter(ctx, addr, request.Pagination)
	if err != nil {
		return nil, err
	}

	var votes []*group.Vote
	pageRes, err := orm.Paginate(it, request.Pagination, &votes)
	if err != nil {
		return nil, err
	}

	return &group.QueryVotesByVoterResponse{
		Votes:      votes,
		Pagination: pageRes,
	}, nil
}

func (s serverImpl) getVote(ctx types.Context, proposalID uint64, voter sdk.AccAddress) (group.Vote, error) {
	var v group.Vote
	return v, s.voteTable.GetOne(ctx, orm.PrimaryKey(&group.Vote{ProposalId: proposalID, Voter: voter.String()}), &v)
}

func (s serverImpl) getVotesByProposal(ctx types.Context, proposalID uint64, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return s.voteByProposalIndex.GetPaginated(ctx, proposalID, pageRequest)
}

func (s serverImpl) getVotesByVoter(ctx types.Context, voter sdk.AccAddress, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return s.voteByVoterIndex.GetPaginated(ctx, voter.Bytes(), pageRequest)
}

func (s serverImpl) Poll(goCtx context.Context, request *group.QueryPollRequest) (*group.QueryPollResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	pollID := request.PollId
	poll, err := s.getPoll(ctx, pollID)
	if err != nil {
		return nil, err
	}

	return &group.QueryPollResponse{Poll: &poll}, nil
}

func (s serverImpl) getPoll(ctx types.Context, pollID uint64) (group.Poll, error) {
	var p group.Poll
	if _, err := s.pollTable.GetOne(ctx, pollID, &p); err != nil {
		return group.Poll{}, sdkerrors.Wrap(err, "load poll")
	}
	return p, nil
}

func (s serverImpl) PollsByGroup(goCtx context.Context, request *group.QueryPollsByGroupRequest) (*group.QueryPollsByGroupResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	groupID := request.GroupId
	it, err := s.getPollsByGroup(ctx, groupID, request.Pagination)
	if err != nil {
		return nil, err
	}

	var polls []*group.Poll
	pageRes, err := orm.Paginate(it, request.Pagination, &polls)
	if err != nil {
		return nil, err
	}

	return &group.QueryPollsByGroupResponse{
		Polls:      polls,
		Pagination: pageRes,
	}, nil
}

func (s serverImpl) getPollsByGroup(ctx types.Context, id uint64, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return s.pollByGroupIndex.GetPaginated(ctx, id, pageRequest)
}

func (s serverImpl) PollsByCreator(goCtx context.Context, request *group.QueryPollsByCreatorRequest) (*group.QueryPollsByCreatorResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Creator)
	if err != nil {
		return nil, err
	}
	it, err := s.getPollsByCreator(ctx, addr, request.Pagination)
	if err != nil {
		return nil, err
	}

	var polls []*group.Poll
	pageRes, err := orm.Paginate(it, request.Pagination, &polls)
	if err != nil {
		return nil, err
	}

	return &group.QueryPollsByCreatorResponse{
		Polls:      polls,
		Pagination: pageRes,
	}, nil
}

func (s serverImpl) getPollsByCreator(ctx types.Context, account sdk.AccAddress, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return s.pollByCreatorIndex.GetPaginated(ctx, account.Bytes(), pageRequest)
}

func (s serverImpl) VoteForPollByPollVoter(goCtx context.Context, request *group.QueryVoteForPollByPollVoterRequest) (*group.QueryVoteForPollByPollVoterResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Voter)
	if err != nil {
		return nil, err
	}
	pollID := request.PollId
	vote, err := s.getVoteForPoll(ctx, pollID, addr)
	if err != nil {
		return nil, err
	}
	return &group.QueryVoteForPollByPollVoterResponse{
		Vote: &vote,
	}, nil
}

func (s serverImpl) VotesForPollByPoll(goCtx context.Context, request *group.QueryVotesForPollByPollRequest) (*group.QueryVotesForPollByPollResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	pollID := request.PollId
	it, err := s.getVotesForPollByPoll(ctx, pollID, request.Pagination)
	if err != nil {
		return nil, err
	}

	var votes []*group.VotePoll
	pageRes, err := orm.Paginate(it, request.Pagination, &votes)
	if err != nil {
		return nil, err
	}

	return &group.QueryVotesForPollByPollResponse{
		Votes:      votes,
		Pagination: pageRes,
	}, nil
}

func (s serverImpl) VotesForPollByVoter(goCtx context.Context, request *group.QueryVotesForPollByVoterRequest) (*group.QueryVotesForPollByVoterResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(request.Voter)
	if err != nil {
		return nil, err
	}
	it, err := s.getVotesForPollByVoter(ctx, addr, request.Pagination)
	if err != nil {
		return nil, err
	}

	var votes []*group.VotePoll
	pageRes, err := orm.Paginate(it, request.Pagination, &votes)
	if err != nil {
		return nil, err
	}

	return &group.QueryVotesForPollByVoterResponse{
		Votes:      votes,
		Pagination: pageRes,
	}, nil
}

func (s serverImpl) getVoteForPoll(ctx types.Context, pollID uint64, voter sdk.AccAddress) (group.VotePoll, error) {
	var v group.VotePoll
	return v, s.votePollTable.GetOne(ctx, orm.PrimaryKey(&group.VotePoll{PollId: pollID, Voter: voter.String()}), &v)
}

func (s serverImpl) getVotesForPollByPoll(ctx types.Context, pollID uint64, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return s.votePollByPollIndex.GetPaginated(ctx, pollID, pageRequest)
}

func (s serverImpl) getVotesForPollByVoter(ctx types.Context, voter sdk.AccAddress, pageRequest *query.PageRequest) (orm.Iterator, error) {
	return s.votePollByVoterIndex.GetPaginated(ctx, voter.Bytes(), pageRequest)
}