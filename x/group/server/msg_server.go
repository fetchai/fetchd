package server

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"reflect"
	"sort"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/bls12381"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gogotypes "github.com/gogo/protobuf/types"

	"github.com/fetchai/fetchd/orm"
	"github.com/fetchai/fetchd/types"
	"github.com/fetchai/fetchd/types/math"
	"github.com/fetchai/fetchd/x/group"
)

// TODO: Revisit this once we have propoer gas fee framework.
// Tracking issues https://github.com/cosmos/cosmos-sdk/issues/9054, https://github.com/cosmos/cosmos-sdk/discussions/9072
const gasCostPerIteration = uint64(20)

func (s serverImpl) validateBlsMember(ctx types.Context, mem group.Member) error {
	addr, err := sdk.AccAddressFromBech32(mem.Address)
	if err != nil {
		return err
	}
	acc := s.accKeeper.GetAccount(ctx.Context, addr)
	if acc == nil {
		return fmt.Errorf("account %s does not exist", mem.Address)
	}

	pk := acc.GetPubKey()
	if pk == nil {
		return fmt.Errorf("account public key not set yet")
	}

	if _, ok := pk.(*bls12381.PubKey); !ok {
		return fmt.Errorf("member account %s is not a bls account", mem.Address)
	}

	if !acc.GetPopValid() {
		return fmt.Errorf("member account %s hasn't validated pop for public key", mem.Address)
	}

	return nil
}

func (s serverImpl) CreateGroup(goCtx context.Context, req *group.MsgCreateGroup) (*group.MsgCreateGroupResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	metadata := req.Metadata
	members := group.Members{Members: req.Members}
	admin := req.Admin
	bls := req.BlsOnly

	if err := members.ValidateBasic(); err != nil {
		return nil, err
	}
	if bls {
		for _, mem := range members.Members {
			if err := s.validateBlsMember(ctx, mem); err != nil {
				return nil, sdkerrors.Wrapf(group.ErrBlsRequired, "member %s failed bls validation: %v", mem.Address, err)
			}
		}
	}

	if len(metadata) > group.MaxMetadataLength {
		return nil, sdkerrors.Wrap(group.ErrMaxLimit, "group metadata")
	}

	totalWeight := math.NewDecFromInt64(0)
	for i := range members.Members {
		m := members.Members[i]

		// Members of a group must have a positive weight.
		weight, err := math.NewPositiveDecFromString(m.Weight)
		if err != nil {
			return nil, err
		}

		// Adding up members weights to compute group total weight.
		totalWeight, err = totalWeight.Add(weight)
		if err != nil {
			return nil, err
		}
	}

	// Create a new group in the groupTable.
	groupInfo := &group.GroupInfo{
		GroupId:     s.groupTable.Sequence().PeekNextVal(ctx),
		Admin:       admin,
		Metadata:    metadata,
		Version:     1,
		TotalWeight: totalWeight.String(),
		BlsOnly:     bls,
	}
	groupID, err := s.groupTable.Create(ctx, groupInfo)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "could not create group")
	}

	// Create new group members in the groupMemberTable.
	for i := range members.Members {
		m := members.Members[i]
		err := s.groupMemberTable.Create(ctx, &group.GroupMember{
			GroupId: groupID,
			Member: &group.Member{
				Address:  m.Address,
				Weight:   m.Weight,
				Metadata: m.Metadata,
			},
		})
		if err != nil {
			return nil, sdkerrors.Wrapf(err, "could not store member %d", i)
		}
	}

	err = ctx.EventManager().EmitTypedEvent(&group.EventCreateGroup{GroupId: groupID})
	if err != nil {
		return nil, err
	}

	return &group.MsgCreateGroupResponse{GroupId: groupID}, nil
}

func (s serverImpl) UpdateGroupMembers(goCtx context.Context, req *group.MsgUpdateGroupMembers) (*group.MsgUpdateGroupMembersResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	action := func(g *group.GroupInfo) error {
		totalWeight, err := math.NewNonNegativeDecFromString(g.TotalWeight)
		if err != nil {
			return err
		}

		if g.BlsOnly {
			for _, mem := range req.MemberUpdates {
				if err := s.validateBlsMember(ctx, mem); err != nil {
					return sdkerrors.Wrapf(group.ErrBlsRequired, "member %s failed bls validation: %v", mem.Address, err)
				}
			}
		}

		for i := range req.MemberUpdates {
			groupMember := group.GroupMember{GroupId: req.GroupId,
				Member: &group.Member{
					Address:  req.MemberUpdates[i].Address,
					Weight:   req.MemberUpdates[i].Weight,
					Metadata: req.MemberUpdates[i].Metadata,
				},
			}

			// Checking if the group member is already part of the group.
			var found bool
			var prevGroupMember group.GroupMember
			switch err := s.groupMemberTable.GetOne(ctx, orm.PrimaryKey(&groupMember), &prevGroupMember); {
			case err == nil:
				found = true
			case orm.ErrNotFound.Is(err):
				found = false
			default:
				return sdkerrors.Wrap(err, "get group member")
			}

			newMemberWeight, err := math.NewNonNegativeDecFromString(groupMember.Member.Weight)
			if err != nil {
				return err
			}

			// Handle delete for members with zero weight.
			if newMemberWeight.IsZero() {
				// We can't delete a group member that doesn't already exist.
				if !found {
					return sdkerrors.Wrap(orm.ErrNotFound, "unknown member")
				}

				previousMemberWeight, err := math.NewNonNegativeDecFromString(prevGroupMember.Member.Weight)
				if err != nil {
					return err
				}

				// Subtract the weight of the group member to delete from the group total weight.
				totalWeight, err = math.SubNonNegative(totalWeight, previousMemberWeight)
				if err != nil {
					return err
				}

				// Delete group member in the groupMemberTable.
				if err := s.groupMemberTable.Delete(ctx, &groupMember); err != nil {
					return sdkerrors.Wrap(err, "delete member")
				}
				continue
			}
			// If group member already exists, handle update
			if found {
				previousMemberWeight, err := math.NewNonNegativeDecFromString(prevGroupMember.Member.Weight)
				if err != nil {
					return err
				}
				// Subtract previous weight from the group total weight.
				totalWeight, err = math.SubNonNegative(totalWeight, previousMemberWeight)
				if err != nil {
					return err
				}
				// Update updated group member in the groupMemberTable.
				if err := s.groupMemberTable.Update(ctx, &groupMember); err != nil {
					return sdkerrors.Wrap(err, "add member")
				}
				// else handle create.
			} else if err := s.groupMemberTable.Create(ctx, &groupMember); err != nil {
				return sdkerrors.Wrap(err, "add member")
			}
			// In both cases (handle + update), we need to add the new member's weight to the group total weight.
			totalWeight, err = totalWeight.Add(newMemberWeight)
			if err != nil {
				return err
			}
		}
		// Update group in the groupTable.
		g.TotalWeight = totalWeight.String()
		g.Version++
		return s.groupTable.Update(ctx, g.GroupId, g)
	}

	err := s.doUpdateGroup(ctx, req, action, "members updated")
	if err != nil {
		return nil, err
	}

	return &group.MsgUpdateGroupMembersResponse{}, nil
}

