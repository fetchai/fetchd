package keeper_test

import (
	"bytes"
	"context"
	"sort"
	"testing"
	"time"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/stretchr/testify/suite"
	tmtime "github.com/tendermint/tendermint/libs/time"
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

	s.blockTime = tmtime.Now()
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
	members := []group.Member{
		{Address: s.accounts[0].Addr.String(), Weight: "1"},
		{Address: s.accounts[1].Addr.String(), Weight: "2"},
		{Address: s.accounts[2].Addr.String(), Weight: "2"},
	}

	groupRes, err := s.app.GroupKeeper.CreateGroup(s.ctx, &group.MsgCreateGroup{
		Admin:   s.groupAdmin.String(),
		Members: members,
	})
	s.Require().NoError(err)
	s.groupID = groupRes.GroupId

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

func (s *TestSuite) TestVoteAgg() {
	proposalReq := &group.MsgSubmitProposal{
		Address:   s.groupPolicyAddr.String(),
		Proposers: []string{s.accounts[0].Addr.String()},
		Metadata:  "valid-metadata",
	}
	s.Require().NoError(proposalReq.SetMsgs([]sdk.Msg{&banktypes.MsgSend{
		FromAddress: s.groupPolicyAddr.String(),
		ToAddress:   s.accounts[2].Addr.String(),

		Amount: sdk.Coins{sdk.NewInt64Coin("token", 100)},
	}}))
	proposal, err := s.app.GroupKeeper.SubmitProposal(s.ctx, proposalReq)
	s.Require().NoError(err)

	vote1 := &group.MsgVote{
		ProposalId: proposal.ProposalId,
		Voter:      s.accounts[0].Addr.String(),
		Option:     group.VOTE_OPTION_YES,
	}
	vote1Sig, err := s.accounts[0].PrivKey.Sign(vote1.GetSignBytes())
	s.Require().NoError(err)

	vote2 := &group.MsgVote{
		ProposalId: proposal.ProposalId,
		Voter:      s.accounts[1].Addr.String(),
		Option:     group.VOTE_OPTION_YES,
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

	_, err = s.app.BlsGroupKeeper.VoteAgg(s.ctx, &blsgroup.MsgVoteAgg{
		Sender:     s.accounts[0].Addr.String(),
		ProposalId: proposal.ProposalId,
		Votes:      allVotes,
		AggSig:     aggSig,
		Exec:       group.Exec_EXEC_UNSPECIFIED,
	})
	s.Require().NoError(err)
}
