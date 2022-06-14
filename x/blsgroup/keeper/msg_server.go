package keeper

import (
	"context"
	"encoding/binary"
	"errors"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/group"
	grouperrors "github.com/cosmos/cosmos-sdk/x/group/errors"

	"github.com/fetchai/fetchd/crypto/keys/bls12381"
	"github.com/fetchai/fetchd/x/blsgroup"
)

var _ blsgroup.MsgServer = Keeper{}

// RegisterBlsGroup checks that all group members have a bls key, and they proven possession of the corresponding private key.
// It then register the groupID and its current version as a bls group to enable the other BLS feature (such as VoteAgg).
// If the group is modified, it's version change must the group must be registered again.
func (k Keeper) RegisterBlsGroup(goCtx context.Context, req *blsgroup.MsgRegisterBlsGroup) (*blsgroup.MsgRegisterBlsGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	groupInfoResp, err := k.groupKeeper.GroupInfo(goCtx, &group.QueryGroupInfoRequest{GroupId: req.GroupId})
	if err != nil {
		return nil, sdkerrors.Wrap(err, "load group")
	}

	// Only current group admin is authorized to register a group
	groupAdmin, err := sdk.AccAddressFromBech32(groupInfoResp.Info.Admin)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "group admin")
	}
	reqAdmin, err := sdk.AccAddressFromBech32(req.Admin)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "request admin")
	}
	if !groupAdmin.Equals(reqAdmin) {
		return nil, sdkerrors.Wrap(grouperrors.ErrUnauthorized, "not group admin")
	}

	if k.isRegisteredBlsGroup(ctx, groupInfoResp.Info) {
		return nil, sdkerrors.Wrap(grouperrors.ErrDuplicate, "already registered")
	}

	groupMemberResp, err := k.groupKeeper.GroupMembers(goCtx, &group.QueryGroupMembersRequest{GroupId: req.GroupId})
	if err != nil {
		return nil, sdkerrors.Wrap(err, "load group members")
	}

	for _, member := range groupMemberResp.Members {
		memAddr, err := sdk.AccAddressFromBech32(member.Member.Address)
		if err != nil {
			return nil, err
		}
		acc := k.accKeeper.GetAccount(ctx, memAddr)
		if acc == nil {
			return nil, sdkerrors.Wrapf(grouperrors.ErrInvalid, "account %s does not exist", memAddr.String())
		}
		pk := acc.GetPubKey()
		if pk == nil {
			return nil, sdkerrors.Wrapf(grouperrors.ErrInvalid, "public key for account %s not set yet", memAddr.String())
		}

		if _, ok := pk.(*bls12381.PubKey); !ok {
			return nil, sdkerrors.Wrapf(grouperrors.ErrInvalid, "public key for account %s is not a BLS key", memAddr.String())
		}
		if !bls12381.IsPopValid(acc) {
			return nil, sdkerrors.Wrapf(grouperrors.ErrInvalid, "account %s have not proven possession of private key yet, make this account sign a transaction first", memAddr.String())
		}
	}

	groupStore := prefix.NewStore(ctx.KVStore(k.key), blsgroup.GroupPrefixKey)
	groupStore.Set(encodeGroupID(req.GroupId), encodeGroupVersion(groupInfoResp.Info.Version))

	return &blsgroup.MsgRegisterBlsGroupResponse{}, nil
}

// UnregisterBlsGroup delete a BLS group registration
func (k Keeper) UnregisterBlsGroup(goCtx context.Context, req *blsgroup.MsgUnregisterBlsGroup) (*blsgroup.MsgUnregisterBlsGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	groupInfoResp, err := k.groupKeeper.GroupInfo(goCtx, &group.QueryGroupInfoRequest{GroupId: req.GroupId})
	if err != nil {
		return nil, sdkerrors.Wrap(err, "load group")
	}

	// Only current group admin is authorized to unregister a group
	groupAdmin, err := sdk.AccAddressFromBech32(groupInfoResp.Info.Admin)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "group admin")
	}
	reqAdmin, err := sdk.AccAddressFromBech32(req.Admin)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "request admin")
	}
	if !groupAdmin.Equals(reqAdmin) {
		return nil, sdkerrors.Wrap(grouperrors.ErrUnauthorized, "not group admin")
	}

	if !k.isRegisteredBlsGroup(ctx, groupInfoResp.Info) {
		return nil, sdkerrors.Wrap(grouperrors.ErrInvalid, "bls group not registered")
	}

	groupStore := prefix.NewStore(ctx.KVStore(k.key), blsgroup.GroupPrefixKey)
	groupStore.Delete(encodeGroupID(req.GroupId))

	return &blsgroup.MsgUnregisterBlsGroupResponse{}, nil
}

func (k Keeper) isRegisteredBlsGroup(ctx sdk.Context, group *group.GroupInfo) bool {
	groupStore := prefix.NewStore(ctx.KVStore(k.key), blsgroup.GroupPrefixKey)
	encodedVersion := groupStore.Get(encodeGroupID(group.Id))
	if encodedVersion == nil {
		return false
	}

	if group.Version != decodeGroupVersion(encodedVersion) {
		return false
	}

	return true
}

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

	policyInfo, err := k.groupKeeper.GroupPolicyInfo(goCtx, &group.QueryGroupPolicyInfoRequest{Address: proposal.GroupPolicyAddress})
	if err != nil {
		return nil, sdkerrors.Wrap(err, "load group policy")
	}

	groupInfo, err := k.groupKeeper.GroupInfo(goCtx, &group.QueryGroupInfoRequest{GroupId: policyInfo.Info.GroupId})
	if err != nil {
		return nil, sdkerrors.Wrap(err, "load group")
	}
	if !k.isRegisteredBlsGroup(ctx, groupInfo.Info) {
		return nil, sdkerrors.Wrap(grouperrors.ErrInvalid, "bls group not registered")
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
			// Duplicate votes are simply skipped rather than failing the whole process
			if errors.Is(err, grouperrors.ErrORMUniqueConstraint) {
				continue
			}
			return nil, err
		}
	}

	return &blsgroup.MsgVoteAggResponse{}, nil
}

func encodeGroupID(groupID uint64) []byte {
	groupIDBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(groupIDBytes, groupID)

	return groupIDBytes
}

func encodeGroupVersion(version uint64) []byte {
	versionBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(versionBytes, version)

	return versionBytes
}

func decodeGroupVersion(versionBytes []byte) uint64 {
	return binary.BigEndian.Uint64(versionBytes)
}