func (s serverImpl) UpdateGroupAdmin(goCtx context.Context, req *group.MsgUpdateGroupAdmin) (*group.MsgUpdateGroupAdminResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	action := func(g *group.GroupInfo) error {
		g.Admin = req.NewAdmin
		g.Version++

		return s.groupTable.Update(ctx, g.GroupId, g)
	}

	err := s.doUpdateGroup(ctx, req, action, "admin updated")
	if err != nil {
		return nil, err
	}

	return &group.MsgUpdateGroupAdminResponse{}, nil
}

func (s serverImpl) UpdateGroupMetadata(goCtx context.Context, req *group.MsgUpdateGroupMetadata) (*group.MsgUpdateGroupMetadataResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	action := func(g *group.GroupInfo) error {
		g.Metadata = req.Metadata
		g.Version++
		return s.groupTable.Update(ctx, g.GroupId, g)
	}

	err := s.doUpdateGroup(ctx, req, action, "metadata updated")
	if err != nil {
		return nil, err
	}

	return &group.MsgUpdateGroupMetadataResponse{}, nil
}

func (s serverImpl) CreateGroupAccount(goCtx context.Context, req *group.MsgCreateGroupAccount) (*group.MsgCreateGroupAccountResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	admin, err := sdk.AccAddressFromBech32(req.GetAdmin())
	if err != nil {
		return nil, sdkerrors.Wrap(err, "request admin")
	}
	policy := req.GetDecisionPolicy()
	groupID := req.GetGroupID()
	metadata := req.GetMetadata()

	g, err := s.getGroupInfo(ctx, groupID)
	if err != nil {
		return nil, err
	}
	groupAdmin, err := sdk.AccAddressFromBech32(g.Admin)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "group admin")
	}
	// Only current group admin is authorized to create a group account for this group.
	if !groupAdmin.Equals(admin) {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "not group admin")
	}

	// Generate group account address.
	var accountAddr sdk.AccAddress
	var accountDerivationKey []byte
	// loop here in the rare case of a collision
	for {
		nextAccVal := s.groupAccountSeq.NextVal(ctx)
		buf := bytes.NewBuffer(nil)
		err = binary.Write(buf, binary.LittleEndian, nextAccVal)
		if err != nil {
			return nil, err
		}

		accountDerivationKey = buf.Bytes()
		accountID := s.key.Derive(accountDerivationKey)
		accountAddr = accountID.Address()

		if s.accKeeper.GetAccount(ctx.Context, accountAddr) != nil {
			// handle a rare collision
			continue
		}

		acc := s.accKeeper.NewAccount(ctx.Context, &authtypes.ModuleAccount{
			BaseAccount: &authtypes.BaseAccount{
				Address: accountAddr.String(),
			},
			Name: accountAddr.String(),
		})
		s.accKeeper.SetAccount(ctx.Context, acc)

		break
	}

	groupAccount, err := group.NewGroupAccountInfo(
		accountAddr,
		groupID,
		admin,
		metadata,
		1,
		policy,
		accountDerivationKey,
	)
	if err != nil {
		return nil, err
	}

	if err := s.groupAccountTable.Create(ctx, &groupAccount); err != nil {
		return nil, sdkerrors.Wrap(err, "could not create group account")
	}

	err = ctx.EventManager().EmitTypedEvent(&group.EventCreateGroupAccount{Address: accountAddr.String()})
	if err != nil {
		return nil, err
	}

	return &group.MsgCreateGroupAccountResponse{Address: accountAddr.String()}, nil
}

func (s serverImpl) UpdateGroupAccountAdmin(goCtx context.Context, req *group.MsgUpdateGroupAccountAdmin) (*group.MsgUpdateGroupAccountAdminResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	action := func(groupAccount *group.GroupAccountInfo) error {
		groupAccount.Admin = req.NewAdmin
		groupAccount.Version++
		return s.groupAccountTable.Update(ctx, groupAccount)
	}

	err := s.doUpdateGroupAccount(ctx, req.Address, req.Admin, action, "group account admin updated")
	if err != nil {
		return nil, err
	}

	return &group.MsgUpdateGroupAccountAdminResponse{}, nil
}

