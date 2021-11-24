package testsuite

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	"github.com/gogo/protobuf/proto"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/fetchai/fetchd/types/testutil/cli"
	"github.com/fetchai/fetchd/types/testutil/network"
	"github.com/fetchai/fetchd/x/group"
	"github.com/fetchai/fetchd/x/group/client"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network

	group         *group.GroupInfo
	groupAccounts []*group.GroupAccountInfo

	groupBls         *group.GroupInfo
	groupAccountsBls []*group.GroupAccountInfo

	proposal *group.Proposal
	vote     *group.Vote

	proposalBls *group.Proposal
	pollBls     *group.Poll
	votePoll    *group.VotePoll
}

const validMetadata = "AQ=="

func NewIntegrationTestSuite(cfg network.Config) *IntegrationTestSuite {
	return &IntegrationTestSuite{cfg: cfg}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	// We execute NewIntegrationTestSuite to set cfg field of IntegrationTestSuite
	s.cfg.NumValidators = 2
	s.network = network.New(s.T(), s.cfg)

	_, err := s.network.WaitForHeight(1)
	s.Require().NoError(err)

	val := s.network.Validators[0]

	// create a new account
	info, _, err := val.ClientCtx.Keyring.NewMnemonic("NewValidator", keyring.English, sdk.FullFundraiserPath, hd.Secp256k1)
	s.Require().NoError(err)

	account := sdk.AccAddress(info.GetPubKey().Address())
	out, err := banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		account,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(2000))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	)
	s.Require().NoError(err, out.String())

	// Create new bls account in the keyring.
	aliceInfo, _, err := val.ClientCtx.Keyring.NewMnemonic("alice", keyring.English, sdk.FullFundraiserPath, hd.Bls12381)
	s.Require().NoError(err)
	aliceAddr := sdk.AccAddress(aliceInfo.GetPubKey().Address())
	out, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		aliceAddr,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(2000))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	)
	s.Require().NoError(err, out.String())

	// perform a transaction to set pop for bls public key
	bobInfo, _, err := val.ClientCtx.Keyring.NewMnemonic("bob", keyring.English, sdk.FullFundraiserPath, hd.Bls12381)
	s.Require().NoError(err)
	bobAddr := sdk.AccAddress(bobInfo.GetPubKey().Address())
	out, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		aliceAddr,
		bobAddr,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(1000))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	)
	s.Require().NoError(err, out.String())

	charlieInfo, _, err := val.ClientCtx.Keyring.NewMnemonic("charlie", keyring.English, sdk.FullFundraiserPath, hd.Bls12381)
	s.Require().NoError(err)
	charlieAddr := sdk.AccAddress(charlieInfo.GetPubKey().Address())
	out, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		bobAddr,
		charlieAddr,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(800))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	)
	s.Require().NoError(err, out.String())

	out, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		charlieAddr,
		bobAddr,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(300))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	)
	s.Require().NoError(err, out.String())

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	// create a group
	validMembers := fmt.Sprintf(`{"members":[{"address":"%s","weight":"3","metadata":"%s"}]}`, val.Address.String(), validMetadata)
	validMembersFile := testutil.WriteToNewTempFile(s.T(), validMembers)
	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.MsgCreateGroupCmd(),
		append(
			[]string{
				val.Address.String(),
				validMetadata,
				validMembersFile.Name(),
			},
			commonFlags...,
		),
	)

	s.Require().NoError(err, out.String())
	var txResp = sdk.TxResponse{}
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())

	s.group = &group.GroupInfo{GroupId: 1, Admin: val.Address.String(), Metadata: []byte{1}, TotalWeight: "3", Version: 1}

	// create 5 group accounts
	for i := 0; i < 5; i++ {
		threshold := i + 1
		if threshold > 3 {
			threshold = 3
		}
		out, err = cli.ExecTestCLICmd(val.ClientCtx, client.MsgCreateGroupAccountCmd(),
			append(
				[]string{
					val.Address.String(),
					"1",
					validMetadata,
					fmt.Sprintf("{\"@type\":\"/fetchai.group.v1alpha1.ThresholdDecisionPolicy\", \"threshold\":\"%d\", \"timeout\":\"30000s\"}", threshold),
				},
				commonFlags...,
			),
		)
		s.Require().NoError(err, out.String())
		s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &txResp), out.String())
		s.Require().Equal(uint32(0), txResp.Code, out.String())

		out, err = cli.ExecTestCLICmd(val.ClientCtx, client.QueryGroupAccountsByGroupCmd(), []string{"1", fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
		s.Require().NoError(err, out.String())
	}

	var res group.QueryGroupAccountsByGroupResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &res))
	s.Require().Equal(len(res.GroupAccounts), 5)
	s.groupAccounts = res.GroupAccounts

	// create a proposal
	validTxFileName := getTxSendFileName(s, s.groupAccounts[0].Address, val.Address.String())
	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.MsgCreateProposalCmd(),
		append(
			[]string{
				s.groupAccounts[0].Address,
				val.Address.String(),
				validTxFileName,
				"",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
			},
			commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())

	// vote
	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.MsgVoteCmd(),
		append(
			[]string{
				"1",
				val.Address.String(),
				"CHOICE_YES",
				"",
			},
			commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())

	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.QueryProposalCmd(), []string{"1", fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err, out.String())

	var proposalRes group.QueryProposalResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &proposalRes))
	s.proposal = proposalRes.Proposal

	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.QueryVoteByProposalVoterCmd(), []string{"1", val.Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err, out.String())

	var voteRes group.QueryVoteByProposalVoterResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &voteRes))
	s.vote = voteRes.Vote

	// Create a bls-only group
	validBlsMembers := fmt.Sprintf(`{"members": [
{
	"address": "%s",
	"weight": "1",
	"metadata": "%s"
},
{
	"address": "%s",
	"weight": "2",
	"metadata": "%s"
},
{
	"address": "%s",
	"weight": "3",
	"metadata": "%s"
}]}`, aliceAddr.String(), validMetadata, bobAddr.String(), validMetadata, charlieAddr.String(), validMetadata)
	validBlsMembersFile := testutil.WriteToNewTempFile(s.T(), validBlsMembers)
	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.MsgCreateGroupCmd(),
		append(
			[]string{
				aliceAddr.String(),
				validMetadata,
				validBlsMembersFile.Name(),
				fmt.Sprintf("--%s", client.FlagBlsOnly),
			},
			commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())

	s.groupBls = &group.GroupInfo{GroupId: 2, Admin: aliceAddr.String(), Metadata: []byte{1}, TotalWeight: "6", Version: 1}

	// create 5 group accounts
	for i := 0; i < 5; i++ {
		threshold := i + 1
		if threshold > 4 {
			threshold = 4
		}
		out, err = cli.ExecTestCLICmd(val.ClientCtx, client.MsgCreateGroupAccountCmd(),
			append(
				[]string{
					aliceAddr.String(),
					"2",
					validMetadata,
					fmt.Sprintf("{\"@type\":\"/fetchai.group.v1alpha1.ThresholdDecisionPolicy\", \"threshold\":\"%d\", \"timeout\":\"30000s\"}", threshold),
				},
				commonFlags...,
			),
		)
		s.Require().NoError(err, out.String())
		s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &txResp), out.String())
		s.Require().Equal(uint32(0), txResp.Code, out.String())

		out, err = cli.ExecTestCLICmd(val.ClientCtx, client.QueryGroupAccountsByGroupCmd(), []string{"2", fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
		s.Require().NoError(err, out.String())
	}

	var resBls group.QueryGroupAccountsByGroupResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &resBls))
	s.Require().Equal(len(resBls.GroupAccounts), 5)
	s.groupAccountsBls = resBls.GroupAccounts

	// give group account 0 some balance
	blsAddr0, err := sdk.AccAddressFromBech32(s.groupAccountsBls[0].Address)
	s.Require().NoError(err)
	_, err = banktestutil.MsgSendExec(
		val.ClientCtx,
		val.Address,
		blsAddr0,
		sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(2000))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	)
	s.Require().NoError(err)

	// create a proposal for bls group
	validTxFileNameBls := getTxSendFileName(s, s.groupAccountsBls[0].Address, aliceAddr.String())
	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.MsgCreateProposalCmd(),
		append(
			[]string{
				s.groupAccountsBls[0].Address,
				aliceAddr.String(),
				validTxFileNameBls,
				"",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, aliceAddr.String()),
			},
			commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())

	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.QueryProposalCmd(), []string{"2", fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &proposalRes))
	s.proposalBls = proposalRes.Proposal

	// create a poll for bls group
	now := time.Now()
	expiryTime := now.Add(time.Second * 3000)
	expiry, err := gogotypes.TimestampProto(expiryTime)
	s.Require().NoError(err)
	expiryStr := expiry.String()

	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.MsgCreatePollCmd(),
		append(
			[]string{
				aliceAddr.String(),
				"2",
				"2022 Election",
				"alice,bob,charlie,linda,tom",
				"2",
				expiryStr,
				"",
			},
			commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())

	// vote-poll
	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.MsgVotePollCmd(),
		append(
			[]string{
				"1",
				aliceAddr.String(),
				"alice,linda",
				"",
			},
			commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())

	var pollRes group.QueryPollResponse
	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.QueryPollCmd(), []string{"1", fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &pollRes))
	s.pollBls = pollRes.Poll

	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.QueryVoteForPollByPollVoterCmd(), []string{"1", aliceAddr.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err, out.String())
	var votePollRes group.QueryVoteForPollByPollVoterResponse
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &votePollRes))
	s.votePoll = votePollRes.Vote
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

