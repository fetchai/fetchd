package keeper

import (
	"context"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/group"
	grouperrors "github.com/cosmos/cosmos-sdk/x/group/errors"

	"github.com/fetchai/fetchd/crypto/keys/bls12381"
	"github.com/fetchai/fetchd/x/blsgroup"
)

var _ blsgroup.MsgServer = Keeper{}

func (k Keeper) VoteAgg(goCtx context.Context, req *blsgroup.MsgVoteAgg) (*blsgroup.MsgVoteAggResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	proposalResp, err := k.groupKeeper.Proposal(goCtx, &group.QueryProposalRequest{ProposalId: req.ProposalId})
	if err != nil {
		return nil, err
	}
	proposal := proposalResp.Proposal

	// Ensure that we can still accept votes for this proposal.
	if proposal.Status != group.PROPOSAL_STATUS_SUBMITTED {
		return nil, sdkerrors.Wrap(grouperrors.ErrInvalid, "proposal not open for voting")
	}
	if proposal.VotingPeriodEnd.Before(ctx.BlockTime()) {
		return nil, sdkerrors.Wrap(grouperrors.ErrExpired, "voting period has ended already")
	}

	policyInfo, err := k.groupKeeper.GroupPolicyInfo(goCtx, &group.QueryGroupPolicyInfoRequest{Address: proposal.Address})
	if err != nil {
		return nil, sdkerrors.Wrap(err, "load group policy")
	}

	electorate, err := k.groupKeeper.GroupInfo(goCtx, &group.QueryGroupInfoRequest{GroupId: policyInfo.Info.GroupId})
	if err != nil {
		return nil, err
	}

	groupMemberResp, err := k.groupKeeper.GroupMembers(goCtx, &group.QueryGroupMembersRequest{GroupId: electorate.Info.Id})
	if err != nil {
		return nil, err
	}
	members := groupMemberResp.Members

	// need the same number of votes than the group have members
	if g, w := len(req.Votes), len(members); g != w {
		return nil, sdkerrors.Wrapf(grouperrors.ErrInvalid, "got %d votes, want %d", g, w)
	}

	signedBytes := make([][]byte, 0, len(req.Votes))
	allVoteMsgs := make([]*group.MsgVote, 0, len(req.Votes))
	pks := make([]cryptotypes.PubKey, 0, len(req.Votes))

	for i, voteOption := range req.Votes {
		member := members[i].Member
		memAddr, err := sdk.AccAddressFromBech32(member.Address)
		if err != nil {
			return nil, err
		}

		acc := k.accKeeper.GetAccount(ctx, memAddr)
		if acc == nil {
			return nil, sdkerrors.Wrapf(grouperrors.ErrInvalid, "account %s does not exist", memAddr.String())
		}
		if !bls12381.IsPopValid(acc) {
			return nil, sdkerrors.Wrapf(grouperrors.ErrInvalid, "account %s have not proven possession of private key yet, make this account sign a transaction first", memAddr.String())
		}
		pk := acc.GetPubKey()
		if pk == nil {
			return nil, sdkerrors.Wrapf(grouperrors.ErrInvalid, "public key for account %s not set yet", memAddr.String())
		}

		if voteOption != group.VOTE_OPTION_UNSPECIFIED {
			msg := &group.MsgVote{
				ProposalId: req.ProposalId,
				Voter:      memAddr.String(),
				Option:     voteOption,
			}
			signedBytes = append(signedBytes, msg.GetSignBytes())
			allVoteMsgs = append(allVoteMsgs, msg)
			pks = append(pks, pk)
		}
	}

	// calculate and consume gas before the verification of the aggregated signature
	err = blsgroup.DefaultAggSigVerifyGasConsumer(ctx.GasMeter(), uint64(len(pks)), uint64(len(allVoteMsgs)), bls12381.DefaultSigVerifyCostBls12381)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "gas consumption for verifying aggregated signature")
	}

	if err = blsgroup.VerifyAggregateSignature(signedBytes, false, req.AggSig, pks); err != nil {
		return nil, err
	}

	for _, msg := range allVoteMsgs {
		msg.Exec = req.Exec
		if _, err := k.groupKeeper.Vote(goCtx, msg); err != nil {
			return nil, err
		}
	}

	return &blsgroup.MsgVoteAggResponse{}, nil
}
