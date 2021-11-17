package group

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/gogo/protobuf/types"
	proto "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMsgCreateGroupValidation(t *testing.T) {
	_, _, myAddr := testdata.KeyTestPubAddr()
	_, _, myOtherAddr := testdata.KeyTestPubAddr()

	specs := map[string]struct {
		src    MsgCreateGroup
		expErr bool
	}{
		"all good with minimum fields set": {
			src: MsgCreateGroup{Admin: myAddr.String()},
		},
		"all good with a member": {
			src: MsgCreateGroup{
				Admin: myAddr.String(),
				Members: []Member{
					{Address: myAddr.String(), Weight: "1"},
				},
			},
		},
		"all good with multiple members": {
			src: MsgCreateGroup{
				Admin: myAddr.String(),
				Members: []Member{
					{Address: myAddr.String(), Weight: "1"},
					{Address: myOtherAddr.String(), Weight: "2"},
				},
			},
		},
		"admin required": {
			src:    MsgCreateGroup{},
			expErr: true,
		},
		"valid admin required": {
			src: MsgCreateGroup{
				Admin: "invalid-address",
			},
			expErr: true,
		},
		"duplicate member addresses not allowed": {
			src: MsgCreateGroup{
				Admin: myAddr.String(),
				Members: []Member{
					{Address: myAddr.String(), Weight: "1"},
					{Address: myAddr.String(), Weight: "2"},
				},
			},
			expErr: true,
		},
		"negative member's weight not allowed": {
			src: MsgCreateGroup{
				Admin: myAddr.String(),
				Members: []Member{
					{Address: myAddr.String(), Weight: "-1"},
				},
			},
			expErr: true,
		},
		"empty member's weight not allowed": {
			src: MsgCreateGroup{
				Admin:   myAddr.String(),
				Members: []Member{{Address: myAddr.String()}},
			},
			expErr: true,
		},
		"zero member's weight not allowed": {
			src: MsgCreateGroup{
				Admin:   myAddr.String(),
				Members: []Member{{Address: myAddr.String(), Weight: "0"}},
			},
			expErr: true,
		},
		"member address required": {
			src: MsgCreateGroup{
				Admin: myAddr.String(),
				Members: []Member{
					{Weight: "1"},
				},
			},
			expErr: true,
		},
		"valid member address required": {
			src: MsgCreateGroup{
				Admin: myAddr.String(),
				Members: []Member{
					{Address: "invalid-address", Weight: "1"},
				},
			},
			expErr: true,
		},
		"group metadata too long": {
			src: MsgCreateGroup{
				Admin:    myAddr.String(),
				Metadata: bytes.Repeat([]byte{1}, 256),
			},
			expErr: true,
		},
		"members metadata too long": {
			src: MsgCreateGroup{
				Admin: myAddr.String(),
				Members: []Member{
					{Address: myAddr.String(), Weight: "1", Metadata: bytes.Repeat([]byte{1}, 256)},
				},
			},
			expErr: true,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			err := spec.src.ValidateBasic()
			if spec.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgCreateGroupSigner(t *testing.T) {
	_, _, myAddr := testdata.KeyTestPubAddr()
	assert.Equal(t, []sdk.AccAddress{myAddr}, MsgCreateGroup{Admin: myAddr.String()}.GetSigners())
}

func TestMsgCreateGroupAccount(t *testing.T) {
	_, _, myAddr := testdata.KeyTestPubAddr()

	specs := map[string]struct {
		admin     sdk.AccAddress
		group     uint64
		metadata  []byte
		threshold string
		timeout   proto.Duration
		expErr    bool
	}{
		"all good with minimum fields set": {
			admin:     myAddr,
			group:     1,
			threshold: "1",
			timeout:   proto.Duration{Seconds: 1},
		},
		"zero threshold not allowed": {
			admin:     myAddr,
			group:     1,
			threshold: "0",
			timeout:   proto.Duration{Seconds: 1},
			expErr:    true,
		},
		"admin required": {
			group:     1,
			threshold: "1",
			timeout:   proto.Duration{Seconds: 1},
			expErr:    true,
		},
		"group required": {
			admin:     myAddr,
			threshold: "1",
			timeout:   proto.Duration{Seconds: 1},
			expErr:    true,
		},
		"decision policy required": {
			admin:  myAddr,
			group:  1,
			expErr: true,
		},
		"decision policy without timeout": {
			admin:     myAddr,
			group:     1,
			threshold: "1",
			expErr:    true,
		},
		"decision policy with invalid timeout": {
			admin:     myAddr,
			group:     1,
			threshold: "1",
			timeout:   proto.Duration{Seconds: -1},
			expErr:    true,
		},
		"decision policy without threshold": {
			admin:   myAddr,
			group:   1,
			timeout: proto.Duration{Seconds: 1},
			expErr:  true,
		},
		"decision policy with negative threshold": {
			admin:     myAddr,
			group:     1,
			threshold: "-1",
			timeout:   proto.Duration{Seconds: 1},
			expErr:    true,
		},
		"metadata too long": {
			admin:     myAddr,
			group:     1,
			metadata:  []byte(strings.Repeat("a", 256)),
			threshold: "1",
			timeout:   proto.Duration{Seconds: 1},
			expErr:    true,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			m, err := NewMsgCreateGroupAccount(
				spec.admin,
				spec.group,
				spec.metadata,
				&ThresholdDecisionPolicy{
					Threshold: spec.threshold,
					Timeout:   spec.timeout,
				},
			)
			require.NoError(t, err)

			if spec.expErr {
				require.Error(t, m.ValidateBasic())
			} else {
				require.NoError(t, m.ValidateBasic())
			}
		})
	}
}

func TestMsgCreateProposalRequest(t *testing.T) {
	_, _, addr := testdata.KeyTestPubAddr()
	groupAccAddr := addr.String()

	_, _, addr = testdata.KeyTestPubAddr()
	memberAddr := addr.String()

	specs := map[string]struct {
		src    MsgCreateProposal
		expErr bool
	}{
		"all good with minimum fields set": {
			src: MsgCreateProposal{
				Address:   groupAccAddr,
				Proposers: []string{memberAddr},
			},
		},
		"group account required": {
			src: MsgCreateProposal{
				Proposers: []string{memberAddr},
			},
			expErr: true,
		},
		"proposers required": {
			src: MsgCreateProposal{
				Address: groupAccAddr,
			},
			expErr: true,
		},
		"valid proposer address required": {
			src: MsgCreateProposal{
				Address:   groupAccAddr,
				Proposers: []string{"invalid-member-address"},
			},
			expErr: true,
		},
		"no duplicate proposers": {
			src: MsgCreateProposal{
				Address:   groupAccAddr,
				Proposers: []string{memberAddr, memberAddr},
			},
			expErr: true,
		},
		"empty proposer address not allowed": {
			src: MsgCreateProposal{
				Address:   groupAccAddr,
				Proposers: []string{memberAddr, ""},
			},
			expErr: true,
		},
		"metadata too long": {
			src: MsgCreateProposal{
				Address:   groupAccAddr,
				Proposers: []string{memberAddr},
				Metadata:  bytes.Repeat([]byte{1}, 256),
			},
			expErr: true,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			err := spec.src.ValidateBasic()
			if spec.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgVote(t *testing.T) {
	_, _, addr := testdata.KeyTestPubAddr()
	memberAddr := addr.String()

	specs := map[string]struct {
		src    MsgVote
		expErr bool
	}{
		"all good with minimum fields set": {
			src: MsgVote{
				ProposalId: 1,
				Choice:     Choice_CHOICE_YES,
				Voter:      memberAddr,
			},
		},
		"proposal required": {
			src: MsgVote{
				Choice: Choice_CHOICE_YES,
				Voter:  memberAddr,
			},
			expErr: true,
		},
		"choice required": {
			src: MsgVote{
				ProposalId: 1,
				Voter:      memberAddr,
			},
			expErr: true,
		},
		"valid choice required": {
			src: MsgVote{
				ProposalId: 1,
				Choice:     5,
				Voter:      memberAddr,
			},
			expErr: true,
		},
		"voter required": {
			src: MsgVote{
				ProposalId: 1,
				Choice:     Choice_CHOICE_YES,
			},
			expErr: true,
		},
		"valid voter address required": {
			src: MsgVote{
				ProposalId: 1,
				Choice:     Choice_CHOICE_YES,
				Voter:      "invalid-member-address",
			},
			expErr: true,
		},
		"empty voters address not allowed": {
			src: MsgVote{
				ProposalId: 1,
				Choice:     Choice_CHOICE_YES,
				Voter:      "",
			},
			expErr: true,
		},
		"metadata too long": {
			src: MsgVote{
				ProposalId: 1,
				Choice:     Choice_CHOICE_YES,
				Voter:      memberAddr,
				Metadata:   bytes.Repeat([]byte{1}, 256),
			},
			expErr: true,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			err := spec.src.ValidateBasic()
			if spec.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgVoteBasicResponse(t *testing.T) {
	sk, pk, addr := testdata.KeyTestPubAddrBls12381()
	memberAddr := addr.String()

	sk2, pk2, addr2 := testdata.KeyTestPubAddrBls12381()

	now := time.Now()
	expiry, err := gogotypes.TimestampProto(now.Add(time.Second * 3000))
	require.NoError(t, err)

	msg := MsgVoteBasic{
		ProposalId: 1,
		Choice:     Choice_CHOICE_YES,
		Expiry:     *expiry,
	}
	err = msg.ValidateBasic()
	require.NoError(t, err)

	bytesToSign := msg.GetSignBytes()

	sig, err := sk.Sign(bytesToSign)
	require.NoError(t, err)

	sig2, err := sk2.Sign(bytesToSign)
	require.NoError(t, err)

	pubKeyAny, err := codectypes.NewAnyWithValue(pk)
	require.NoError(t, err)

	pubKeyAny2, err := codectypes.NewAnyWithValue(pk2)
	require.NoError(t, err)

	specs := map[string]struct {
		src    MsgVoteBasicResponse
		expErr bool
		sigErr bool
	}{
		"all good": {
			src: MsgVoteBasicResponse{
				ProposalId: 1,
				Choice:     Choice_CHOICE_YES,
				Expiry:     *expiry,
				Voter:      memberAddr,
				PubKey:     pubKeyAny,
				Sig:        sig,
			},
		},
		"proposal required": {
			src: MsgVoteBasicResponse{
				Choice: Choice_CHOICE_YES,
				Expiry: *expiry,
				Voter:  memberAddr,
				PubKey: pubKeyAny,
				Sig:    sig,
			},
			expErr: true,
		},
		"choice required": {
			src: MsgVoteBasicResponse{
				ProposalId: 1,
				Expiry:     *expiry,
				Voter:      memberAddr,
				PubKey:     pubKeyAny,
				Sig:        sig,
			},
			expErr: true,
		},
		"valid choice required": {
			src: MsgVoteBasicResponse{
				ProposalId: 1,
				Choice:     5,
				Expiry:     *expiry,
				Voter:      memberAddr,
				PubKey:     pubKeyAny,
				Sig:        sig,
			},
			expErr: true,
		},
		"voter required": {
			src: MsgVoteBasicResponse{
				ProposalId: 1,
				Choice:     Choice_CHOICE_YES,
				Expiry:     *expiry,
				PubKey:     pubKeyAny,
				Sig:        sig,
			},
			expErr: true,
		},
		"valid voter address required": {
			src: MsgVoteBasicResponse{
				ProposalId: 1,
				Choice:     Choice_CHOICE_YES,
				Expiry:     *expiry,
				Voter:      "invalid member address",
				PubKey:     pubKeyAny,
				Sig:        sig,
			},
			expErr: true,
		},
		"wrong voter": {
			src: MsgVoteBasicResponse{
				ProposalId: 1,
				Choice:     Choice_CHOICE_YES,
				Expiry:     *expiry,
				Voter:      addr2.String(),
				PubKey:     pubKeyAny,
				Sig:        sig,
			},
			sigErr: true,
		},
		"wrong choice": {
			src: MsgVoteBasicResponse{
				ProposalId: 1,
				Choice:     Choice_CHOICE_NO,
				Expiry:     *expiry,
				Voter:      memberAddr,
				PubKey:     pubKeyAny,
				Sig:        sig,
			},
			sigErr: true,
		},
		"wrong signature": {
			src: MsgVoteBasicResponse{
				ProposalId: 1,
				Choice:     Choice_CHOICE_YES,
				Expiry:     *expiry,
				Voter:      addr.String(),
				PubKey:     pubKeyAny,
				Sig:        sig2,
			},
			sigErr: true,
		},
		"empty public key": {
			src: MsgVoteBasicResponse{
				ProposalId: 1,
				Choice:     Choice_CHOICE_YES,
				Expiry:     *expiry,
				Voter:      memberAddr,
				Sig:        sig,
			},
			expErr: true,
		},
		"wrong public key": {
			src: MsgVoteBasicResponse{
				ProposalId: 1,
				Choice:     Choice_CHOICE_YES,
				Expiry:     *expiry,
				Voter:      memberAddr,
				PubKey:     pubKeyAny2,
				Sig:        sig,
			},
			sigErr: true,
		},
		"empty voters address not allowed": {
			src: MsgVoteBasicResponse{
				ProposalId: 1,
				Choice:     Choice_CHOICE_YES,
				Expiry:     *expiry,
				PubKey:     pubKeyAny,
				Sig:        sig,
			},
			expErr: true,
		},
		"empty signature": {
			src: MsgVoteBasicResponse{
				ProposalId: 1,
				Choice:     Choice_CHOICE_YES,
				Expiry:     *expiry,
				Voter:      memberAddr,
				PubKey:     pubKeyAny,
			},
			expErr: true,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			err := spec.src.ValidateBasic()
			if spec.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				err = spec.src.VerifySignature()
				if spec.sigErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			}
		})
	}
}

func TestMsgVoteAggRequest(t *testing.T) {
	_, _, addr := testdata.KeyTestPubAddrBls12381()
	memberAddr := addr.String()

	now := time.Now()
	expiry, err := gogotypes.TimestampProto(now.Add(time.Second * 3000))
	require.NoError(t, err)

	specs := map[string]struct {
		src    MsgVoteAgg
		expErr bool
	}{
		"all good with minimus fields": {
			src: MsgVoteAgg{
				Sender:     memberAddr,
				ProposalId: 1,
				Votes:      []Choice{Choice_CHOICE_YES, Choice_CHOICE_NO},
				Expiry:     *expiry,
				AggSig:     []byte("does not check signature"),
			},
		},
		"proposal required": {
			src: MsgVoteAgg{
				Sender: memberAddr,
				Votes:  []Choice{Choice_CHOICE_YES, Choice_CHOICE_NO},
				Expiry: *expiry,
				AggSig: []byte("does not check signature"),
			},
			expErr: true,
		},
		"votes required": {
			src: MsgVoteAgg{
				Sender:     memberAddr,
				ProposalId: 1,
				Expiry:     *expiry,
				AggSig:     []byte("does not check signature"),
			},
			expErr: true,
		},
		"valid votes required": {
			src: MsgVoteAgg{
				Sender:     memberAddr,
				ProposalId: 1,
				Votes:      []Choice{5, Choice_CHOICE_NO},
				Expiry:     *expiry,
				AggSig:     []byte("does not check signature"),
			},
			expErr: true,
		},
		"sender required": {
			src: MsgVoteAgg{
				ProposalId: 1,
				Votes:      []Choice{Choice_CHOICE_YES, Choice_CHOICE_NO},
				Expiry:     *expiry,
				AggSig:     []byte("does not check signature"),
			},
			expErr: true,
		},
		"valid sender address required": {
			src: MsgVoteAgg{
				Sender:     "invalid sender address",
				ProposalId: 1,
				Votes:      []Choice{Choice_CHOICE_YES, Choice_CHOICE_NO},
				Expiry:     *expiry,
				AggSig:     []byte("does not check signature"),
			},
			expErr: true,
		},
		"empty signature": {
			src: MsgVoteAgg{
				Sender:     memberAddr,
				ProposalId: 1,
				Votes:      []Choice{Choice_CHOICE_YES, Choice_CHOICE_NO},
				Expiry:     *expiry,
			},
			expErr: true,
		},
		"metadata too long": {
			src: MsgVoteAgg{
				Sender:     memberAddr,
				ProposalId: 1,
				Votes:      []Choice{Choice_CHOICE_YES, Choice_CHOICE_NO},
				Expiry:     *expiry,
				AggSig:     []byte("does not check signature"),
				Metadata:   bytes.Repeat([]byte{1}, 256),
			},
			expErr: true,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			err := spec.src.ValidateBasic()
			if spec.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgCreatePollRequest(t *testing.T) {
	_, _, addr := testdata.KeyTestPubAddr()
	memberAddr := addr.String()

	now := time.Now()
	expiry, err := gogotypes.TimestampProto(now.Add(time.Second * 3000))
	require.NoError(t, err)

	longTitle := strings.Repeat("my title", 256)
	manyOptions := make([]string, 300)
	for i, _ := range manyOptions {
		manyOptions[i] = fmt.Sprintf("option-%d", i)
	}

	specs := map[string]struct {
		src    MsgCreatePoll
		expErr bool
	}{
		"all good": {
			src: MsgCreatePoll{
				GroupId:   1,
				Title:     "2022 Election",
				Options:   Options{Titles: []string{"alice", "bob"}},
				Creator:   memberAddr,
				VoteLimit: 1,
				Timeout:   *expiry,
			},
		},
		"group id required": {
			src: MsgCreatePoll{
				Title:     "2022 Election",
				Options:   Options{Titles: []string{"alice", "bob"}},
				Creator:   memberAddr,
				VoteLimit: 1,
				Timeout:   *expiry,
			},
			expErr: true,
		},
		"Creator required": {
			src: MsgCreatePoll{
				GroupId:   1,
				Title:     "2022 Election",
				Options:   Options{Titles: []string{"alice", "bob"}},
				VoteLimit: 1,
				Timeout:   *expiry,
			},
			expErr: true,
		},
		"valid creator address required": {
			src: MsgCreatePoll{
				GroupId:   1,
				Title:     "2022 Election",
				Options:   Options{Titles: []string{"alice", "bob"}},
				Creator:   "invalid member address",
				VoteLimit: 1,
				Timeout:   *expiry,
			},
			expErr: true,
		},
		"title required": {
			src: MsgCreatePoll{
				GroupId:   1,
				Options:   Options{Titles: []string{"alice", "bob"}},
				Creator:   memberAddr,
				VoteLimit: 1,
				Timeout:   *expiry,
			},
			expErr: true,
		},
		"options required": {
			src: MsgCreatePoll{
				GroupId:   1,
				Title:     "2022 Election",
				Creator:   memberAddr,
				VoteLimit: 1,
				Timeout:   *expiry,
			},
			expErr: true,
		},
		"empty option": {
			src: MsgCreatePoll{
				GroupId:   1,
				Title:     "2022 Election",
				Options:   Options{Titles: []string{"alice", ""}},
				Creator:   memberAddr,
				VoteLimit: 1,
				Timeout:   *expiry,
			},
			expErr: true,
		},
		"repeated options": {
			src: MsgCreatePoll{
				GroupId:   1,
				Title:     "2022 Election",
				Options:   Options{Titles: []string{"alice", "alice"}},
				Creator:   memberAddr,
				VoteLimit: 1,
				Timeout:   *expiry,
			},
			expErr: true,
		},
		"vote limit required": {
			src: MsgCreatePoll{
				GroupId: 1,
				Title:   "2022 Election",
				Options: Options{Titles: []string{"alice", "bob"}},
				Creator: memberAddr,
				Timeout: *expiry,
			},
			expErr: true,
		},
		"vote limit exceed": {
			src: MsgCreatePoll{
				GroupId:   1,
				Title:     "2022 Election",
				Options:   Options{Titles: []string{"alice", "bob"}},
				Creator:   memberAddr,
				VoteLimit: 3,
				Timeout:   *expiry,
			},
			expErr: true,
		},
		"metadata too long": {
			src: MsgCreatePoll{
				GroupId:   1,
				Title:     "2022 Election",
				Options:   Options{Titles: []string{"alice", "bob"}},
				Creator:   memberAddr,
				VoteLimit: 1,
				Metadata:  bytes.Repeat([]byte{1}, 256),
				Timeout:   *expiry,
			},
			expErr: true,
		},
		"title too long": {
			src: MsgCreatePoll{
				GroupId:   1,
				Title:     longTitle,
				Options:   Options{Titles: []string{"alice", "bob"}},
				Creator:   memberAddr,
				VoteLimit: 1,
				Timeout:   *expiry,
			},
			expErr: true,
		},
		"option title too long": {
			src: MsgCreatePoll{
				GroupId:   1,
				Title:     "2022 Election",
				Options:   Options{Titles: []string{"alice", longTitle}},
				Creator:   memberAddr,
				VoteLimit: 1,
				Timeout:   *expiry,
			},
			expErr: true,
		},
		"too many options": {
			src: MsgCreatePoll{
				GroupId:   1,
				Title:     "2022 Election",
				Options:   Options{Titles: manyOptions},
				Creator:   memberAddr,
				VoteLimit: 2,
				Timeout:   *expiry,
			},
			expErr: true,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			err := spec.src.ValidateBasic()
			if spec.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgVotePoll(t *testing.T) {
	_, _, addr := testdata.KeyTestPubAddr()
	memberAddr := addr.String()

	specs := map[string]struct {
		src    MsgVotePoll
		expErr bool
	}{
		"all good with minimum fields set": {
			src: MsgVotePoll{
				PollId:  1,
				Voter:   memberAddr,
				Options: Options{Titles: []string{"alice"}},
			},
		},
		"poll required": {
			src: MsgVotePoll{
				Voter:   memberAddr,
				Options: Options{Titles: []string{"alice"}},
			},
			expErr: true,
		},
		"options required": {
			src: MsgVotePoll{
				PollId: 1,
				Voter:  memberAddr,
			},
			expErr: true,
		},
		"valid options required": {
			src: MsgVotePoll{
				PollId:  1,
				Voter:   memberAddr,
				Options: Options{Titles: []string{"alice", ""}},
			},
			expErr: true,
		},
		"voter required": {
			src: MsgVotePoll{
				PollId:  1,
				Options: Options{Titles: []string{"alice"}},
			},
			expErr: true,
		},
		"valid voter address required": {
			src: MsgVotePoll{
				PollId:  1,
				Voter:   "invalid member address",
				Options: Options{Titles: []string{"alice"}},
			},
			expErr: true,
		},
		"empty voters address not allowed": {
			src: MsgVotePoll{
				PollId:  1,
				Voter:   "",
				Options: Options{Titles: []string{"alice"}},
			},
			expErr: true,
		},
		"metadata too long": {
			src: MsgVotePoll{
				PollId:   1,
				Voter:    memberAddr,
				Options:  Options{Titles: []string{"alice"}},
				Metadata: bytes.Repeat([]byte{1}, 256),
			},
			expErr: true,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			err := spec.src.ValidateBasic()
			if spec.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgVotePollAggRequest(t *testing.T) {
	_, _, addr := testdata.KeyTestPubAddrBls12381()
	memberAddr := addr.String()

	now := time.Now()
	expiry, err := gogotypes.TimestampProto(now.Add(time.Second * 3000))
	require.NoError(t, err)

	specs := map[string]struct {
		src    MsgVotePollAgg
		expErr bool
	}{
		"all good with minimus fields": {
			src: MsgVotePollAgg{
				Sender: memberAddr,
				PollId: 1,
				Votes:  []Options{Options{Titles: []string{"alice", "bob"}}, Options{Titles: []string{"alice"}}},
				Expiry: *expiry,
				AggSig: []byte("does not check signature"),
			},
		},
		"poll required": {
			src: MsgVotePollAgg{
				Sender: memberAddr,
				Votes:  []Options{Options{Titles: []string{"alice", "bob"}}, Options{Titles: []string{"alice"}}},
				Expiry: *expiry,
				AggSig: []byte("does not check signature"),
			},
			expErr: true,
		},
		"votes required": {
			src: MsgVotePollAgg{
				Sender: memberAddr,
				PollId: 1,
				Expiry: *expiry,
				AggSig: []byte("does not check signature"),
			},
			expErr: true,
		},
		"valid votes required": {
			src: MsgVotePollAgg{
				Sender: memberAddr,
				PollId: 1,
				Votes:  []Options{Options{Titles: []string{"alice", ""}}, Options{Titles: []string{"alice"}}},
				Expiry: *expiry,
				AggSig: []byte("does not check signature"),
			},
			expErr: true,
		},
		"sender required": {
			src: MsgVotePollAgg{
				PollId: 1,
				Votes:  []Options{Options{Titles: []string{"alice", "bob"}}, Options{Titles: []string{"alice"}}},
				Expiry: *expiry,
				AggSig: []byte("does not check signature"),
			},
			expErr: true,
		},
		"valid sender address required": {
			src: MsgVotePollAgg{
				Sender: "invalid sender address",
				PollId: 1,
				Votes:  []Options{Options{Titles: []string{"alice", "bob"}}, Options{Titles: []string{"alice"}}},
				Expiry: *expiry,
				AggSig: []byte("does not check signature"),
			},
			expErr: true,
		},
		"empty signature": {
			src: MsgVotePollAgg{
				Sender: memberAddr,
				PollId: 1,
				Votes:  []Options{Options{Titles: []string{"alice", "bob"}}, Options{Titles: []string{"alice"}}},
				Expiry: *expiry,
			},
			expErr: true,
		},
		"metadata too long": {
			src: MsgVotePollAgg{
				Sender:   memberAddr,
				PollId:   1,
				Votes:    []Options{Options{Titles: []string{"alice", "bob"}}, Options{Titles: []string{"alice"}}},
				Expiry:   *expiry,
				AggSig:   []byte("does not check signature"),
				Metadata: bytes.Repeat([]byte{1}, 256),
			},
			expErr: true,
		},
	}
	for msg, spec := range specs {
		t.Run(msg, func(t *testing.T) {
			err := spec.src.ValidateBasic()
			if spec.expErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