func (s serverImpl) UpdateGroupAccountDecisionPolicy(goCtx context.Context, req *group.MsgUpdateGroupAccountDecisionPolicy) (*group.MsgUpdateGroupAccountDecisionPolicyResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	policy := req.GetDecisionPolicy()

	action := func(groupAccount *group.GroupAccountInfo) error {
		err := groupAccount.SetDecisionPolicy(policy)
		if err != nil {
			return err
		}

		groupAccount.Version++
		return s.groupAccountTable.Update(ctx, groupAccount)
	}

	err := s.doUpdateGroupAccount(ctx, req.Address, req.Admin, action, "group account decision policy updated")
	if err != nil {
		return nil, err
	}

	return &group.MsgUpdateGroupAccountDecisionPolicyResponse{}, nil
}

func (s serverImpl) UpdateGroupAccountMetadata(goCtx context.Context, req *group.MsgUpdateGroupAccountMetadata) (*group.MsgUpdateGroupAccountMetadataResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	metadata := req.GetMetadata()

	action := func(groupAccount *group.GroupAccountInfo) error {
		groupAccount.Metadata = metadata
		groupAccount.Version++
		return s.groupAccountTable.Update(ctx, groupAccount)
	}

	err := s.doUpdateGroupAccount(ctx, req.Address, req.Admin, action, "group account metadata updated")
	if err != nil {
		return nil, err
	}

	return &group.MsgUpdateGroupAccountMetadataResponse{}, nil
}

func (s serverImpl) CreateProposal(goCtx context.Context, req *group.MsgCreateProposal) (*group.MsgCreateProposalResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	accountAddress, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "request group account")
	}
	metadata := req.Metadata
	proposers := req.Proposers
	msgs := req.GetMsgs()

	account, err := s.getGroupAccountInfo(ctx, accountAddress.Bytes())
	if err != nil {
		return nil, sdkerrors.Wrap(err, "load group account")
	}

	g, err := s.getGroupInfo(ctx, account.GroupId)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "get group by account")
	}

	// Only members of the group can submit a new proposal.
	for i := range proposers {
		if !s.groupMemberTable.Has(ctx, orm.PrimaryKey(&group.GroupMember{GroupId: g.GroupId, Member: &group.Member{Address: proposers[i]}})) {
			return nil, sdkerrors.Wrapf(group.ErrUnauthorized, "not in group: %s", proposers[i])
		}
	}

	// Check that if the messages require signers, they are all equal to the given group account.
	if err := ensureMsgAuthZ(msgs, accountAddress); err != nil {
		return nil, err
	}

	blockTime, err := gogotypes.TimestampProto(ctx.BlockTime())
	if err != nil {
		return nil, sdkerrors.Wrap(err, "block time conversion")
	}

	policy := account.GetDecisionPolicy()
	if policy == nil {
		return nil, sdkerrors.Wrap(group.ErrEmpty, "nil policy")
	}

	// Prevent proposal that can not succeed.
	err = policy.Validate(g)
	if err != nil {
		return nil, err
	}

	// Define proposal timout.
	// The voting window begins as soon as the proposal is submitted.
	timeout := policy.GetTimeout()
	window, err := gogotypes.DurationFromProto(&timeout)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "maxVotingWindow time conversion")
	}
	endTime, err := gogotypes.TimestampProto(ctx.BlockTime().Add(window))
	if err != nil {
		return nil, sdkerrors.Wrap(err, "end time conversion")
	}

	m := &group.Proposal{
		ProposalId:          s.proposalTable.Sequence().PeekNextVal(ctx),
		Address:             req.Address,
		Metadata:            metadata,
		Proposers:           proposers,
		SubmittedAt:         *blockTime,
		GroupVersion:        g.Version,
		GroupAccountVersion: account.Version,
		Result:              group.ProposalResultUnfinalized,
		Status:              group.ProposalStatusSubmitted,
		ExecutorResult:      group.ProposalExecutorResultNotRun,
		Timeout:             *endTime,
		VoteState: group.Tally{
			YesCount:     "0",
			NoCount:      "0",
			AbstainCount: "0",
			VetoCount:    "0",
		},
	}
	if err := m.SetMsgs(msgs); err != nil {
		return nil, sdkerrors.Wrap(err, "create proposal")
	}

	id, err := s.proposalTable.Create(ctx, m)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "create proposal")
	}

	err = ctx.EventManager().EmitTypedEvent(&group.EventCreateProposal{ProposalId: id})
	if err != nil {
		return nil, err
	}

	// Try to execute proposal immediately
	if req.Exec == group.Exec_EXEC_TRY {
		// Consider proposers as Yes votes
		for i := range proposers {
			ctx.GasMeter().ConsumeGas(gasCostPerIteration, "vote on proposal")
			_, err = s.Vote(ctx, &group.MsgVote{
				ProposalId: id,
				Voter:      proposers[i],
				Choice:     group.Choice_CHOICE_YES,
			})
			if err != nil {
				return &group.MsgCreateProposalResponse{ProposalId: id}, sdkerrors.Wrap(err, "The proposal was created but failed on vote")
			}
		}
		// Then try to execute the proposal
		_, err = s.Exec(ctx, &group.MsgExec{
			ProposalId: id,
			// We consider the first proposer as the MsgExecRequest signer
			// but that could be revisited (eg using the group account)
			Signer: proposers[0],
		})
		if err != nil {
			return &group.MsgCreateProposalResponse{ProposalId: id}, sdkerrors.Wrap(err, "The proposal was created but failed on exec")
		}
	}

	return &group.MsgCreateProposalResponse{ProposalId: id}, nil
}

