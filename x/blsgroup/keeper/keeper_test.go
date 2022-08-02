package keeper_test

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"testing"
	"time"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdktestdata "github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/fetchai/fetchd/app"
	"github.com/fetchai/fetchd/crypto/keys/bls12381"
	"github.com/fetchai/fetchd/testutil"
	"github.com/fetchai/fetchd/testutil/testdata"
	"github.com/fetchai/fetchd/x/blsgroup"
)

type testAccount struct {
	Pubkey  cryptotypes.PubKey
	PrivKey cryptotypes.PrivKey
	Addr    sdk.AccAddress
	Weight  uint64
}

type TestSuite struct {
	suite.Suite

	app             *app.App
	sdkCtx          sdk.Context
	ctx             context.Context
	accounts        []testAccount
	groupAdmin      sdk.AccAddress
	groupID         uint64
	groupPolicyAddr sdk.AccAddress
	policy          group.DecisionPolicy
	blockTime       time.Time
}

func (s *TestSuite) SetupTest() {
	app := testutil.Setup(s.T(), false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	s.blockTime = time.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: s.blockTime})

	s.app = app
	s.sdkCtx = ctx
	s.ctx = sdk.WrapSDKContext(ctx)

	s.accounts = make([]testAccount, 3)
	pubkeys := make([]cryptotypes.PubKey, 3)
	for i := 0; i < 3; i++ {
		priv, pub, addr := testdata.KeyTestPubAddrBls12381()
		s.accounts[i] = testAccount{
			Pubkey:  pub,
			PrivKey: priv,
			Addr:    addr,
			Weight:  2,
		}
		pubkeys[i] = pub
	}
	testutil.AddTestAddrsFromPubKeys(app, ctx, pubkeys, sdk.NewInt(30000000))

	// accounts need to be sorted to properly produce aggregated signatures
	sort.SliceStable(s.accounts, func(i, j int) bool {
		return bytes.Compare(s.accounts[i].Addr, s.accounts[j].Addr) < 0
	})

	// make pop checks happy
	for _, testAcc := range s.accounts {
		acc := s.app.AccountKeeper.GetAccount(ctx, testAcc.Addr)
		s.Require().NoError(acc.SetPubKey(testAcc.Pubkey))
		s.Require().NoError(acc.SetSequence(1))
		s.app.AccountKeeper.SetAccount(ctx, acc)
	}

	s.groupAdmin = s.accounts[0].Addr

	// Initial group, group policy and balance setup
	members := []group.MemberRequest{
		{Address: s.accounts[0].Addr.String(), Weight: fmt.Sprintf("%d", s.accounts[0].Weight)},
		{Address: s.accounts[1].Addr.String(), Weight: fmt.Sprintf("%d", s.accounts[1].Weight)},
		{Address: s.accounts[2].Addr.String(), Weight: fmt.Sprintf("%d", s.accounts[2].Weight)},
	}

	groupRes, err := s.app.GroupKeeper.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin:   s.groupAdmin.String(),
		Members: members,
	})
	s.Require().NoError(err)
	s.groupID = groupRes.GroupId

	// register group as a BLS group
	_, err = s.app.BlsGroupKeeper.RegisterBlsGroup(s.ctx, &blsgroup.MsgRegisterBlsGroup{
		Admin:   s.groupAdmin.String(),
		GroupId: s.groupID,
	})
	s.Require().NoError(err)

	policy := group.NewPercentageDecisionPolicy(
		"0.5",
		time.Second,
		0,
	)
	policyReq := &group.MsgCreateGroupPolicy{
		Admin:   s.groupAdmin.String(),
		GroupId: s.groupID,
	}
	err = policyReq.SetDecisionPolicy(policy)
	s.Require().NoError(err)
	policyRes, err := s.app.GroupKeeper.CreateGroupPolicy(s.ctx, policyReq)
	s.Require().NoError(err)
	s.policy = policy
	addr, err := sdk.AccAddressFromBech32(policyRes.Address)
	s.Require().NoError(err)
	s.groupPolicyAddr = addr
	s.Require().NoError(bankutil.FundAccount(s.app.BankKeeper, s.sdkCtx, s.groupPolicyAddr, sdk.Coins{sdk.NewInt64Coin("token", 10000)}))

}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestRegisterBlsGroup() {
	unregisteredGroup, err := s.app.GroupKeeper.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin: s.groupAdmin.String(),
		Members: []group.MemberRequest{
			{Address: s.accounts[0].Addr.String(), Weight: "1"},
			{Address: s.accounts[1].Addr.String(), Weight: "2"},
		},
	})
	s.Require().NoError(err)

	_, nonBlsPubkey, nonBlsButExistingAddr := sdktestdata.KeyTestPubAddr()
	testutil.AddTestAddrsFromPubKeys(s.app, s.sdkCtx, []cryptotypes.PubKey{nonBlsPubkey}, sdk.NewInt(30000000))
	acc := s.app.AccountKeeper.GetAccount(s.sdkCtx, nonBlsButExistingAddr)
	s.Require().NoError(acc.SetPubKey(nonBlsPubkey))
	s.app.AccountKeeper.SetAccount(s.sdkCtx, acc)
	nonBlsMemberGroup, err := s.app.GroupKeeper.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin: s.groupAdmin.String(),
		Members: []group.MemberRequest{
			{Address: s.accounts[0].Addr.String(), Weight: "1"},
			{Address: nonBlsButExistingAddr.String(), Weight: "2"},
		},
	})
	s.Require().NoError(err)

	_, _, blsButNonExistingAddr := testdata.KeyTestPubAddrBls12381()
	nonExistingMemberGroup, err := s.app.GroupKeeper.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin: s.groupAdmin.String(),
		Members: []group.MemberRequest{
			{Address: s.accounts[0].Addr.String(), Weight: "1"},
			{Address: blsButNonExistingAddr.String(), Weight: "2"},
		},
	})
	s.Require().NoError(err)

	_, blsButPubkeyNotSetPubkey, blsButPubkeyNotSetAddr := testdata.KeyTestPubAddrBls12381()
	testutil.AddTestAddrsFromPubKeys(s.app, s.sdkCtx, []cryptotypes.PubKey{blsButPubkeyNotSetPubkey}, sdk.NewInt(30000000))
	memberPubkeyNotSetGroup, err := s.app.GroupKeeper.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin: s.groupAdmin.String(),
		Members: []group.MemberRequest{
			{Address: s.accounts[0].Addr.String(), Weight: "1"},
			{Address: blsButPubkeyNotSetAddr.String(), Weight: "2"},
		},
	})
	s.Require().NoError(err)

	_, blsMissingPOPPubkey, blsMissingPOPAddr := testdata.KeyTestPubAddrBls12381()
	testutil.AddTestAddrsFromPubKeys(s.app, s.sdkCtx, []cryptotypes.PubKey{blsMissingPOPPubkey}, sdk.NewInt(30000000))
	acc = s.app.AccountKeeper.GetAccount(s.sdkCtx, blsMissingPOPAddr)
	s.Require().NoError(acc.SetPubKey(blsMissingPOPPubkey))
	s.app.AccountKeeper.SetAccount(s.sdkCtx, acc)
	memberMissingPOPGroup, err := s.app.GroupKeeper.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin: s.groupAdmin.String(),
		Members: []group.MemberRequest{
			{Address: s.accounts[0].Addr.String(), Weight: "1"},
			{Address: blsMissingPOPAddr.String(), Weight: "2"},
		},
	})
	s.Require().NoError(err)

	testcases := []struct {
		Description string
		Request     *blsgroup.MsgRegisterBlsGroup
		ExpectError bool
		Err         error
	}{
		{
			Description: "valid registration",
			Request: &blsgroup.MsgRegisterBlsGroup{
				Admin:   s.groupAdmin.String(),
				GroupId: unregisteredGroup.GroupId,
			},
			ExpectError: false,
		},
		{
			Description: "already registrered",
			Request: &blsgroup.MsgRegisterBlsGroup{
				Admin:   s.groupAdmin.String(),
				GroupId: unregisteredGroup.GroupId,
			},
			ExpectError: true,
			Err:         blsgroup.ErrDuplicate,
		},
		{
			Description: "unknown group",
			Request: &blsgroup.MsgRegisterBlsGroup{
				Admin:   s.groupAdmin.String(),
				GroupId: 65535,
			},
			ExpectError: true,
			Err:         sdkerrors.ErrNotFound,
		},
		{
			Description: "non-bls key member",
			Request: &blsgroup.MsgRegisterBlsGroup{
				Admin:   s.groupAdmin.String(),
				GroupId: nonBlsMemberGroup.GroupId,
			},
			ExpectError: true,
			Err:         blsgroup.ErrInvalid,
		},
		{
			Description: "member account does not exists",
			Request: &blsgroup.MsgRegisterBlsGroup{
				Admin:   s.groupAdmin.String(),
				GroupId: nonExistingMemberGroup.GroupId,
			},
			ExpectError: true,
			Err:         blsgroup.ErrInvalid,
		},
		{
			Description: "member account pubkey not set",
			Request: &blsgroup.MsgRegisterBlsGroup{
				Admin:   s.groupAdmin.String(),
				GroupId: memberPubkeyNotSetGroup.GroupId,
			},
			ExpectError: true,
			Err:         blsgroup.ErrInvalid,
		},
		{
			Description: "member account missing POP",
			Request: &blsgroup.MsgRegisterBlsGroup{
				Admin:   s.groupAdmin.String(),
				GroupId: memberMissingPOPGroup.GroupId,
			},
			ExpectError: true,
			Err:         blsgroup.ErrInvalid,
		},
		{
			Description: "not admin",
			Request: &blsgroup.MsgRegisterBlsGroup{
				Admin:   s.accounts[2].Addr.String(),
				GroupId: s.groupID,
			},
			ExpectError: true,
			Err:         blsgroup.ErrUnauthorized,
		},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.Description, func() {
			_, err := s.app.BlsGroupKeeper.RegisterBlsGroup(s.ctx, tc.Request)
			if tc.ExpectError {
				if tc.Err != nil {
					s.Require().ErrorIs(err, tc.Err)
				} else {
					s.Require().Error(err)
				}
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *TestSuite) TestRegisterModifiedBlsGroup() {
	blsGroup, err := s.app.GroupKeeper.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin: s.groupAdmin.String(),
		Members: []group.MemberRequest{
			{Address: s.accounts[0].Addr.String(), Weight: "1"},
			{Address: s.accounts[1].Addr.String(), Weight: "2"},
		},
	})
	s.Require().NoError(err)

	_, err = s.app.BlsGroupKeeper.RegisterBlsGroup(s.ctx, &blsgroup.MsgRegisterBlsGroup{
		Admin:   s.groupAdmin.String(),
		GroupId: blsGroup.GroupId,
	})
	s.Require().NoError(err, "unexpected error on initial group registration")

	_, err = s.app.BlsGroupKeeper.RegisterBlsGroup(s.ctx, &blsgroup.MsgRegisterBlsGroup{
		Admin:   s.groupAdmin.String(),
		GroupId: blsGroup.GroupId,
	})
	s.Require().ErrorIs(blsgroup.ErrDuplicate, err, "expected duplicate registration")

	_, err = s.app.GroupKeeper.UpdateGroupMembers(s.sdkCtx, &group.MsgUpdateGroupMembers{
		Admin:   s.groupAdmin.String(),
		GroupId: blsGroup.GroupId,
		MemberUpdates: []group.MemberRequest{
			{Address: s.accounts[0].Addr.String(), Weight: "3"},
			{Address: s.accounts[1].Addr.String(), Weight: "3"},
			{Address: s.accounts[2].Addr.String(), Weight: "3"},
		},
	})
	s.Require().NoError(err, "unexpected error on updating group members")

	_, err = s.app.BlsGroupKeeper.RegisterBlsGroup(s.ctx, &blsgroup.MsgRegisterBlsGroup{
		Admin:   s.groupAdmin.String(),
		GroupId: blsGroup.GroupId,
	})
	s.Require().NoError(err, "unexpected error on group registration after modification")
}