func (s *IntegrationTestSuite) TestTxCreateGroup() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	validMembers := fmt.Sprintf(`{"members": [{
	  "address": "%s",
		"weight": "1",
		"metadata": "%s"
	}]}`, val.Address.String(), validMetadata)
	validMembersFile := testutil.WriteToNewTempFile(s.T(), validMembers)

	invalidMembersAddress := `{"members": [{
	"address": "",
	"weight": "1"
}]}`
	invalidMembersAddressFile := testutil.WriteToNewTempFile(s.T(), invalidMembersAddress)

	invalidMembersWeight := fmt.Sprintf(`{"members": [{
	  "address": "%s",
		"weight": "0"
	}]}`, val.Address.String())
	invalidMembersWeightFile := testutil.WriteToNewTempFile(s.T(), invalidMembersWeight)

	invalidMembersMetadata := fmt.Sprintf(`{"members": [{
	  "address": "%s",
		"weight": "1",
		"metadata": "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQ=="
	}]}`, val.Address.String())
	invalidMembersMetadataFile := testutil.WriteToNewTempFile(s.T(), invalidMembersMetadata)

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					val.Address.String(),
					"",
					validMembersFile.Name(),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with amino-json",
			append(
				[]string{
					val.Address.String(),
					"",
					validMembersFile.Name(),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"group metadata too long",
			append(
				[]string{
					val.Address.String(),
					"AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQ==",
					"",
				},
				commonFlags...,
			),
			true,
			"metadata: limit exceeded",
			nil,
			0,
		},
		{
			"invalid members address",
			append(
				[]string{
					val.Address.String(),
					"null",
					invalidMembersAddressFile.Name(),
				},
				commonFlags...,
			),
			true,
			"message validation failed: members: address: empty address string is not allowed",
			nil,
			0,
		},
		{
			"invalid members weight",
			append(
				[]string{
					val.Address.String(),
					"null",
					invalidMembersWeightFile.Name(),
				},
				commonFlags...,
			),
			true,
			"message validation failed: member weight: expected a positive decimal, got 0",
			nil,
			0,
		},
		{
			"members metadata too long",
			append(
				[]string{
					val.Address.String(),
					"null",
					invalidMembersMetadataFile.Name(),
				},
				commonFlags...,
			),
			true,
			"member metadata: limit exceeded",
			nil,
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgCreateGroupCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
			s.Require().NoError(s.network.WaitForNextBlock())
		})
	}
}