func (s serverImpl) Vote(goCtx context.Context, req *group.MsgVote) (*group.MsgVoteResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	id := req.ProposalId
	choice := req.Choice
	metadata := req.Metadata

	blockTime, err := gogotypes.TimestampProto(ctx.BlockTime())
	if err != nil {
		return nil, err
	}
	proposal, err := s.getProposal(ctx, id)
	if err != nil {
		return nil, err
	}
	// Ensure that we can still accept votes for this proposal.
	if proposal.Status != group.ProposalStatusSubmitted {
		return nil, sdkerrors.Wrap(group.ErrInvalid, "proposal not open for voting")
	}
	votingPeriodEnd, err := gogotypes.TimestampFromProto(&proposal.Timeout)
	if err != nil {
		return nil, err
	}
	if votingPeriodEnd.Before(ctx.BlockTime()) || votingPeriodEnd.Equal(ctx.BlockTime()) {
		return nil, sdkerrors.Wrap(group.ErrExpired, "voting period has ended already")
	}

	// Ensure that group account hasn't been modified since the proposal submission.
	address, err := sdk.AccAddressFromBech32(proposal.Address)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "group account")
	}
	accountInfo, err := s.getGroupAccountInfo(ctx, address.Bytes())
	if err != nil {
		return nil, sdkerrors.Wrap(err, "load group account")
	}
	if proposal.GroupAccountVersion != accountInfo.Version {
		return nil, sdkerrors.Wrap(group.ErrModified, "group account was modified")
	}

	// Ensure that group hasn't been modified since the proposal submission.
	electorate, err := s.getGroupInfo(ctx, accountInfo.GroupId)
	if err != nil {
		return nil, err
	}
	if electorate.Version != proposal.GroupVersion {
		return nil, sdkerrors.Wrap(group.ErrModified, "group was modified")
	}

	// Count and store votes.
	voterAddr := req.Voter
	voter := group.GroupMember{GroupId: electorate.GroupId, Member: &group.Member{Address: voterAddr}}
	if err := s.groupMemberTable.GetOne(ctx, orm.PrimaryKey(&voter), &voter); err != nil {
		return nil, sdkerrors.Wrapf(err, "address: %s", voterAddr)
	}
	newVote := group.Vote{
		ProposalId:  id,
		Voter:       voterAddr,
		Choice:      choice,
		Metadata:    metadata,
		SubmittedAt: *blockTime,
	}
	if err := proposal.VoteState.Add(newVote, voter.Member.Weight); err != nil {
		return nil, sdkerrors.Wrap(err, "add new vote")
	}

	// The ORM will return an error if the vote already exists,
	// making sure than a voter hasn't already voted.
	if err := s.voteTable.Create(ctx, &newVote); err != nil {
		return nil, sdkerrors.Wrap(err, "store vote")
	}

	// Run tally with new votes to close early.
	if err := doTally(ctx, &proposal, electorate, accountInfo); err != nil {
		return nil, err
	}

	if err = s.proposalTable.Update(ctx, id, &proposal); err != nil {
		return nil, err
	}

	err = ctx.EventManager().EmitTypedEvent(&group.EventVote{ProposalId: id})
	if err != nil {
		return nil, err
	}

	// Try to execute proposal immediately
	if req.Exec == group.Exec_EXEC_TRY {
		_, err = s.Exec(ctx, &group.MsgExec{
			ProposalId: id,
			Signer:     voterAddr,
		})
		if err != nil {
			return nil, err
		}
	}

	return &group.MsgVoteResponse{}, nil
}

// doTally updates the proposal status and tally if necessary based on the group account's decision policy.
func doTally(ctx types.Context, p *group.Proposal, electorate group.GroupInfo, accountInfo group.GroupAccountInfo) error {
	policy := accountInfo.GetDecisionPolicy()
	submittedAt, err := gogotypes.TimestampFromProto(&p.SubmittedAt)
	if err != nil {
		return err
	}
	switch result, err := policy.Allow(p.VoteState, electorate.TotalWeight, ctx.BlockTime().Sub(submittedAt)); {
	case err != nil:
		return sdkerrors.Wrap(err, "policy execution")
	case result.Allow && result.Final:
		p.Result = group.ProposalResultAccepted
		p.Status = group.ProposalStatusClosed
	case !result.Allow && result.Final:
		p.Result = group.ProposalResultRejected
		p.Status = group.ProposalStatusClosed
	}
	return nil
}