func (s *TestSuite) TestUnregisterBlsGroup() {
	nonRegisteredBlsGroup, err := s.app.GroupKeeper.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin: s.groupAdmin.String(),
		Members: []group.MemberRequest{
			{Address: s.accounts[0].Addr.String(), Weight: "1"},
			{Address: s.accounts[1].Addr.String(), Weight: "2"},
		},
	})
	s.Require().NoError(err)

	registeredBlsGroup, err := s.app.GroupKeeper.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin: s.groupAdmin.String(),
		Members: []group.MemberRequest{
			{Address: s.accounts[0].Addr.String(), Weight: "1"},
			{Address: s.accounts[1].Addr.String(), Weight: "2"},
		},
	})
	s.Require().NoError(err)
	_, err = s.app.BlsGroupKeeper.RegisterBlsGroup(s.ctx, &blsgroup.MsgRegisterBlsGroup{
		Admin:   s.groupAdmin.String(),
		GroupId: registeredBlsGroup.GroupId,
	})
	s.Require().NoError(err)

	testcases := []struct {
		Description string
		Request     *blsgroup.MsgUnregisterBlsGroup
		ExpectError bool
		Err         error
	}{
		{
			Description: "not registered yet",
			Request: &blsgroup.MsgUnregisterBlsGroup{
				Admin:   s.groupAdmin.String(),
				GroupId: nonRegisteredBlsGroup.GroupId,
			},
			ExpectError: true,
			Err:         blsgroup.ErrInvalid,
		},
		{
			Description: "unknown group",
			Request: &blsgroup.MsgUnregisterBlsGroup{
				Admin:   s.groupAdmin.String(),
				GroupId: 65535,
			},
			ExpectError: true,
			Err:         sdkerrors.ErrNotFound,
		},
		{
			Description: "not admin",
			Request: &blsgroup.MsgUnregisterBlsGroup{
				Admin:   s.accounts[1].Addr.String(),
				GroupId: registeredBlsGroup.GroupId,
			},
			ExpectError: true,
			Err:         blsgroup.ErrUnauthorized,
		},
		{
			Description: "valid",
			Request: &blsgroup.MsgUnregisterBlsGroup{
				Admin:   s.groupAdmin.String(),
				GroupId: registeredBlsGroup.GroupId,
			},
			ExpectError: false,
		},
	}

	for _, tc := range testcases {
		tc := tc
		s.Run(tc.Description, func() {
			_, err := s.app.BlsGroupKeeper.UnregisterBlsGroup(s.ctx, tc.Request)
			if tc.ExpectError {
				if tc.Err != nil {
					s.Require().ErrorIs(err, tc.Err)
				} else {
					s.Require().Error(err)
				}
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

func (s *TestSuite) TestVoteAggExecute() {
	proposalReq := &group.MsgSubmitProposal{
		GroupPolicyAddress: s.groupPolicyAddr.String(),
		Proposers:          []string{s.accounts[0].Addr.String()},
		Metadata:           "valid-metadata",
	}
	amountTransfered := sdk.NewInt64Coin("token", 100)
	s.Require().NoError(proposalReq.SetMsgs([]sdk.Msg{&banktypes.MsgSend{
		FromAddress: s.groupPolicyAddr.String(),
		ToAddress:   s.accounts[2].Addr.String(),
		Amount:      sdk.Coins{amountTransfered},
	}}))
	proposal, err := s.app.GroupKeeper.SubmitProposal(s.ctx, proposalReq)
	s.Require().NoError(err)

	timeoutHeight := s.sdkCtx.BlockHeight() + 10

	vote1 := &blsgroup.MsgVote{
		ProposalId:    proposal.ProposalId,
		Voter:         s.accounts[0].Addr.String(),
		Option:        group.VOTE_OPTION_YES,
		TimeoutHeight: timeoutHeight,
	}
	vote1Sig, err := s.accounts[0].PrivKey.Sign(vote1.GetSignBytes())
	s.Require().NoError(err)

	vote2 := &blsgroup.MsgVote{
		ProposalId:    proposal.ProposalId,
		Voter:         s.accounts[1].Addr.String(),
		Option:        group.VOTE_OPTION_YES,
		TimeoutHeight: timeoutHeight,
	}
	vote2Sig, err := s.accounts[1].PrivKey.Sign(vote2.GetSignBytes())
	s.Require().NoError(err)

	allVotes := []group.VoteOption{
		group.VOTE_OPTION_YES,
		group.VOTE_OPTION_YES,
		group.VOTE_OPTION_UNSPECIFIED,
	}

	aggSig, err := bls12381.AggregateSignature([][]byte{vote1Sig, vote2Sig})
	s.Require().NoError(err)

	beforeBalance := s.app.BankKeeper.GetBalance(s.sdkCtx, s.accounts[2].Addr, "token")

	_, err = s.app.BlsGroupKeeper.VoteAgg(s.ctx, &blsgroup.MsgVoteAgg{
		Sender:        s.accounts[0].Addr.String(),
		ProposalId:    proposal.ProposalId,
		Votes:         allVotes,
		AggSig:        aggSig,
		TimeoutHeight: timeoutHeight,
		Exec:          group.Exec_EXEC_TRY,
	})
	s.Require().NoError(err)

	// Match the Msg defined in the proposal (transfer 100token to account2)
	afterBalance := s.app.BankKeeper.GetBalance(s.sdkCtx, s.accounts[2].Addr, "token")
	s.Require().Equal(
		beforeBalance.Add(amountTransfered).Amount.Int64(),
		afterBalance.Amount.Int64(),
		"proposal execution must have transferred tokens",
	)

	// Successfully executed proposals are auto prunned right after execution
	_, err = s.app.GroupKeeper.Proposal(s.ctx, &group.QueryProposalRequest{ProposalId: proposal.ProposalId})
	s.Require().ErrorIs(err, sdkerrors.ErrNotFound)
}

func (s *TestSuite) TestVoteAggNoExecute() {
	proposalReq := &group.MsgSubmitProposal{
		GroupPolicyAddress: s.groupPolicyAddr.String(),
		Proposers:          []string{s.accounts[0].Addr.String()},
		Metadata:           "valid-metadata",
	}

	amountTransfered := sdk.NewInt64Coin("token", 100)
	s.Require().NoError(proposalReq.SetMsgs([]sdk.Msg{&banktypes.MsgSend{
		FromAddress: s.groupPolicyAddr.String(),
		ToAddress:   s.accounts[2].Addr.String(),
		Amount:      sdk.Coins{amountTransfered},
	}}))
	proposal, err := s.app.GroupKeeper.SubmitProposal(s.ctx, proposalReq)
	s.Require().NoError(err)

	timeoutHeight := s.sdkCtx.BlockHeight() + 10

	vote1 := &blsgroup.MsgVote{
		ProposalId:    proposal.ProposalId,
		Voter:         s.accounts[0].Addr.String(),
		Option:        group.VOTE_OPTION_YES,
		TimeoutHeight: timeoutHeight,
	}
	vote1Sig, err := s.accounts[0].PrivKey.Sign(vote1.GetSignBytes())
	s.Require().NoError(err)

	vote2 := &blsgroup.MsgVote{
		ProposalId:    proposal.ProposalId,
		Voter:         s.accounts[1].Addr.String(),
		Option:        group.VOTE_OPTION_YES,
		TimeoutHeight: timeoutHeight,
	}
	vote2Sig, err := s.accounts[1].PrivKey.Sign(vote2.GetSignBytes())
	s.Require().NoError(err)

	aggSig, err := bls12381.AggregateSignature([][]byte{vote1Sig})
	s.Require().NoError(err)

	_, err = s.app.BlsGroupKeeper.VoteAgg(s.ctx, &blsgroup.MsgVoteAgg{
		Sender:     s.accounts[0].Addr.String(),
		ProposalId: proposal.ProposalId,
		Votes: []group.VoteOption{
			group.VOTE_OPTION_YES,
			group.VOTE_OPTION_UNSPECIFIED,
			group.VOTE_OPTION_UNSPECIFIED,
		},
		AggSig:        aggSig,
		TimeoutHeight: timeoutHeight,
		Exec:          group.Exec_EXEC_UNSPECIFIED,
	})
	s.Require().NoError(err)

	propopsalResp, err := s.app.GroupKeeper.Proposal(s.ctx, &group.QueryProposalRequest{ProposalId: proposal.ProposalId})
	s.Require().NoError(err)
	s.Require().Equal(group.PROPOSAL_STATUS_SUBMITTED, propopsalResp.Proposal.Status)
	s.Require().Equal(group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN, propopsalResp.Proposal.ExecutorResult)

	votesResp, err := s.app.GroupKeeper.VotesByProposal(s.ctx, &group.QueryVotesByProposalRequest{ProposalId: proposal.ProposalId})
	s.Require().NoError(err)
	s.Require().Equal(1, len(votesResp.Votes))
	s.Require().Equal(group.VOTE_OPTION_YES, votesResp.Votes[0].Option)
	s.Require().Equal(vote1.Voter, votesResp.Votes[0].Voter)

	aggSig, err = bls12381.AggregateSignature([][]byte{vote2Sig})
	s.Require().NoError(err)

	_, err = s.app.BlsGroupKeeper.VoteAgg(s.ctx, &blsgroup.MsgVoteAgg{
		Sender:     s.accounts[0].Addr.String(),
		ProposalId: proposal.ProposalId,
		Votes: []group.VoteOption{
			group.VOTE_OPTION_UNSPECIFIED,
			group.VOTE_OPTION_YES,
			group.VOTE_OPTION_UNSPECIFIED,
		},
		AggSig:        aggSig,
		TimeoutHeight: timeoutHeight,
		Exec:          group.Exec_EXEC_UNSPECIFIED,
	})
	s.Require().NoError(err)

	propopsalResp, err = s.app.GroupKeeper.Proposal(s.ctx, &group.QueryProposalRequest{ProposalId: proposal.ProposalId})
	s.Require().NoError(err)
	s.Require().Equal(group.PROPOSAL_STATUS_SUBMITTED, propopsalResp.Proposal.Status)
	s.Require().Equal(group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN, propopsalResp.Proposal.ExecutorResult)

	votesResp, err = s.app.GroupKeeper.VotesByProposal(s.ctx, &group.QueryVotesByProposalRequest{ProposalId: proposal.ProposalId})
	s.Require().NoError(err)
	s.Require().Equal(2, len(votesResp.Votes))
	s.Require().Equal(group.VOTE_OPTION_YES, votesResp.Votes[0].Option)
	s.Require().Equal(group.VOTE_OPTION_YES, votesResp.Votes[1].Option)
	s.Require().Equal(vote1.Voter, votesResp.Votes[0].Voter)
	s.Require().Equal(vote2.Voter, votesResp.Votes[1].Voter)

	s.Require().NoError(s.app.GroupKeeper.TallyProposalsAtVPEnd(s.sdkCtx.WithBlockTime(s.sdkCtx.BlockTime().Add(s.policy.GetVotingPeriod() + 1))))

	proposalResp, err := s.app.GroupKeeper.Proposal(s.ctx, &group.QueryProposalRequest{ProposalId: proposal.ProposalId})
	s.Require().NoError(err)
	s.Require().Equal(group.PROPOSAL_STATUS_ACCEPTED, proposalResp.Proposal.Status)
	s.Require().Equal(group.PROPOSAL_EXECUTOR_RESULT_NOT_RUN, proposalResp.Proposal.ExecutorResult)
	s.Require().Equal(group.TallyResult{
		YesCount:        fmt.Sprintf("%d", s.accounts[0].Weight+s.accounts[1].Weight),
		AbstainCount:    "0",
		NoCount:         "0",
		NoWithVetoCount: "0",
	}, proposalResp.Proposal.FinalTallyResult)
}

func (s *TestSuite) TestVoteAggDuplicateVote() {
	proposalReq := &group.MsgSubmitProposal{
		GroupPolicyAddress: s.groupPolicyAddr.String(),
		Proposers:          []string{s.accounts[0].Addr.String()},
		Metadata:           "valid-metadata",
	}

	amountTransfered := sdk.NewInt64Coin("token", 100)
	s.Require().NoError(proposalReq.SetMsgs([]sdk.Msg{&banktypes.MsgSend{
		FromAddress: s.groupPolicyAddr.String(),
		ToAddress:   s.accounts[2].Addr.String(),
		Amount:      sdk.Coins{amountTransfered},
	}}))
	proposal, err := s.app.GroupKeeper.SubmitProposal(s.ctx, proposalReq)
	s.Require().NoError(err)

	timeoutHeight := s.sdkCtx.BlockHeight() + 10

	vote1 := &blsgroup.MsgVote{
		ProposalId:    proposal.ProposalId,
		Voter:         s.accounts[0].Addr.String(),
		Option:        group.VOTE_OPTION_YES,
		TimeoutHeight: timeoutHeight,
	}
	vote1Sig, err := s.accounts[0].PrivKey.Sign(vote1.GetSignBytes())
	s.Require().NoError(err)

	aggSig, err := bls12381.AggregateSignature([][]byte{vote1Sig})
	s.Require().NoError(err)

	_, err = s.app.BlsGroupKeeper.VoteAgg(s.ctx, &blsgroup.MsgVoteAgg{
		Sender:     s.accounts[0].Addr.String(),
		ProposalId: proposal.ProposalId,
		Votes: []group.VoteOption{
			group.VOTE_OPTION_YES,
			group.VOTE_OPTION_UNSPECIFIED,
			group.VOTE_OPTION_UNSPECIFIED,
		},
		AggSig:        aggSig,
		Exec:          group.Exec_EXEC_UNSPECIFIED,
		TimeoutHeight: timeoutHeight,
	})
	s.Require().NoError(err)

	newVote1 := &blsgroup.MsgVote{
		ProposalId:    proposal.ProposalId,
		Voter:         s.accounts[0].Addr.String(),
		Option:        group.VOTE_OPTION_NO,
		TimeoutHeight: timeoutHeight,
	}
	newVote1Sig, err := s.accounts[0].PrivKey.Sign(newVote1.GetSignBytes())
	s.Require().NoError(err)

	aggSig, err = bls12381.AggregateSignature([][]byte{newVote1Sig})
	s.Require().NoError(err)

	_, err = s.app.BlsGroupKeeper.VoteAgg(s.ctx, &blsgroup.MsgVoteAgg{
		Sender:     s.accounts[0].Addr.String(),
		ProposalId: proposal.ProposalId,
		Votes: []group.VoteOption{
			group.VOTE_OPTION_NO,
			group.VOTE_OPTION_UNSPECIFIED,
			group.VOTE_OPTION_UNSPECIFIED,
		},
		AggSig:        aggSig,
		TimeoutHeight: timeoutHeight,
		Exec:          group.Exec_EXEC_UNSPECIFIED,
	})
	s.Require().NoError(err)

	// No error, but original vote have not been modified
	votesResp, err := s.app.GroupKeeper.VotesByProposal(s.ctx, &group.QueryVotesByProposalRequest{ProposalId: proposal.ProposalId})
	s.Require().NoError(err)
	s.Require().Equal(1, len(votesResp.Votes))
	s.Require().Equal(group.VOTE_OPTION_YES, votesResp.Votes[0].Option)
}

func (s *TestSuite) TestVoteAggTimeout() {
	testCases := []struct {
		Description string
		VoteTimeout int64
		AggTimeout  int64
		ExpectErr   bool
		Err         error
	}{
		{
			Description: "timeout ok",
			VoteTimeout: s.sdkCtx.BlockHeight() + 1,
			AggTimeout:  s.sdkCtx.BlockHeight() + 1,
			ExpectErr:   false,
		},
		{
			Description: "expired vote",
			VoteTimeout: s.sdkCtx.BlockHeight() - 1,
			AggTimeout:  s.sdkCtx.BlockHeight() - 1,
			ExpectErr:   true,
			Err:         blsgroup.ErrExpired,
		},
		{
			Description: "invalid vote timeout",
			VoteTimeout: 0,
			AggTimeout:  s.sdkCtx.BlockHeight() + 1,
			ExpectErr:   true,
			Err:         blsgroup.ErrSignatureVerification,
		},
		{
			Description: "invalid aggregated vote timeout",
			VoteTimeout: s.sdkCtx.BlockHeight() + 1,
			AggTimeout:  0,
			ExpectErr:   true,
			Err:         blsgroup.ErrInvalid,
		},
		{
			Description: "timeout valid but mismatches",
			VoteTimeout: s.sdkCtx.BlockHeight() + 1,
			AggTimeout:  s.sdkCtx.BlockHeight() + 2,
			ExpectErr:   true,
			Err:         blsgroup.ErrSignatureVerification,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.Description, func() {
			proposalReq := &group.MsgSubmitProposal{
				GroupPolicyAddress: s.groupPolicyAddr.String(),
				Proposers:          []string{s.accounts[0].Addr.String()},
				Metadata:           "valid-metadata",
			}

			amountTransfered := sdk.NewInt64Coin("token", 100)
			s.Require().NoError(proposalReq.SetMsgs([]sdk.Msg{&banktypes.MsgSend{
				FromAddress: s.groupPolicyAddr.String(),
				ToAddress:   s.accounts[2].Addr.String(),
				Amount:      sdk.Coins{amountTransfered},
			}}))
			proposal, err := s.app.GroupKeeper.SubmitProposal(s.ctx, proposalReq)
			s.Require().NoError(err)

			vote1 := &blsgroup.MsgVote{
				ProposalId:    proposal.ProposalId,
				Voter:         s.accounts[0].Addr.String(),
				Option:        group.VOTE_OPTION_YES,
				TimeoutHeight: tc.VoteTimeout,
			}
			vote1Sig, err := s.accounts[0].PrivKey.Sign(vote1.GetSignBytes())
			s.Require().NoError(err)

			aggSig, err := bls12381.AggregateSignature([][]byte{vote1Sig})
			s.Require().NoError(err)

			_, err = s.app.BlsGroupKeeper.VoteAgg(s.ctx, &blsgroup.MsgVoteAgg{
				Sender:     s.accounts[0].Addr.String(),
				ProposalId: proposal.ProposalId,
				Votes: []group.VoteOption{
					group.VOTE_OPTION_YES,
					group.VOTE_OPTION_UNSPECIFIED,
					group.VOTE_OPTION_UNSPECIFIED,
				},
				AggSig:        aggSig,
				Exec:          group.Exec_EXEC_UNSPECIFIED,
				TimeoutHeight: tc.AggTimeout,
			})

			if tc.ExpectErr {
				if tc.Err != nil {
					s.Require().ErrorIs(err, tc.Err)
				} else {
					s.Require().Error(err)
				}
			} else {
				s.Require().NoError(err)
			}
		})
	}

}

func (s *TestSuite) TestGetAllGroupMembers() {
	members := make([]group.MemberRequest, 67)
	for i := 0; i < len(members); i++ {
		_, _, addr := testdata.KeyTestPubAddrBls12381()
		members[i] = group.MemberRequest{
			Address: addr.String(),
			Weight:  "2",
		}
	}

	groupRes, err := s.app.GroupKeeper.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin:   s.groupAdmin.String(),
		Members: members,
	})
	s.Require().NoError(err)

	respMembers, err := s.app.BlsGroupKeeper.GetAllGroupMembers(s.ctx, groupRes.GroupId)
	s.Require().NoError(err)
	s.Require().Equal(len(members), len(respMembers))
}