func (s *IntegrationTestSuite) TestTxUpdateGroupAdmin() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	validMembers := fmt.Sprintf(`{"members": [{
	  "address": "%s",
		"weight": "1",
		"metadata": "%s"
	}]}`, val.Address.String(), validMetadata)
	validMembersFile := testutil.WriteToNewTempFile(s.T(), validMembers)
	out, err := cli.ExecTestCLICmd(val.ClientCtx, client.MsgCreateGroupCmd(),
		append(
			[]string{
				val.Address.String(),
				validMetadata,
				validMembersFile.Name(),
			},
			commonFlags...,
		),
	)

	s.Require().NoError(err, out.String())

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					val.Address.String(),
					"4",
					s.network.Validators[1].Address.String(),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with amino-json",
			append(
				[]string{
					val.Address.String(),
					"5",
					s.network.Validators[1].Address.String(),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"group id invalid",
			append(
				[]string{
					val.Address.String(),
					"",
					s.network.Validators[1].Address.String(),
				},
				commonFlags...,
			),
			true,
			"strconv.ParseUint: parsing \"\": invalid syntax",
			nil,
			0,
		},
		{
			"group doesn't exist",
			append(
				[]string{
					val.Address.String(),
					"12345",
					s.network.Validators[1].Address.String(),
				},
				commonFlags...,
			),
			true,
			"not found",
			nil,
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgUpdateGroupAdminCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
			s.Require().NoError(s.network.WaitForNextBlock())
		})
	}
}

func (s *IntegrationTestSuite) TestTxUpdateGroupMetadata() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					val.Address.String(),
					"3",
					validMetadata,
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with amino-json",
			append(
				[]string{
					val.Address.String(),
					"3",
					validMetadata,
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"group metadata too long",
			append(
				[]string{
					val.Address.String(),
					strconv.FormatUint(s.group.GroupId, 10),
					"AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQ==",
				},
				commonFlags...,
			),
			true,
			"metadata: limit exceeded",
			nil,
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgUpdateGroupMetadataCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
			s.Require().NoError(s.network.WaitForNextBlock())
		})
	}
}

func (s *IntegrationTestSuite) TestTxUpdateGroupMembers() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	validUpdatedMembersFileName := testutil.WriteToNewTempFile(s.T(), fmt.Sprintf(`{"members": [{
		"address": "%s",
		"weight": "0",
		"metadata": "%s"
	}, {
		"address": "%s",
		"weight": "1",
		"metadata": "%s"
	}]}`, val.Address.String(), validMetadata, s.groupAccounts[0].Address, validMetadata)).Name()

	invalidMembersMetadata := fmt.Sprintf(`{"members": [{
	  "address": "%s",
		"weight": "1",
		"metadata": "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQ=="
	}]}`, val.Address.String())
	invalidMembersMetadataFileName := testutil.WriteToNewTempFile(s.T(), invalidMembersMetadata).Name()

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					val.Address.String(),
					"3",
					validUpdatedMembersFileName,
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with amino-json",
			append(
				[]string{
					val.Address.String(),
					"3",
					testutil.WriteToNewTempFile(s.T(), fmt.Sprintf(`{"members": [{
		"address": "%s",
		"weight": "2",
		"metadata": "%s"
	}]}`, s.groupAccounts[0].Address, validMetadata)).Name(),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"group member metadata too long",
			append(
				[]string{
					val.Address.String(),
					strconv.FormatUint(s.group.GroupId, 10),
					invalidMembersMetadataFileName,
				},
				commonFlags...,
			),
			true,
			"member metadata: limit exceeded",
			nil,
			0,
		},
		{
			"group doesn't exist",
			append(
				[]string{
					val.Address.String(),
					"12345",
					validUpdatedMembersFileName,
				},
				commonFlags...,
			),
			true,
			"not found",
			nil,
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgUpdateGroupMembersCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
			s.Require().NoError(s.network.WaitForNextBlock())
		})
	}
}