func (s serverImpl) VoteAgg(goCtx context.Context, req *group.MsgVoteAgg) (*group.MsgVoteAggResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	id := req.ProposalId
	choices := req.Votes
	metadata := req.Metadata
	votesExpire := req.Expiry

	blockTime, err := gogotypes.TimestampProto(ctx.BlockTime())
	if err != nil {
		return nil, err
	}

	if votesExpire.Compare(blockTime) <= 0 {
		return nil, sdkerrors.Wrap(group.ErrExpired, "the aggregated votes have expired")
	}

	proposal, err := s.getProposal(ctx, id)
	if err != nil {
		return nil, err
	}
	// Ensure that we can still accept votes for this proposal.
	if proposal.Status != group.ProposalStatusSubmitted {
		return nil, sdkerrors.Wrap(group.ErrInvalid, "proposal not open for voting")
	}
	if proposal.Timeout.Compare(blockTime) <= 0 {
		return nil, sdkerrors.Wrap(group.ErrExpired, "voting period has ended already")
	}

	// Ensure that group account hasn't been modified since the proposal submission.
	address, err := sdk.AccAddressFromBech32(proposal.Address)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "group account")
	}
	accountInfo, err := s.getGroupAccountInfo(ctx, address.Bytes())
	if err != nil {
		return nil, sdkerrors.Wrap(err, "load group account")
	}
	if proposal.GroupAccountVersion != accountInfo.Version {
		return nil, sdkerrors.Wrap(group.ErrModified, "group account was modified")
	}

	// Ensure that group hasn't been modified since the proposal submission.
	electorate, err := s.getGroupInfo(ctx, accountInfo.GroupId)
	if err != nil {
		return nil, err
	}
	if electorate.Version != proposal.GroupVersion {
		return nil, sdkerrors.Wrap(group.ErrModified, "group was modified")
	}

	membersIter, err := s.groupMemberByGroupIndex.Get(ctx, electorate.GroupId)
	if err != nil {
		return nil, err
	}

	var votes []group.Vote
	var weights []string
	choiceMap := make(map[group.Choice]bool, len(group.Choice_name))
	pkMap := make(map[group.Choice][]cryptotypes.PubKey, len(group.Choice_name))
	msgMap := make(map[group.Choice][]byte, len(group.Choice_name))
	for i := 0; ; i++ {
		var mem group.GroupMember
		_, err := membersIter.LoadNext(&mem)
		if err != nil {
			if orm.ErrIteratorDone.Is(err) {
				if i < len(choices) {
					return nil, sdkerrors.Wrap(group.ErrInvalid, "too many votes")
				}
				break
			}
			return nil, err
		}

		memAddr, err := sdk.AccAddressFromBech32(mem.Member.Address)
		if err != nil {
			return nil, err
		}

		if i >= len(choices) {
			return nil, sdkerrors.Wrap(group.ErrInvalid, "not enough votes")
		}

		acc := s.accKeeper.GetAccount(ctx.Context, memAddr)
		if acc == nil {
			return nil, sdkerrors.Wrapf(group.ErrInvalid, "account %s does not exist", memAddr.String())
		}
		pk := acc.GetPubKey()
		if pk == nil {
			return nil, sdkerrors.Wrapf(group.ErrInvalid, "public key for account %s not set yet", memAddr.String())
		}

		if choices[i] != group.Choice_CHOICE_UNSPECIFIED {
			vote := group.Vote{
				ProposalId:  id,
				Voter:       mem.Member.Address,
				Choice:      choices[i],
				Metadata:    metadata,
				SubmittedAt: *blockTime,
			}
			votes = append(votes, vote)
			weights = append(weights, mem.Member.Weight)

			_, ok := choiceMap[choices[i]]
			if !ok {
				choiceMap[choices[i]] = true
				msg := group.MsgVoteBasic{
					ProposalId: id,
					Choice:     choices[i],
					Expiry:     votesExpire,
				}
				msgMap[choices[i]] = msg.GetSignBytes()
			}
			pkMap[choices[i]] = append(pkMap[choices[i]], pk)
		}
	}

	// calculate and consume gas before the verification of the aggregated signature
	numChoice := uint64(len(choiceMap))
	numPk := uint64(len(votes))
	params := s.accKeeper.GetParams(ctx.Context)
	err = DefaultAggSigVerifyGasConsumer(ctx.GasMeter(), numPk, numChoice, params)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "gas consumption for verifying aggregated signature")
	}

	msgBytes := make([][]byte, 0, numChoice)
	pkss := make([][]cryptotypes.PubKey, 0, numChoice)
	for c := range msgMap {
		msgBytes = append(msgBytes, msgMap[c])
		pkss = append(pkss, pkMap[c])
	}

	if err = group.VerifyAggregateSignature(msgBytes, false, req.AggSig, pkss); err != nil {
		return nil, err
	}

	// Count and store votes.
	for i := range votes {
		err := proposal.VoteState.Add(votes[i], weights[i])
		if err == nil {
			// If the vote already exists, skip the new vote
			if err := s.voteTable.Create(ctx, &votes[i]); err != nil {
				if orm.ErrUniqueConstraint.Is(err) {
					if err := proposal.VoteState.Sub(votes[i], weights[i]); err != nil {
						return nil, sdkerrors.Wrap(err, "sub new vote")
					}
					continue
				}
				return nil, err
			}
		}
	}

	if err := doTally(ctx, &proposal, electorate, accountInfo); err != nil {
		return nil, err
	}

	if err = s.proposalTable.Set(ctx, id, &proposal); err != nil {
		return nil, err
	}

	err = ctx.EventManager().EmitTypedEvent(&group.EventVote{ProposalId: id})
	if err != nil {
		return nil, err
	}

	// Try to execute proposal immediately
	if req.Exec == group.Exec_EXEC_TRY {
		_, err = s.Exec(ctx, &group.MsgExec{
			ProposalId: id,
			Signer:     req.Sender,
		})
		if err != nil {
			return nil, err
		}
	}

	return &group.MsgVoteAggResponse{}, nil
}

