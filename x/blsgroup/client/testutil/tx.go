package testutil

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fetchai/fetchd/crypto/hd"
	"github.com/stretchr/testify/suite"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/client/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	client "github.com/cosmos/cosmos-sdk/x/group/client/cli"

	blsclient "github.com/fetchai/fetchd/x/blsgroup/client/cli"
)

type IntegrationTestSuite struct {
	suite.Suite

	cfg     network.Config
	network *network.Network

	group       *group.GroupInfo
	groupPolicy *group.GroupPolicyInfo
	proposal    *group.Proposal

	accountBls1 sdk.AccAddress
	accountBls2 sdk.AccAddress
	accountBls3 sdk.AccAddress
}

const validMetadata = "metadata"

func NewIntegrationTestSuite(cfg network.Config) *IntegrationTestSuite {
	return &IntegrationTestSuite{cfg: cfg}
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	var err error
	s.network, err = network.New(s.T(), s.T().TempDir(), s.cfg)
	s.Require().NoError(err)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)

	val := s.network.Validators[0]

	// create new accounts
	info1, _, err := val.ClientCtx.Keyring.NewMnemonic("Bls1", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Bls12381)
	s.Require().NoError(err)
	info2, _, err := val.ClientCtx.Keyring.NewMnemonic("Bls2", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Bls12381)
	s.Require().NoError(err)
	info3, _, err := val.ClientCtx.Keyring.NewMnemonic("Bls3", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Bls12381)
	s.Require().NoError(err)

	pk1, err := info1.GetPubKey()
	s.Require().NoError(err)
	pk2, err := info2.GetPubKey()
	s.Require().NoError(err)
	pk3, err := info3.GetPubKey()
	s.Require().NoError(err)

	s.accountBls1 = sdk.AccAddress(pk1.Address())
	s.accountBls2 = sdk.AccAddress(pk2.Address())
	s.accountBls3 = sdk.AccAddress(pk3.Address())

	for _, account := range []sdk.AccAddress{s.accountBls1, s.accountBls2, s.accountBls3} {
		_, err := banktestutil.MsgSendExec(
			val.ClientCtx,
			val.Address,
			account,
			sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(2000))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		)
		s.Require().NoError(err)

		// send back 1 token in order to set the pubkey of the account
		_, err = banktestutil.MsgSendExec(
			val.ClientCtx,
			account,
			val.Address,
			sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(1))), fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
			fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
			fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
		)
		s.Require().NoError(err)
	}
	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	memberWeight := "3"
	// create a group
	validMembers := fmt.Sprintf(`
	{
		"members": [
			{
				"address": "%s",
				"weight": "%s",
				"metadata": "%s"
			},
			{
				"address": "%s",
				"weight": "%s",
				"metadata": "%s"
			},
			{
				"address": "%s",
				"weight": "%s",
				"metadata": "%s"
			}
		]
	}`,
		s.accountBls1.String(), memberWeight, validMetadata,
		s.accountBls2.String(), memberWeight, validMetadata,
		s.accountBls3.String(), memberWeight, validMetadata,
	)
	validMembersFile := testutil.WriteToNewTempFile(s.T(), validMembers)

	out, err := cli.ExecTestCLICmd(val.ClientCtx, client.MsgCreateGroupCmd(),
		append(
			[]string{
				s.accountBls1.String(),
				validMetadata,
				validMembersFile.Name(),
			},
			commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	var txResp = sdk.TxResponse{}
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())

	s.group = &group.GroupInfo{Id: 1, Admin: s.accountBls1.String(), Metadata: validMetadata, TotalWeight: "9", Version: 1}

	_, err = cli.ExecTestCLICmd(val.ClientCtx, blsclient.MsgRegisterBlsGroupCmd(),
		append(
			[]string{"1", "--from", s.group.Admin},
			commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())

	// create group policy
	policyStr := "{\"@type\":\"/cosmos.group.v1.PercentageDecisionPolicy\", \"percentage\":\"0.5\", \"windows\":{\"voting_period\":\"300s\"}}"
	policyFile := testutil.TempFile(s.T())
	defer os.Remove(policyFile.Name())
	_, err = policyFile.WriteString(policyStr)
	s.Require().NoError(err)

	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.MsgCreateGroupPolicyCmd(),
		append(
			[]string{
				s.accountBls1.String(),
				"1",
				validMetadata,
				policyFile.Name(),
			},
			commonFlags...,
		),
	)

	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())

	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.QueryGroupPoliciesByGroupCmd(), []string{"1", fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err, out.String())

	var res group.QueryGroupPoliciesByGroupResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &res))
	s.Require().Equal(len(res.GroupPolicies), 1)
	s.groupPolicy = res.GroupPolicies[0]

	// create a proposal
	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.MsgSubmitProposalCmd(),
		append(
			[]string{
				s.createCLIProposal(
					s.groupPolicy.Address, s.accountBls1.String(),
					s.groupPolicy.Address, val.Address.String(),
					""),
			},
			commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())

	out, err = cli.ExecTestCLICmd(val.ClientCtx, client.QueryProposalCmd(), []string{"1", fmt.Sprintf("--%s=json", tmcli.OutputFlag)})
	s.Require().NoError(err, out.String())

	var proposalRes group.QueryProposalResponse
	s.Require().NoError(val.ClientCtx.Codec.UnmarshalJSON(out.Bytes(), &proposalRes))
	s.proposal = proposalRes.Proposal
}