func (s *IntegrationTestSuite) TestTxCreateGroupAccount() {
	val := s.network.Validators[0]
	wrongAdmin := s.network.Validators[1].Address
	clientCtx := val.ClientCtx

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	groupID := s.group.GroupId

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					val.Address.String(),
					fmt.Sprintf("%v", groupID),
					validMetadata,
					"{\"@type\":\"/fetchai.group.v1alpha1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"timeout\":\"1s\"}",
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with amino-json",
			append(
				[]string{
					val.Address.String(),
					fmt.Sprintf("%v", groupID),
					validMetadata,
					"{\"@type\":\"/fetchai.group.v1alpha1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"timeout\":\"1s\"}",
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"wrong admin",
			append(
				[]string{
					wrongAdmin.String(),
					fmt.Sprintf("%v", groupID),
					validMetadata,
					"{\"@type\":\"/fetchai.group.v1alpha1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"timeout\":\"1s\"}",
				},
				commonFlags...,
			),
			true,
			"key not found",
			&sdk.TxResponse{},
			0,
		},
		{
			"metadata too long",
			append(
				[]string{
					val.Address.String(),
					fmt.Sprintf("%v", groupID),
					strings.Repeat("a", 500),
					"{\"@type\":\"/fetchai.group.v1alpha1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"timeout\":\"1s\"}",
				},
				commonFlags...,
			),
			true,
			"metadata: limit exceeded",
			&sdk.TxResponse{},
			0,
		},
		{
			"wrong group id",
			append(
				[]string{
					val.Address.String(),
					"10",
					validMetadata,
					"{\"@type\":\"/fetchai.group.v1alpha1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"timeout\":\"1s\"}",
				},
				commonFlags...,
			),
			true,
			"not found",
			&sdk.TxResponse{},
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgCreateGroupAccountCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
			s.Require().NoError(s.network.WaitForNextBlock())
		})
	}
}

func (s *IntegrationTestSuite) TestTxUpdateGroupAccountAdmin() {
	val := s.network.Validators[0]
	newAdmin := s.network.Validators[1].Address
	clientCtx := val.ClientCtx
	groupAccount := s.groupAccounts[3]

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					groupAccount.Admin,
					groupAccount.Address,
					newAdmin.String(),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with amino-json",
			append(
				[]string{
					groupAccount.Admin,
					s.groupAccounts[4].Address,
					newAdmin.String(),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"wrong admin",
			append(
				[]string{
					newAdmin.String(),
					groupAccount.Address,
					newAdmin.String(),
				},
				commonFlags...,
			),
			true,
			"key not found",
			&sdk.TxResponse{},
			0,
		},
		{
			"wrong group account",
			append(
				[]string{
					groupAccount.Admin,
					newAdmin.String(),
					newAdmin.String(),
				},
				commonFlags...,
			),
			true,
			"load group account: not found",
			&sdk.TxResponse{},
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgUpdateGroupAccountAdminCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
			s.Require().NoError(s.network.WaitForNextBlock())
		})
	}
}

func (s *IntegrationTestSuite) TestTxUpdateGroupAccountDecisionPolicy() {
	val := s.network.Validators[0]
	newAdmin := s.network.Validators[1].Address
	clientCtx := val.ClientCtx
	groupAccount := s.groupAccounts[2]

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					groupAccount.Admin,
					groupAccount.Address,
					"{\"@type\":\"/fetchai.group.v1alpha1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"timeout\":\"40000s\"}",
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with amino-json",
			append(
				[]string{
					groupAccount.Admin,
					groupAccount.Address,
					"{\"@type\":\"/fetchai.group.v1alpha1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"timeout\":\"50000s\"}",
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"wrong admin",
			append(
				[]string{
					newAdmin.String(),
					groupAccount.Address,
					"{\"@type\":\"/fetchai.group.v1alpha1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"timeout\":\"1s\"}",
				},
				commonFlags...,
			),
			true,
			"key not found",
			&sdk.TxResponse{},
			0,
		},
		{
			"wrong group account",
			append(
				[]string{
					groupAccount.Admin,
					newAdmin.String(),
					"{\"@type\":\"/fetchai.group.v1alpha1.ThresholdDecisionPolicy\", \"threshold\":\"1\", \"timeout\":\"1s\"}",
				},
				commonFlags...,
			),
			true,
			"load group account: not found",
			&sdk.TxResponse{},
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgUpdateGroupAccountDecisionPolicyCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
			s.Require().NoError(s.network.WaitForNextBlock())
		})
	}
}

func (s *IntegrationTestSuite) TestTxUpdateGroupAccountMetadata() {
	val := s.network.Validators[0]
	newAdmin := s.network.Validators[1].Address
	clientCtx := val.ClientCtx
	groupAccount := s.groupAccounts[2]

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					groupAccount.Admin,
					groupAccount.Address,
					validMetadata,
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with amino-json",
			append(
				[]string{
					groupAccount.Admin,
					groupAccount.Address,
					validMetadata,
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"long metadata",
			append(
				[]string{
					groupAccount.Admin,
					groupAccount.Address,
					strings.Repeat("a", 500),
				},
				commonFlags...,
			),
			true,
			"metadata: limit exceeded",
			&sdk.TxResponse{},
			0,
		},
		{
			"wrong admin",
			append(
				[]string{
					newAdmin.String(),
					groupAccount.Address,
					validMetadata,
				},
				commonFlags...,
			),
			true,
			"key not found",
			&sdk.TxResponse{},
			0,
		},
		{
			"wrong group account",
			append(
				[]string{
					groupAccount.Admin,
					newAdmin.String(),
					validMetadata,
				},
				commonFlags...,
			),
			true,
			"load group account: not found",
			&sdk.TxResponse{},
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgUpdateGroupAccountMetadataCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
			s.Require().NoError(s.network.WaitForNextBlock())
		})
	}
}