// Exec executes the messages from a proposal.
func (s serverImpl) Exec(goCtx context.Context, req *group.MsgExec) (*group.MsgExecResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	id := req.ProposalId

	proposal, err := s.getProposal(ctx, id)
	if err != nil {
		return nil, err
	}

	if proposal.Status != group.ProposalStatusSubmitted && proposal.Status != group.ProposalStatusClosed {
		return nil, sdkerrors.Wrapf(group.ErrInvalid, "not possible with proposal status %s", proposal.Status.String())
	}

	address, err := sdk.AccAddressFromBech32(proposal.Address)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "group account")
	}
	accountInfo, err := s.getGroupAccountInfo(ctx, address.Bytes())
	if err != nil {
		return nil, sdkerrors.Wrap(err, "load group account")
	}

	storeUpdates := func() (*group.MsgExecResponse, error) {
		if err := s.proposalTable.Update(ctx, id, &proposal); err != nil {
			return nil, err
		}
		return &group.MsgExecResponse{}, nil
	}

	if proposal.Status == group.ProposalStatusSubmitted {
		// Ensure that group account hasn't been modified before tally.
		if proposal.GroupAccountVersion != accountInfo.Version {
			proposal.Result = group.ProposalResultUnfinalized
			proposal.Status = group.ProposalStatusAborted
			return storeUpdates()
		}

		electorate, err := s.getGroupInfo(ctx, accountInfo.GroupId)
		if err != nil {
			return nil, sdkerrors.Wrap(err, "load group")
		}

		// Ensure that group hasn't been modified before tally.
		if electorate.Version != proposal.GroupVersion {
			proposal.Result = group.ProposalResultUnfinalized
			proposal.Status = group.ProposalStatusAborted
			return storeUpdates()
		}
		if err := doTally(ctx, &proposal, electorate, accountInfo); err != nil {
			return nil, err
		}
	}

	// Execute proposal payload.
	if proposal.Status == group.ProposalStatusClosed && proposal.Result == group.ProposalResultAccepted && proposal.ExecutorResult != group.ProposalExecutorResultSuccess {
		logger := ctx.Logger().With("module", fmt.Sprintf("x/%s", group.ModuleName))
		// Cashing context so that we don't update the store in case of failure.
		ctx, flush := ctx.CacheContext()

		err := s.execMsgs(sdk.WrapSDKContext(ctx), accountInfo.DerivationKey, proposal)
		if err != nil {
			proposal.ExecutorResult = group.ProposalExecutorResultFailure
			proposalType := reflect.TypeOf(proposal).String()
			logger.Info("proposal execution failed", "cause", err, "type", proposalType, "proposalID", id)
		} else {
			proposal.ExecutorResult = group.ProposalExecutorResultSuccess
			flush()
		}
	}

	// Update proposal in proposalTable
	res, err := storeUpdates()
	if err != nil {
		return nil, err
	}

	err = ctx.EventManager().EmitTypedEvent(&group.EventExec{ProposalId: id})
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (s serverImpl) CreatePoll(goCtx context.Context, req *group.MsgCreatePoll) (*group.MsgCreatePollResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	creator := req.Creator
	groupID := req.GroupId
	endTime := req.Timeout
	metadata := req.Metadata
	title := req.Title
	options := req.Options
	limit := req.VoteLimit

	g, err := s.getGroupInfo(ctx, groupID)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "get group by account")
	}

	// Only members of the group can submit a new poll.
	if !s.groupMemberTable.Has(ctx, orm.PrimaryKey(&group.GroupMember{GroupId: g.GroupId, Member: &group.Member{Address: creator}})) {
		return nil, sdkerrors.Wrapf(group.ErrUnauthorized, "not in group: %s", creator)
	}

	blockTime, err := gogotypes.TimestampProto(ctx.BlockTime())
	if err != nil {
		return nil, sdkerrors.Wrap(err, "block time conversion")
	}

	if endTime.Compare(blockTime) <= 0 {
		return nil, sdkerrors.Wrap(group.ErrExpired, "poll already expired")
	}

	sort.Strings(options.Titles)

	m := &group.Poll{
		PollId:       s.pollTable.Sequence().PeekNextVal(ctx),
		GroupId:      groupID,
		Title:        title,
		Options:      options,
		Creator:      creator,
		VoteLimit:    limit,
		Metadata:     metadata,
		SubmittedAt:  *blockTime,
		GroupVersion: g.Version,
		Status:       group.PollStatusSubmitted,
		Timeout:      endTime,
	}

	id, err := s.pollTable.Create(ctx, m)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "create poll")
	}

	err = ctx.EventManager().EmitTypedEvent(&group.EventCreatePoll{PollId: id})
	if err != nil {
		return nil, err
	}

	return &group.MsgCreatePollResponse{PollId: id}, nil
}