func (s *IntegrationTestSuite) TestTxVoteAgg() {
	val := s.network.Validators[0]
	clientCtx := val.ClientCtx
	var commonFlags = []string{
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastBlock),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))).String()),
	}

	curHeight, err := s.network.LatestHeight()
	s.Require().NoError(err)
	timeoutHeight := fmt.Sprintf("%d", curHeight+5)

	// ensure non bls keys are rejected
	_, err = cli.ExecTestCLICmd(clientCtx, blsclient.MsgVoteCmd(),
		[]string{"1", val.Address.String(), "VOTE_OPTION_YES", timeoutHeight},
	)
	s.Require().Error(err)

	// generate signed vote for bls1 and bls2 accounts
	vote1Path := filepath.Join(s.T().TempDir(), "bls1.vote")
	vote2Path := filepath.Join(s.T().TempDir(), "bls2.vote")

	out, err := cli.ExecTestCLICmd(clientCtx, blsclient.MsgVoteCmd(),
		[]string{"1", s.accountBls1.String(), "VOTE_OPTION_YES", timeoutHeight},
	)
	s.Require().NoError(err, out.String())
	// s.T().Logf("vote bls1: %s", out.String())
	s.Require().NoError(ioutil.WriteFile(vote1Path, out.Bytes(), 0600))

	out, err = cli.ExecTestCLICmd(clientCtx, blsclient.MsgVoteCmd(),
		[]string{"1", s.accountBls2.String(), "VOTE_OPTION_YES", timeoutHeight},
	)
	s.Require().NoError(err, out.String())
	// s.T().Logf("vote bls2: %s", out.String())
	s.Require().NoError(ioutil.WriteFile(vote2Path, out.Bytes(), 0600))

	out, err = cli.ExecTestCLICmd(clientCtx, client.QueryGroupMembersCmd(),
		[]string{"1", "--output=json"},
	)
	s.Require().NoError(err, out.String())
	// s.T().Logf("group members: %s", out.String())
	groupMembersFilePath := filepath.Join(s.T().TempDir(), "members")
	s.Require().NoError(ioutil.WriteFile(groupMembersFilePath, out.Bytes(), 0600))

	out, err = cli.ExecTestCLICmd(clientCtx, blsclient.MsgVoteAggCmd(),
		append(
			[]string{"1", groupMembersFilePath, vote1Path, vote2Path,
				fmt.Sprintf("--from=%s", s.accountBls1.String())},
			commonFlags...,
		),
	)
	s.Require().NoError(err, out.String())
	var txResp sdk.TxResponse
	s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
	s.Require().Equal(uint32(0), txResp.Code, out.String())
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
	s.network.Cleanup()
}

// createCLIProposal writes a CLI proposal with a MsgSend to a file. Returns
// the path to the JSON file.
func (s *IntegrationTestSuite) createCLIProposal(groupPolicyAddress, proposer, sendFrom, sendTo, metadata string) string {
	_, err := base64.StdEncoding.DecodeString(metadata)
	s.Require().NoError(err)

	msg := banktypes.MsgSend{
		FromAddress: sendFrom,
		ToAddress:   sendTo,
		Amount:      sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdk.NewInt(10))),
	}
	msgJSON, err := s.cfg.Codec.MarshalInterfaceJSON(&msg)
	s.Require().NoError(err)

	p := client.CLIProposal{
		GroupPolicyAddress: groupPolicyAddress,
		Messages:           []json.RawMessage{msgJSON},
		Metadata:           metadata,
		Proposers:          []string{proposer},
	}

	bz, err := json.Marshal(&p)
	s.Require().NoError(err)

	return testutil.WriteToNewTempFile(s.T(), string(bz)).Name()
}
