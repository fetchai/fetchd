package testsuite

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/fetchai/fetchd/types/testutil/cli"
	"github.com/fetchai/fetchd/x/group"
	"github.com/fetchai/fetchd/x/group/client"
	tmcli "github.com/tendermint/tendermint/libs/cli"
)

func (s *IntegrationTestSuite) TestQueryGroupInfo() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
	}{
		{
			"group not found",
			[]string{"12345", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"not found: invalid request",
			0,
		},
		{
			"group id invalid",
			[]string{"", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"strconv.ParseUint: parsing \"\": invalid syntax",
			0,
		},
		{
			"group found",
			[]string{strconv.FormatUint(s.group.GroupId, 10), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryGroupInfoCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var g group.GroupInfo
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &g))
				s.Require().Equal(s.group.GroupId, g.GroupId)
				s.Require().Equal(s.group.Admin, g.Admin)
				s.Require().Equal(s.group.TotalWeight, g.TotalWeight)
				s.Require().Equal(s.group.Metadata, g.Metadata)
				s.Require().Equal(s.group.Version, g.Version)
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryGroupMembers() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name          string
		args          []string
		expectErr     bool
		expectErrMsg  string
		expectedCode  uint32
		expectMembers []*group.GroupMember
	}{
		{
			"no group",
			[]string{"12345", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.GroupMember{},
		},
		{
			"members found",
			[]string{strconv.FormatUint(s.group.GroupId, 10), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.GroupMember{
				{
					GroupId: s.group.GroupId,
					Member: &group.Member{
						Address:  val.Address.String(),
						Weight:   "3",
						Metadata: []byte{1},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryGroupMembersCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var res group.QueryGroupMembersResponse
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &res))
				s.Require().Equal(len(tc.expectMembers), len(res.Members))
				for i := range res.Members {
					s.Require().Equal(tc.expectMembers[i].GroupId, res.Members[i].GroupId)
					s.Require().Equal(tc.expectMembers[i].Member.Address, res.Members[i].Member.Address)
					s.Require().Equal(tc.expectMembers[i].Member.Metadata, res.Members[i].Member.Metadata)
					s.Require().Equal(tc.expectMembers[i].Member.Weight, res.Members[i].Member.Weight)
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryGroupsByAdmin() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
		expectGroups []*group.GroupInfo
	}{
		{
			"invalid admin address",
			[]string{"invalid"},
			true,
			"decoding bech32 failed: invalid bech32 string",
			0,
			[]*group.GroupInfo{},
		},
		{
			"no group",
			[]string{s.network.Validators[1].Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.GroupInfo{},
		},
		{
			"found groups",
			[]string{val.Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.GroupInfo{
				s.group,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryGroupsByAdminCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var res group.QueryGroupsByAdminResponse
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &res))
				s.Require().Equal(len(tc.expectGroups), len(res.Groups))
				for i := range res.Groups {
					s.Require().Equal(tc.expectGroups[i].GroupId, res.Groups[i].GroupId)
					s.Require().Equal(tc.expectGroups[i].Metadata, res.Groups[i].Metadata)
					s.Require().Equal(tc.expectGroups[i].Version, res.Groups[i].Version)
					s.Require().Equal(tc.expectGroups[i].TotalWeight, res.Groups[i].TotalWeight)
					s.Require().Equal(tc.expectGroups[i].Admin, res.Groups[i].Admin)
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryGroupAccountInfo() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
	}{
		{
			"invalid account address",
			[]string{"invalid", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"decoding bech32 failed: invalid bech32",
			0,
		},
		{
			"group account not found",
			[]string{val.Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"not found: invalid request",
			0,
		},
		{
			"group account found",
			[]string{s.groupAccounts[0].Address, fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryGroupAccountInfoCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var g group.GroupAccountInfo
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &g))
				s.Require().Equal(s.groupAccounts[0].GroupId, g.GroupId)
				s.Require().Equal(s.groupAccounts[0].Address, g.Address)
				s.Require().Equal(s.groupAccounts[0].Admin, g.Admin)
				s.Require().Equal(s.groupAccounts[0].Metadata, g.Metadata)
				s.Require().Equal(s.groupAccounts[0].Version, g.Version)
				s.Require().Equal(s.groupAccounts[0].GetDecisionPolicy(), g.GetDecisionPolicy())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryGroupAccountsByGroup() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name                string
		args                []string
		expectErr           bool
		expectErrMsg        string
		expectedCode        uint32
		expectGroupAccounts []*group.GroupAccountInfo
	}{
		{
			"invalid group id",
			[]string{""},
			true,
			"strconv.ParseUint: parsing \"\": invalid syntax",
			0,
			[]*group.GroupAccountInfo{},
		},
		{
			"no group account",
			[]string{"12345", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.GroupAccountInfo{},
		},
		{
			"found group accounts",
			[]string{strconv.FormatUint(s.group.GroupId, 10), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.GroupAccountInfo{
				s.groupAccounts[0],
				s.groupAccounts[1],
				s.groupAccounts[2],
				s.groupAccounts[3],
				s.groupAccounts[4],
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryGroupAccountsByGroupCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var res group.QueryGroupAccountsByGroupResponse
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &res))
				s.Require().Equal(len(tc.expectGroupAccounts), len(res.GroupAccounts))
				for i := range res.GroupAccounts {
					s.Require().Equal(tc.expectGroupAccounts[i].GroupId, res.GroupAccounts[i].GroupId)
					s.Require().Equal(tc.expectGroupAccounts[i].Metadata, res.GroupAccounts[i].Metadata)
					s.Require().Equal(tc.expectGroupAccounts[i].Version, res.GroupAccounts[i].Version)
					s.Require().Equal(tc.expectGroupAccounts[i].Admin, res.GroupAccounts[i].Admin)
					s.Require().Equal(tc.expectGroupAccounts[i].GetDecisionPolicy(), res.GroupAccounts[i].GetDecisionPolicy())
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryGroupAccountsByAdmin() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name                string
		args                []string
		expectErr           bool
		expectErrMsg        string
		expectedCode        uint32
		expectGroupAccounts []*group.GroupAccountInfo
	}{
		{
			"invalid admin address",
			[]string{"invalid"},
			true,
			"decoding bech32 failed: invalid bech32 string",
			0,
			[]*group.GroupAccountInfo{},
		},
		{
			"no group account",
			[]string{s.network.Validators[1].Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.GroupAccountInfo{},
		},
		{
			"found group accounts",
			[]string{val.Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.GroupAccountInfo{
				s.groupAccounts[0],
				s.groupAccounts[1],
				s.groupAccounts[2],
				s.groupAccounts[3],
				s.groupAccounts[4],
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryGroupAccountsByAdminCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var res group.QueryGroupAccountsByAdminResponse
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &res))
				s.Require().Equal(len(tc.expectGroupAccounts), len(res.GroupAccounts))
				for i := range res.GroupAccounts {
					s.Require().Equal(tc.expectGroupAccounts[i].GroupId, res.GroupAccounts[i].GroupId)
					s.Require().Equal(tc.expectGroupAccounts[i].Metadata, res.GroupAccounts[i].Metadata)
					s.Require().Equal(tc.expectGroupAccounts[i].Version, res.GroupAccounts[i].Version)
					s.Require().Equal(tc.expectGroupAccounts[i].Admin, res.GroupAccounts[i].Admin)
					s.Require().Equal(tc.expectGroupAccounts[i].GetDecisionPolicy(), res.GroupAccounts[i].GetDecisionPolicy())
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryProposal() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
	}{
		{
			"not found",
			[]string{"12345", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"not found",
			0,
		},
		{
			"invalid proposal id",
			[]string{"", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"strconv.ParseUint: parsing \"\": invalid syntax",
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryProposalCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryProposalsByGroupAccount() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name            string
		args            []string
		expectErr       bool
		expectErrMsg    string
		expectedCode    uint32
		expectProposals []*group.Proposal
	}{
		{
			"invalid group account address",
			[]string{"invalid"},
			true,
			"decoding bech32 failed: invalid bech32 string",
			0,
			[]*group.Proposal{},
		},
		{
			"no group account",
			[]string{s.network.Validators[1].Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.Proposal{},
		},
		{
			"found proposals",
			[]string{s.groupAccounts[0].Address, fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.Proposal{
				s.proposal,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryProposalsByGroupAccountCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var res group.QueryProposalsByGroupAccountResponse
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &res))
				s.Require().Equal(len(tc.expectProposals), len(res.Proposals))
				for i := range res.Proposals {
					s.Require().Equal(tc.expectProposals[i], res.Proposals[i])
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryVoteByProposalVoter() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
	}{
		{
			"invalid voter address",
			[]string{"1", "invalid", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"decoding bech32 failed: invalid bech32",
			0,
		},
		{
			"invalid proposal id",
			[]string{"", val.Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"strconv.ParseUint: parsing \"\": invalid syntax",
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryVoteByProposalVoterCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryVotesByProposal() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
		expectVotes  []*group.Vote
	}{
		{
			"invalid proposal id",
			[]string{"", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"strconv.ParseUint: parsing \"\": invalid syntax",
			0,
			[]*group.Vote{},
		},
		{
			"no votes",
			[]string{"12345", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.Vote{},
		},
		{
			"found votes",
			[]string{"1", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.Vote{
				s.vote,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryVotesByProposalCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var res group.QueryVotesByProposalResponse
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &res))
				s.Require().Equal(len(tc.expectVotes), len(res.Votes))
				for i := range res.Votes {
					s.Require().Equal(tc.expectVotes[i], res.Votes[i])
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryVotesByVoter() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
		expectVotes  []*group.Vote
	}{
		{
			"invalid voter address",
			[]string{"abcd", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"decoding bech32 failed: invalid bech32",
			0,
			[]*group.Vote{},
		},
		{
			"no votes",
			[]string{s.groupAccounts[0].Address, fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"",
			0,
			[]*group.Vote{},
		},
		{
			"found votes",
			[]string{val.Address.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.Vote{
				s.vote,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryVotesByVoterCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var res group.QueryVotesByVoterResponse
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &res))
				s.Require().Equal(len(tc.expectVotes), len(res.Votes))
				for i := range res.Votes {
					s.Require().Equal(tc.expectVotes[i], res.Votes[i])
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryPoll() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
	}{
		{
			"not found",
			[]string{"12345", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"not found",
			0,
		},
		{
			"invalid poll id",
			[]string{"", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"strconv.ParseUint: parsing \"\": invalid syntax",
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryPollCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryPollsByGroup() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
		expectPolls  []*group.Poll
	}{
		{
			"invalid group id",
			[]string{"abcd", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"invalid syntax",
			0,
			[]*group.Poll{},
		},
		{
			"not found",
			[]string{"12345", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"",
			0,
			[]*group.Poll{},
		},
		{
			"found polls",
			[]string{fmt.Sprintf("%d", s.pollBls.GroupId), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.Poll{
				s.pollBls,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryPollsByGroupCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var res group.QueryPollsByGroupResponse
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &res))
				s.Require().Equal(len(tc.expectPolls), len(res.Polls))
				for i := range res.Polls {
					s.Require().Equal(tc.expectPolls[i], res.Polls[i])
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryVoteForPollByPollVoter() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	aliceInfo, err := clientCtx.Keyring.Key("alice")
	s.Require().NoError(err)
	aliceAddr := sdk.AccAddress(aliceInfo.GetPubKey().Address())

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
	}{
		{
			"all good",
			[]string{"1", aliceAddr.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
		},
		{
			"invalid voter address",
			[]string{"1", "invalid", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"decoding bech32 failed: invalid bech32",
			0,
		},
		{
			"invalid poll id",
			[]string{"", aliceAddr.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"strconv.ParseUint: parsing \"\": invalid syntax",
			0,
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryVoteForPollByPollVoterCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryVotesForPollByPoll() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
		expectVotes  []*group.VotePoll
	}{
		{
			"invalid proposal id",
			[]string{"", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"strconv.ParseUint: parsing \"\": invalid syntax",
			0,
			[]*group.VotePoll{},
		},
		{
			"no votes",
			[]string{"12345", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.VotePoll{},
		},
		{
			"found votes",
			[]string{"1", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.VotePoll{
				s.votePoll,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryVotesForPollByPollCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var res group.QueryVotesForPollByPollResponse
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &res))
				s.Require().Equal(len(tc.expectVotes), len(res.Votes))
				for i := range res.Votes {
					s.Require().Equal(tc.expectVotes[i], res.Votes[i])
				}
			}
		})
	}
}

func (s *IntegrationTestSuite) TestQueryVotesForPollByVoter() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx

	aliceInfo, err := clientCtx.Keyring.Key("alice")
	s.Require().NoError(err)
	aliceAddr := sdk.AccAddress(aliceInfo.GetPubKey().Address())

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		expectedCode uint32
		expectVotes  []*group.VotePoll
	}{
		{
			"invalid voter address",
			[]string{"abcd", fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"decoding bech32 failed: invalid bech32",
			0,
			[]*group.VotePoll{},
		},
		{
			"no votes",
			[]string{s.groupAccounts[0].Address, fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			true,
			"",
			0,
			[]*group.VotePoll{},
		},
		{
			"found votes",
			[]string{aliceAddr.String(), fmt.Sprintf("--%s=json", tmcli.OutputFlag)},
			false,
			"",
			0,
			[]*group.VotePoll{
				s.votePoll,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := client.QueryVotesForPollByVoterCmd()

			out, err := cli.ExecTestCLICmd(clientCtx, cmd, tc.args)
			if tc.expectErr {
				s.Require().Contains(out.String(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err, out.String())

				var res group.QueryVotesForPollByVoterResponse
				s.Require().NoError(clientCtx.JSONMarshaler.UnmarshalJSON(out.Bytes(), &res))
				s.Require().Equal(len(tc.expectVotes), len(res.Votes))
				for i := range res.Votes {
					s.Require().Equal(tc.expectVotes[i], res.Votes[i])
				}
			}
		})
	}
}