func (s serverImpl) VotePoll(goCtx context.Context, req *group.MsgVotePoll) (*group.MsgVotePollResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	id := req.PollId
	options := req.Options
	metadata := req.Metadata

	blockTime, err := gogotypes.TimestampProto(ctx.BlockTime())
	if err != nil {
		return nil, err
	}
	poll, err := s.getPoll(ctx, id)
	if err != nil {
		return nil, err
	}
	// Ensure that we can still accept votes for this poll.
	if poll.Status != group.PollStatusSubmitted {
		return nil, sdkerrors.Wrap(group.ErrInvalid, "poll not open for voting")
	}
	if poll.Timeout.Compare(blockTime) <= 0 {
		poll.Status = group.PollStatusClosed
		if err = s.pollTable.Set(ctx, id, &poll); err != nil {
			return nil, err
		}
		return nil, sdkerrors.Wrap(group.ErrExpired, "voting period for the poll has ended")
	}

	if err := assertOptionsSubset(options, poll.Options, "options"); err != nil {
		return nil, err
	}

	if len(options.Titles) > int(poll.VoteLimit) {
		return nil, sdkerrors.Wrap(group.ErrInvalid, "voter options exceed limit")
	}

	// Ensure that group hasn't been modified since the proposal submission.
	electorate, err := s.getGroupInfo(ctx, poll.GroupId)
	if err != nil {
		return nil, err
	}
	if electorate.Version != poll.GroupVersion {
		return nil, sdkerrors.Wrap(group.ErrModified, "group was modified")
	}

	// Count and store votes.
	voterAddr := req.Voter
	voter := group.GroupMember{GroupId: electorate.GroupId, Member: &group.Member{Address: voterAddr}}
	if err := s.groupMemberTable.GetOne(ctx, orm.PrimaryKey(&voter), &voter); err != nil {
		return nil, sdkerrors.Wrapf(err, "address: %s", voterAddr)
	}

	sort.Strings(options.Titles)

	newVote := group.VotePoll{
		PollId:      id,
		Voter:       voterAddr,
		Options:     options,
		Metadata:    metadata,
		SubmittedAt: *blockTime,
	}
	if err := poll.VoteState.Add(newVote, voter.Member.Weight); err != nil {
		return nil, sdkerrors.Wrap(err, "add new vote")
	}

	// The ORM will return an error if the vote already exists,
	// making sure than a voter hasn't already voted.
	if err := s.votePollTable.Create(ctx, &newVote); err != nil {
		return nil, sdkerrors.Wrap(err, "store vote")
	}

	sort.SliceStable(poll.VoteState.Entries, func(i, j int) bool {
		return poll.VoteState.Entries[i].OptionTitle < poll.VoteState.Entries[j].OptionTitle
	})

	if err = s.pollTable.Set(ctx, id, &poll); err != nil {
		return nil, err
	}

	err = ctx.EventManager().EmitTypedEvent(&group.EventVotePoll{PollId: id})
	if err != nil {
		return nil, err
	}

	return &group.MsgVotePollResponse{}, nil
}

// VotePollAgg processes an aggregated votes for poll. The votes are signed and aggregated
// on an option-basis instead of voter-basis, which means the verification time of the
// aggregated signature is linear of the number of options instead of the number of voters.
// The total number of options should be kept small so the verification of the aggregated signature can be more efficient.
func (s serverImpl) VotePollAgg(goCtx context.Context, req *group.MsgVotePollAgg) (*group.MsgVotePollAggResponse, error) {
	ctx := types.UnwrapSDKContext(goCtx)
	id := req.PollId
	votes := req.Votes
	metadata := req.Metadata
	votesExpire := req.Expiry

	blockTime, err := gogotypes.TimestampProto(ctx.BlockTime())
	if err != nil {
		return nil, err
	}

	if votesExpire.Compare(blockTime) <= 0 {
		return nil, sdkerrors.Wrap(group.ErrExpired, "the aggregated votes have expired")
	}

	poll, err := s.getPoll(ctx, id)
	if err != nil {
		return nil, err
	}
	// Ensure that we can still accept votes for this poll.
	if poll.Status != group.PollStatusSubmitted {
		return nil, sdkerrors.Wrap(group.ErrInvalid, "poll not open for voting")
	}
	if poll.Timeout.Compare(blockTime) <= 0 {
		poll.Status = group.PollStatusClosed
		if err = s.pollTable.Set(ctx, id, &poll); err != nil {
			return nil, err
		}
		return nil, sdkerrors.Wrap(group.ErrExpired, "voting period for the poll has ended")
	}

	// Ensure that group hasn't been modified since the proposal submission.
	electorate, err := s.getGroupInfo(ctx, poll.GroupId)
	if err != nil {
		return nil, err
	}
	if electorate.Version != poll.GroupVersion {
		return nil, sdkerrors.Wrap(group.ErrModified, "group was modified")
	}

	for _, v := range votes {
		if err = assertOptionsSubset(v, poll.Options, "aggregated vote for poll"); err != nil {
			return nil, sdkerrors.Wrap(group.ErrInvalid, err.Error())
		}
		if len(v.Titles) > int(poll.VoteLimit) {
			return nil, sdkerrors.Wrap(group.ErrInvalid, "voter options exceed limit")
		}
	}

	membersIter, err := s.groupMemberByGroupIndex.Get(ctx, electorate.GroupId)
	if err != nil {
		return nil, err
	}

	// Count and store votes.
	var weights []string
	pkMap := make(map[string][]cryptotypes.PubKey, len(poll.Options.Titles))
	msgMap := make(map[string][]byte, len(poll.Options.Titles))
	var votesStore []group.VotePoll
	numPk := uint64(0)
	for i := 0; ; i++ {
		var mem group.GroupMember
		_, err := membersIter.LoadNext(&mem)
		if err != nil {
			if orm.ErrIteratorDone.Is(err) {
				if i < len(votes) {
					return nil, sdkerrors.Wrap(group.ErrInvalid, "too many votes")
				}
				break
			}
			return nil, err
		}

		memAddr, err := sdk.AccAddressFromBech32(mem.Member.Address)
		if err != nil {
			return nil, err
		}

		if i >= len(votes) {
			return nil, sdkerrors.Wrap(group.ErrInvalid, "not enough votes")
		}

		acc := s.accKeeper.GetAccount(ctx.Context, memAddr)
		if acc == nil {
			return nil, sdkerrors.Wrapf(group.ErrInvalid, "account %s does not exist", memAddr.String())
		}
		pk := acc.GetPubKey()
		if pk == nil {
			return nil, sdkerrors.Wrapf(group.ErrInvalid, "public key for account %s not set yet", memAddr.String())
		}

		if len(votes[i].Titles) != 0 {
			vote := group.VotePoll{
				PollId:      id,
				Voter:       mem.Member.Address,
				Options:     votes[i],
				Metadata:    metadata,
				SubmittedAt: *blockTime,
			}
			votesStore = append(votesStore, vote)
			weights = append(weights, mem.Member.Weight)
			for _, ot := range votes[i].Titles {
				if _, ok := msgMap[ot]; !ok {
					msg := group.MsgVotePollBasic{
						PollId: id,
						Option: ot,
						Expiry: votesExpire,
					}
					msgMap[ot] = msg.GetSignBytes()
				}
				pkMap[ot] = append(pkMap[ot], pk)
				numPk++
			}
		}
	}

	// calculate and consume gas before the verification of the aggregated signature
	numVotedOption := uint64(len(msgMap))
	params := s.accKeeper.GetParams(ctx.Context)
	err = DefaultAggSigVerifyGasConsumer(ctx.GasMeter(), numPk, numVotedOption, params)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "gas consumption for verifying aggregated signature")
	}

	msgBytes := make([][]byte, 0, numVotedOption)
	pkss := make([][]cryptotypes.PubKey, 0, numVotedOption)
	for c := range msgMap {
		msgBytes = append(msgBytes, msgMap[c])
		pkss = append(pkss, pkMap[c])
	}

	if err = group.VerifyAggregateSignature(msgBytes, false, req.AggSig, pkss); err != nil {
		return nil, err
	}

	for i := range votesStore {
		// skip the vote if error
		err := poll.VoteState.Add(votesStore[i], weights[i])
		if err != nil {
			continue
		}
		// If the vote already exists, skip the new vote
		if err := s.votePollTable.Create(ctx, &votesStore[i]); err != nil {
			if orm.ErrUniqueConstraint.Is(err) {
				if err := poll.VoteState.Sub(votesStore[i], weights[i]); err != nil {
					return nil, sdkerrors.Wrap(err, "sub new vote")
				}
				continue
			}
			return nil, err
		}
	}

	sort.SliceStable(poll.VoteState.Entries, func(i, j int) bool {
		return poll.VoteState.Entries[i].OptionTitle < poll.VoteState.Entries[j].OptionTitle
	})
	if err = s.pollTable.Set(ctx, id, &poll); err != nil {
		return nil, err
	}

	err = ctx.EventManager().EmitTypedEvent(&group.EventVotePoll{PollId: id})
	if err != nil {
		return nil, err
	}

	return &group.MsgVotePollAggResponse{}, nil
}