func (s *IntegrationTestSuite) TestTxCreateProposal() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	validTxFileName := getTxSendFileName(s, s.groupAccounts[0].Address, val.Address.String())
	unauthzTxFileName := getTxSendFileName(s, val.Address.String(), s.groupAccounts[0].Address)
	validTxFileName2 := getTxSendFileName(s, s.groupAccounts[3].Address, val.Address.String())

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					s.groupAccounts[0].Address,
					val.Address.String(),
					validTxFileName,
					"",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with try exec",
			append(
				[]string{
					s.groupAccounts[0].Address,
					val.Address.String(),
					validTxFileName,
					"",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
					fmt.Sprintf("--%s=try", client.FlagExec),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with try exec, not enough yes votes for proposal to pass",
			append(
				[]string{
					s.groupAccounts[3].Address,
					val.Address.String(),
					validTxFileName2,
					"",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
					fmt.Sprintf("--%s=try", client.FlagExec),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with amino-json",
			append(
				[]string{
					s.groupAccounts[0].Address,
					val.Address.String(),
					validTxFileName,
					"",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"metadata too long",
			append(
				[]string{
					s.groupAccounts[0].Address,
					val.Address.String(),
					validTxFileName,
					"AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQ==",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				},
				commonFlags...,
			),
			true,
			"metadata: limit exceeded",
			nil,
			0,
		},
		{
			"unauthorized msg",
			append(
				[]string{
					s.groupAccounts[0].Address,
					val.Address.String(),
					unauthzTxFileName,
					"",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				},
				commonFlags...,
			),
			true,
			"msg does not have group account authorization: unauthorized",
			nil,
			0,
		},
		{
			"invalid proposers",
			append(
				[]string{
					s.groupAccounts[0].Address,
					"invalid",
					validTxFileName,
					"",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				},
				commonFlags...,
			),
			true,
			"proposers: decoding bech32 failed",
			nil,
			0,
		},
		{
			"invalid group account",
			append(
				[]string{
					"invalid",
					val.Address.String(),
					validTxFileName,
					"",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				},
				commonFlags...,
			),
			true,
			"group account: decoding bech32 failed",
			nil,
			0,
		},
		{
			"no group account",
			append(
				[]string{
					val.Address.String(),
					val.Address.String(),
					validTxFileName,
					"",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				},
				commonFlags...,
			),
			true,
			"group account: not found",
			nil,
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgCreateProposalCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
			s.Require().NoError(s.network.WaitForNextBlock())
		})
	}
}

func (s *IntegrationTestSuite) TestTxVote() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	validTxFileName := getTxSendFileName(s, s.groupAccounts[1].Address, val.Address.String())
	for i := 0; i < 2; i++ {
		out, err := cli.ExecTestCLICmd(val.ClientCtx, client.MsgCreateProposalCmd(),
			append(
				[]string{
					s.groupAccounts[1].Address,
					val.Address.String(),
					validTxFileName,
					"",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				},
				commonFlags...,
			),
		)
		s.Require().NoError(err, out.String())
	}

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					"3",
					val.Address.String(),
					"CHOICE_YES",
					"",
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with try exec",
			append(
				[]string{
					"8",
					val.Address.String(),
					"CHOICE_YES",
					"",
					fmt.Sprintf("--%s=try", client.FlagExec),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with try exec, not enough yes votes for proposal to pass",
			append(
				[]string{
					"9",
					val.Address.String(),
					"CHOICE_NO",
					"",
					fmt.Sprintf("--%s=try", client.FlagExec),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with amino-json",
			append(
				[]string{
					"6",
					val.Address.String(),
					"CHOICE_YES",
					"",
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"invalid proposal id",
			append(
				[]string{
					"abcd",
					val.Address.String(),
					"CHOICE_YES",
					"",
				},
				commonFlags...,
			),
			true,
			"invalid syntax",
			nil,
			0,
		},
		{
			"proposal not found",
			append(
				[]string{
					"1234",
					val.Address.String(),
					"CHOICE_YES",
					"",
				},
				commonFlags...,
			),
			true,
			"proposal: not found",
			nil,
			0,
		},
		{
			"metadata too long",
			append(
				[]string{
					"3",
					val.Address.String(),
					"CHOICE_YES",
					"AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQ==",
				},
				commonFlags...,
			),
			true,
			"metadata: limit exceeded",
			nil,
			0,
		},
		{
			"invalid choice",
			append(
				[]string{
					"3",
					val.Address.String(),
					"INVALID_CHOICE",
					"",
				},
				commonFlags...,
			),
			true,
			"not a valid vote choice",
			nil,
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgVoteCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
			s.Require().NoError(s.network.WaitForNextBlock())
		})
	}
}

func (s *IntegrationTestSuite) TestTxVoteAgg() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	aliceInfo, err := clientCtx.Keyring.Key("alice")
	s.Require().NoError(err)
	aliceAddr := sdk.AccAddress(aliceInfo.GetPubKey().Address())

	bobInfo, err := clientCtx.Keyring.Key("bob")
	s.Require().NoError(err)
	bobAddr := sdk.AccAddress(bobInfo.GetPubKey().Address())

	charlieInfo, err := clientCtx.Keyring.Key("charlie")
	s.Require().NoError(err)
	charlieAddr := sdk.AccAddress(charlieInfo.GetPubKey().Address())

	// query group members
	cmd := client.QueryGroupMembersCmd()
	out, err := cli.ExecTestCLICmd(clientCtx, cmd, []string{strconv.FormatUint(s.groupBls.GroupId, 10), fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err, out.String())
	var res group.QueryGroupMembersResponse
	s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &res))
	fileGroupMembers := testutil.WriteToNewTempFile(s.T(), out.String()).Name()

	// create a proposal for bls group
	var txResp = sdk.TxResponse{}
	validTxFileNameBls := getTxSendFileName(s, s.groupAccountsBls[3].Address, aliceAddr.String())
	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.MsgCreateProposalCmd(),
		append(
			[]string{
				s.groupAccountsBls[3].Address,
				aliceAddr.String(),
				validTxFileNameBls,
				"",
				fmt.Sprintf("--%s=%s", flags.FlagFrom, aliceAddr.String()),
			},
			commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())
	s.network.WaitForNextBlock()

	voteFlags := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}

	now := time.Now()
	expiryTime := now.Add(time.Second * 30)
	expiry, err := gogotypes.TimestampProto(expiryTime)
	s.Require().NoError(err)
	expiryStr := expiry.String()

	// basic vote from alice
	cmd = client.GetVoteBasicCmd()
	out, err = cli.ExecTestCLICmd(clientCtx, cmd,
		append(
			[]string{
				aliceAddr.String(),
				"2",
				expiryStr,
				"CHOICE_YES",
			},
			voteFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	aliceVoteYes := testutil.WriteToNewTempFile(s.T(), out.String()).Name()

	cmd = client.GetVerifyVoteBasicCmd()
	out, err = cli.ExecTestCLICmd(clientCtx, cmd, []string{aliceVoteYes})
	s.Require().NoError(err, out.String())

	// basic vote from bob
	cmd = client.GetVoteBasicCmd()
	out, err = cli.ExecTestCLICmd(clientCtx, cmd,
		append(
			[]string{
				bobAddr.String(),
				"2",
				expiryStr,
				"CHOICE_NO",
			},
			voteFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	bobVoteNo := testutil.WriteToNewTempFile(s.T(), out.String()).Name()

	cmd = client.GetVerifyVoteBasicCmd()
	out, err = cli.ExecTestCLICmd(clientCtx, cmd, []string{bobVoteNo})
	s.Require().NoError(err, out.String())

	// basic vote from charlie
	cmd = client.GetVoteBasicCmd()
	out, err = cli.ExecTestCLICmd(clientCtx, cmd,
		append(
			[]string{
				charlieAddr.String(),
				"2",
				expiryStr,
				"CHOICE_YES",
			},
			voteFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	charlieVoteYes := testutil.WriteToNewTempFile(s.T(), out.String()).Name()
	cmd = client.GetVerifyVoteBasicCmd()
	out, err = cli.ExecTestCLICmd(clientCtx, cmd, []string{charlieVoteYes})
	s.Require().NoError(err, out.String())

	// aggregate everyone's vote
	cmd = client.MsgVoteAggCmd()
	voteAggFlags := append(commonFlags, fmt.Sprintf("--%s=try", client.FlagExec))
	out, err = cli.ExecTestCLICmd(clientCtx, cmd,
		append(
			[]string{
				aliceAddr.String(),
				"2",
				expiryStr,
				fileGroupMembers,
				aliceVoteYes,
				bobVoteNo,
				charlieVoteYes,
			},
			voteAggFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())

	var proposalRes group.QueryProposalResponse
	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.QueryProposalCmd(), []string{"2", fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &proposalRes))
	s.Require().Equal(proposalRes.Proposal.Status, group.ProposalStatusClosed)
	s.Require().Equal(proposalRes.Proposal.Result, group.ProposalResultAccepted)
	s.Require().Equal(proposalRes.Proposal.ExecutorResult, group.ProposalExecutorResultSuccess)
}

func (s *IntegrationTestSuite) TestTxExec() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	// create proposals and vote
	for i := 4; i <= 5; i++ {
		validTxFileName := getTxSendFileName(s, s.groupAccounts[0].Address, val.Address.String())
		out, err := cli.ExecTestCLICmd(val.ClientCtx, client.MsgCreateProposalCmd(),
			append(
				[]string{
					s.groupAccounts[0].Address,
					val.Address.String(),
					validTxFileName,
					"",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				},
				commonFlags...,
			),
		)
		s.Require().NoError(err, out.String())

		out, err = cli.ExecTestCLICmd(val.ClientCtx, client.MsgVoteCmd(),
			append(
				[]string{
					fmt.Sprintf("%d", i),
					val.Address.String(),
					"CHOICE_YES",
					"",
				},
				commonFlags...,
			),
		)
		s.Require().NoError(err, out.String())
	}

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					"4",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"with amino-json",
			append(
				[]string{
					"5",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
					fmt.Sprintf("--%s=%s", flags.FlagSignMode, flags.SignModeLegacyAminoJSON),
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"invalid proposal id",
			append(
				[]string{
					"abcd",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				},
				commonFlags...,
			),
			true,
			"invalid syntax",
			nil,
			0,
		},
		{
			"proposal not found",
			append(
				[]string{
					"1234",
					fmt.Sprintf("--%s=%s", flags.FlagFrom, val.Address.String()),
				},
				commonFlags...,
			),
			true,
			"proposal: not found",
			nil,
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgExecCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
			s.Require().NoError(s.network.WaitForNextBlock())
		})
	}
}

func (s *IntegrationTestSuite) TestTxCreatePoll() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	aliceInfo, err := clientCtx.Keyring.Key("alice")
	s.Require().NoError(err)
	aliceAddr := sdk.AccAddress(aliceInfo.GetPubKey().Address())

	now := time.Now()
	deadline := now.Add(time.Second * 3000)
	timeout, err := gogotypes.TimestampProto(deadline)
	s.Require().NoError(err)
	timeoutStr := timeout.String()

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					aliceAddr.String(),
					"2",
					"2023 Election",
					"alice,bob,charlie,eva",
					"2",
					timeoutStr,
					validMetadata,
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"poll expired",
			append(
				[]string{
					aliceAddr.String(),
					"2",
					"2022 Election",
					"alice,bob,charlie",
					"2",
					"2021-08-15T12:00:00Z",
					validMetadata,
				},
				commonFlags...,
			),
			true,
			"deadline of the poll has passed",
			nil,
			0,
		},
		{
			"title too long",
			append(
				[]string{
					aliceAddr.String(),
					"2",
					"AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQ==",
					"alice,bob,charlie",
					"2",
					timeoutStr,
					validMetadata,
				},
				commonFlags...,
			),
			true,
			"poll title: limit exceeded",
			nil,
			0,
		},
		{
			"metadata too long",
			append(
				[]string{
					aliceAddr.String(),
					"2",
					"2022 Election",
					"alice,bob,charlie",
					"2",
					timeoutStr,
					"AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQ==",
				},
				commonFlags...,
			),
			true,
			"metadata: limit exceeded",
			nil,
			0,
		},
		{
			"invalid creator",
			append(
				[]string{
					"invalid",
					"2",
					"2022 Election",
					"alice,bob,charlie",
					"2",
					timeoutStr,
					"",
				},
				commonFlags...,
			),
			true,
			"The specified item could not be found in the keyring",
			nil,
			0,
		},
		{
			"invalid group id",
			append(
				[]string{
					aliceAddr.String(),
					"invalid",
					"2022 Election",
					"alice,bob,charlie,linda,tom",
					"2",
					timeoutStr,
					validMetadata,
				},
				commonFlags...,
			),
			true,
			"strconv.ParseUint: parsing",
			nil,
			0,
		},
		{
			"invalid limit",
			append(
				[]string{
					aliceAddr.String(),
					"2",
					"2022 Election",
					"alice,bob,charlie",
					"5",
					timeoutStr,
					validMetadata,
				},
				commonFlags...,
			),
			true,
			"vote limit exceeds the number of options: invalid value",
			nil,
			0,
		},
		{
			"repeated options",
			append(
				[]string{
					aliceAddr.String(),
					"2",
					"2022 Election",
					"alice,bob,bob",
					"2",
					timeoutStr,
					validMetadata,
				},
				commonFlags...,
			),
			true,
			"duplicate value",
			nil,
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgCreatePollCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
			s.Require().NoError(s.network.WaitForNextBlock())
		})
	}
}

func (s *IntegrationTestSuite) TestTxVotePoll() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	aliceInfo, err := clientCtx.Keyring.Key("alice")
	s.Require().NoError(err)
	aliceAddr := sdk.AccAddress(aliceInfo.GetPubKey().Address())

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
		expectedCode uint32
	}{
		{
			"correct data",
			append(
				[]string{
					"2",
					aliceAddr.String(),
					"alice,bob",
					"",
				},
				commonFlags...,
			),
			false,
			"",
			&sdk.TxResponse{},
			0,
		},
		{
			"invalid poll id",
			append(
				[]string{
					"abcd",
					aliceAddr.String(),
					"alice,bob",
					"",
				},
				commonFlags...,
			),
			true,
			"invalid syntax",
			nil,
			0,
		},
		{
			"poll not found",
			append(
				[]string{
					"1234",
					aliceAddr.String(),
					"alice,bob",
					"",
				},
				commonFlags...,
			),
			true,
			"load poll: not found",
			nil,
			0,
		},
		{
			"metadata too long",
			append(
				[]string{
					"1",
					aliceAddr.String(),
					"alice,bob",
					"AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQ==",
				},
				commonFlags...,
			),
			true,
			"metadata: limit exceeded",
			nil,
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.MsgVotePollCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), tc.respType), out.String())

				txResp := tc.respType.(*sdk.TxResponse)
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
			s.Require().NoError(s.network.WaitForNextBlock())
		})
	}
}

func (s *IntegrationTestSuite) TestTxVotePollAgg() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	aliceInfo, err := clientCtx.Keyring.Key("alice")
	s.Require().NoError(err)
	aliceAddr := sdk.AccAddress(aliceInfo.GetPubKey().Address())

	bobInfo, err := clientCtx.Keyring.Key("bob")
	s.Require().NoError(err)
	bobAddr := sdk.AccAddress(bobInfo.GetPubKey().Address())

	charlieInfo, err := clientCtx.Keyring.Key("charlie")
	s.Require().NoError(err)
	charlieAddr := sdk.AccAddress(charlieInfo.GetPubKey().Address())

	// query group members
	cmd := client.QueryGroupMembersCmd()
	out, err := cli.ExecTestCLICmd(clientCtx, cmd, []string{strconv.FormatUint(s.groupBls.GroupId, 10), fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err, out.String())
	var res group.QueryGroupMembersResponse
	s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &res))
	fileGroupMembers := testutil.WriteToNewTempFile(s.T(), out.String()).Name()

	now := time.Now()
	timeout, err := gogotypes.TimestampProto(now.Add(time.Second * 3000))
	s.Require().NoError(err)

	expiry, err := gogotypes.TimestampProto(now.Add(time.Second * 1000))
	s.Require().NoError(err)
	expiryStr := expiry.String()

	// create a poll for bls group
	var txResp = sdk.TxResponse{}
	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.MsgCreatePollCmd(),
		append(
			[]string{
				aliceAddr.String(),
				"2",
				"2024 Election",
				"alice,bob,clare,david,eva,fred,george",
				"3",
				timeout.String(),
				"",
			},
			commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())

	voteFlags := []string{
		fmt.Sprintf("--%s=json", tmcli.OutputFlag),
	}

	pollID := "3"

	// basic vote from alice
	cmd = client.GetVotePollBasicCmd()
	out, err = cli.ExecTestCLICmd(clientCtx, cmd,
		append(
			[]string{
				aliceAddr.String(),
				pollID,
				expiryStr,
				"alice",
				"david",
				"eva",
			},
			voteFlags...,
		),
	)
	s.Require().NoError(err)
	aliceVote := testutil.WriteToNewTempFile(s.T(), out.String()).Name()

	cmd = client.GetVerifyVotePollBasicCmd()
	out, err = cli.ExecTestCLICmd(clientCtx, cmd, []string{aliceVote})
	s.Require().NoError(err, out.String())

	// basic vote from bob
	cmd = client.GetVotePollBasicCmd()
	out, err = cli.ExecTestCLICmd(clientCtx, cmd,
		append(
			[]string{
				bobAddr.String(),
				pollID,
				expiryStr,
				"bob",
				"david",
			},
			voteFlags...,
		),
	)
	s.Require().NoError(err)
	bobVote := testutil.WriteToNewTempFile(s.T(), out.String()).Name()

	cmd = client.GetVerifyVotePollBasicCmd()
	out, err = cli.ExecTestCLICmd(clientCtx, cmd, []string{bobVote})
	s.Require().NoError(err, out.String())

	// basic vote from charlie
	cmd = client.GetVotePollBasicCmd()
	out, err = cli.ExecTestCLICmd(clientCtx, cmd,
		append(
			[]string{
				charlieAddr.String(),
				pollID,
				expiryStr,
				"david",
				"eva",
			},
			voteFlags...,
		),
	)
	s.Require().NoError(err)
	charlieVote := testutil.WriteToNewTempFile(s.T(), out.String()).Name()
	cmd = client.GetVerifyVotePollBasicCmd()
	out, err = cli.ExecTestCLICmd(clientCtx, cmd, []string{charlieVote})
	s.Require().NoError(err, out.String())

	// aggregate everyone's vote
	cmd = client.MsgVotePollAggCmd()
	out, err = cli.ExecTestCLICmd(clientCtx, cmd,
		append(
			[]string{
				aliceAddr.String(),
				pollID,
				expiryStr,
				fileGroupMembers,
				aliceVote,
				bobVote,
				charlieVote,
			},
			commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())

	var pollRes group.QueryPollResponse
	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.QueryPollCmd(), []string{pollID, fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &pollRes))
	s.Require().Equal(pollRes.Poll.Status, group.PollStatusSubmitted)
	s.Require().Equal(pollRes.Poll.VoteState.Entries, []*group.TallyPollEntry{
		{OptionTitle: "alice", Weight: "1"},
		{OptionTitle: "bob", Weight: "2"},
		{OptionTitle: "david", Weight: "6"},
		{OptionTitle: "eva", Weight: "4"},
	})

	// test invalid
	// duplicate options in basic votes
	cmd = client.GetVotePollBasicCmd()
	_, err = cli.ExecTestCLICmd(clientCtx, cmd,
		append(
			[]string{
				aliceAddr.String(),
				pollID,
				expiryStr,
				"alice",
				"alice",
			},
			voteFlags...,
		),
	)
	s.Require().Contains(err.Error(), "duplicate value")
}

func getTxSendFileName(s *IntegrationTestSuite, from string, to string) string {
	tx := fmt.Sprintf(
		`{"body":{"messages":[{"@type":"/cosmos.bank.v1beta1.MsgSend","from_address":"%s","to_address":"%s","amount":[{"denom":"%s","amount":"10"}]}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[],"fee":{"amount":[],"gas_limit":"200000","payer":"","granter":""}},"signatures":[]}`,
		from, to, s.cfg.BondDenom,
	)
	return testutil.WriteToNewTempFile(s.T(), tx).Name()
}