type authNGroupReq interface {
	GetGroupID() uint64
	GetAdmin() string
}

type actionFn func(m *group.GroupInfo) error
type groupAccountActionFn func(m *group.GroupAccountInfo) error

// doUpdateGroupAccount first makes sure that the group account admin initiated the group account update,
// before performing the group account update and emitting an event.
func (s serverImpl) doUpdateGroupAccount(ctx types.Context, groupAccount string, admin string, action groupAccountActionFn, note string) error {
	groupAccountAddress, err := sdk.AccAddressFromBech32(groupAccount)
	if err != nil {
		return sdkerrors.Wrap(err, "group admin")
	}

	groupAccountInfo, err := s.getGroupAccountInfo(ctx, groupAccountAddress.Bytes())
	if err != nil {
		return sdkerrors.Wrap(err, "load group account")
	}

	groupAccountAdmin, err := sdk.AccAddressFromBech32(admin)
	if err != nil {
		return sdkerrors.Wrap(err, "group account admin")
	}

	// Only current group account admin is authorized to update a group account.
	if groupAccountAdmin.String() != groupAccountInfo.Admin {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "not group account admin")
	}

	if err := action(&groupAccountInfo); err != nil {
		return sdkerrors.Wrap(err, note)
	}

	err = ctx.EventManager().EmitTypedEvent(&group.EventUpdateGroupAccount{Address: admin})
	if err != nil {
		return err
	}

	return nil
}

// doUpdateGroup first makes sure that the group admin initiated the group update,
// before performing the group update and emitting an event.
func (s serverImpl) doUpdateGroup(ctx types.Context, req authNGroupReq, action actionFn, note string) error {
	err := s.doAuthenticated(ctx, req, action, note)
	if err != nil {
		return err
	}

	err = ctx.EventManager().EmitTypedEvent(&group.EventUpdateGroup{GroupId: req.GetGroupID()})
	if err != nil {
		return err
	}

	return nil
}

// doAuthenticated makes sure that the group admin initiated the request,
// and perform the provided action on the group.
func (s serverImpl) doAuthenticated(ctx types.Context, req authNGroupReq, action actionFn, note string) error {
	group, err := s.getGroupInfo(ctx, req.GetGroupID())
	if err != nil {
		return err
	}
	admin, err := sdk.AccAddressFromBech32(group.Admin)
	if err != nil {
		return sdkerrors.Wrap(err, "group admin")
	}
	reqAdmin, err := sdk.AccAddressFromBech32(req.GetAdmin())
	if err != nil {
		return sdkerrors.Wrap(err, "request admin")
	}
	if !admin.Equals(reqAdmin) {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "not group admin")
	}
	if err := action(&group); err != nil {
		return sdkerrors.Wrap(err, note)
	}
	return nil
}

func assertOptionsSubset(a group.Options, b group.Options, description string) error {
	for _, x := range a.Titles {
		foundX := false
		for _, y := range b.Titles {
			if x == y {
				foundX = true
				break
			}
		}
		if !foundX {
			return sdkerrors.Wrap(group.ErrInvalid, description)
		}
	}
	return nil
}
