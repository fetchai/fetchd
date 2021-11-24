package client

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"sort"

	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/bls12381"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	gogotypes "github.com/gogo/protobuf/types"
	"github.com/spf13/cobra"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/fetchai/fetchd/x/group"
)

const (
	FlagExec    = "exec"
	ExecTry     = "try"
	FlagBlsOnly = "bls"
)

// TxCmd returns a root CLI command handler for all x/group transaction commands.
func TxCmd(name string) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        name,
		Short:                      "Group transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		MsgCreateGroupCmd(),
		MsgUpdateGroupAdminCmd(),
		MsgUpdateGroupMetadataCmd(),
		MsgUpdateGroupMembersCmd(),
		MsgCreateGroupAccountCmd(),
		MsgUpdateGroupAccountAdminCmd(),
		MsgUpdateGroupAccountDecisionPolicyCmd(),
		MsgUpdateGroupAccountMetadataCmd(),
		MsgCreateProposalCmd(),
		MsgCreatePollCmd(),
		MsgVoteCmd(),
		MsgVoteAggCmd(),
		MsgExecCmd(),
		MsgVotePollCmd(),
		MsgVotePollAggCmd(),
		GetVoteBasicCmd(),
		GetVerifyVoteBasicCmd(),
		GetVotePollBasicCmd(),
		GetVerifyVotePollBasicCmd(),
	)

	return txCmd
}

// MsgCreateGroupCmd creates a CLI command for Msg/CreateGroup.
func MsgCreateGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "create-group [admin] [metadata] [members-json-file]",
		Short: "Create a group which is an aggregation " +
			"of member accounts with associated weights and " +
			"an administrator account. Note, the '--from' flag is " +
			"ignored as it is implied from [admin].",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a group which is an aggregation of member accounts with associated weights and
an administrator account. Note, the '--from' flag is ignored as it is implied from [admin].
Members accounts can be given through a members JSON file that contains an array of members.

Example:
$ %s tx group create-group [admin] [metadata] [members-json-file]

Where members.json contains:

{
	"members": [
		{
			"address": "addr1",
			"weight": "1",
			"metadata": "some metadata"
		},
		{
			"address": "addr2",
			"weight": "1",
			"metadata": "some metadata"
		}
	]
}
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			members, err := parseMembers(clientCtx, args[2])
			if err != nil {
				return err
			}

			b, err := base64.StdEncoding.DecodeString(args[1])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "metadata is malformed, proper base64 string is required")
			}

			bls, err := cmd.Flags().GetBool(FlagBlsOnly)
			if err != nil {
				return err
			}

			msg := &group.MsgCreateGroup{
				Admin:    clientCtx.GetFromAddress().String(),
				Members:  members,
				Metadata: b,
				BlsOnly:  bls,
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().Bool(FlagBlsOnly, false, "Only accept members with bls public keys")

	return cmd
}

// MsgUpdateGroupMembersCmd creates a CLI command for Msg/UpdateGroupMembers.
func MsgUpdateGroupMembersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group-members [admin] [group-id] [members-json-file]",
		Short: "Update a group's members. Set a member's weight to \"0\" to delete it.",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Update a group's members

Example:
$ %s tx group update-group-members [admin] [group-id] [members-json-file]

Where members.json contains:

{
	"members": [
		{
			"address": "addr1",
			"weight": "1",
			"metadata": "some new metadata"
		},
		{
			"address": "addr2",
			"weight": "0",
			"metadata": "some metadata"
		}
	]
}

Set a member's weight to "0" to delete it.
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			members, err := parseMembers(clientCtx, args[2])
			if err != nil {
				return err
			}

			groupID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			msg := &group.MsgUpdateGroupMembers{
				Admin:         clientCtx.GetFromAddress().String(),
				MemberUpdates: members,
				GroupId:       groupID,
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgUpdateGroupAdminCmd creates a CLI command for Msg/UpdateGroupAdmin.
func MsgUpdateGroupAdminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group-admin [admin] [group-id] [new-admin]",
		Short: "Update a group's admin",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			groupID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			msg := &group.MsgUpdateGroupAdmin{
				Admin:    clientCtx.GetFromAddress().String(),
				NewAdmin: args[2],
				GroupId:  groupID,
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgUpdateGroupMetadataCmd creates a CLI command for Msg/UpdateGroupMetadata.
func MsgUpdateGroupMetadataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group-metadata [admin] [group-id] [metadata]",
		Short: "Update a group's metadata",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			groupID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			b, err := base64.StdEncoding.DecodeString(args[2])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "metadata is malformed, proper base64 string is required")
			}

			msg := &group.MsgUpdateGroupMetadata{
				Admin:    clientCtx.GetFromAddress().String(),
				Metadata: b,
				GroupId:  groupID,
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgCreateGroupAccountCmd creates a CLI command for Msg/CreateGroupAccount.
func MsgCreateGroupAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "create-group-account [admin] [group-id] [metadata] [decision-policy]",
		Short: "Create a group account which is an account " +
			"associated with a group and a decision policy. " +
			"Note, the '--from' flag is " +
			"ignored as it is implied from [admin].",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create a group account which is an account associated with a group and a decision policy.
Note, the '--from' flag is ignored as it is implied from [admin].

Example:
$ %s tx group create-group-account [admin] [group-id] [metadata] \
'{"@type":"/fetchai.group.v1alpha1.ThresholdDecisionPolicy", "threshold":"1", "timeout":"1s"}'
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			groupID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			var policy group.DecisionPolicy
			if err := clientCtx.JSONMarshaler.UnmarshalInterfaceJSON([]byte(args[3]), &policy); err != nil {
				return err
			}

			b, err := base64.StdEncoding.DecodeString(args[2])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "metadata is malformed, proper base64 string is required")
			}

			msg, err := group.NewMsgCreateGroupAccount(
				clientCtx.GetFromAddress(),
				groupID,
				b,
				policy,
			)
			if err != nil {
				return err
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgUpdateGroupAccountAdminCmd creates a CLI command for Msg/UpdateGroupAccountAdmin.
func MsgUpdateGroupAccountAdminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group-account-admin [admin] [group-account] [new-admin]",
		Short: "Update a group account admin",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &group.MsgUpdateGroupAccountAdmin{
				Admin:    clientCtx.GetFromAddress().String(),
				Address:  args[1],
				NewAdmin: args[2],
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgUpdateGroupAccountDecisionPolicyCmd creates a CLI command for Msg/UpdateGroupAccountDecisionPolicy.
func MsgUpdateGroupAccountDecisionPolicyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group-account-policy [admin] [group-account] [decision-policy]",
		Short: "Update a group account decision policy",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var policy group.DecisionPolicy
			if err := clientCtx.JSONMarshaler.UnmarshalInterfaceJSON([]byte(args[2]), &policy); err != nil {
				return err
			}

			accountAddress, err := sdk.AccAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			msg, err := group.NewMsgUpdateGroupAccountDecisionPolicyRequest(
				clientCtx.GetFromAddress(),
				accountAddress,
				policy,
			)
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgUpdateGroupAccountMetadataCmd creates a CLI command for Msg/MsgUpdateGroupAccountMetadata.
func MsgUpdateGroupAccountMetadataCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-group-account-metadata [admin] [group-account] [new-metadata]",
		Short: "Update a group account metadata",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			b, err := base64.StdEncoding.DecodeString(args[2])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "metadata is malformed, proper base64 string is required")
			}

			msg := &group.MsgUpdateGroupAccountMetadata{
				Admin:    clientCtx.GetFromAddress().String(),
				Address:  args[1],
				Metadata: b,
			}
			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgCreateProposalCmd creates a CLI command for Msg/CreateProposal.
func MsgCreateProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-proposal [group-account] [proposer[,proposer]*] [msg_tx_json_file] [metadata]",
		Short: "Submit a new proposal",
		Long: `Submit a new proposal.

Parameters:
			group-account: address of the group account
			proposer: comma separated (no spaces) list of proposer account addresses. Example: "addr1,addr2" 
			Metadata: metadata for the proposal
			msg_tx_json_file: path to json file with messages that will be executed if the proposal is accepted.
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			proposers := strings.Split(args[1], ",")
			for i := range proposers {
				proposers[i] = strings.TrimSpace(proposers[i])
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			theTx, err := authclient.ReadTxFromFile(clientCtx, args[2])
			if err != nil {
				return err
			}
			msgs := theTx.GetMsgs()

			b, err := base64.StdEncoding.DecodeString(args[3])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "metadata is malformed, proper base64 string is required")
			}

			execStr, err := cmd.Flags().GetString(FlagExec)
			if err != nil {
				return err
			}

			msg, err := group.NewMsgCreateProposalRequest(
				args[0],
				proposers,
				msgs,
				b,
				execFromString(execStr),
			)
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagExec, "", "Set to 1 to try to execute proposal immediately after creation (proposers signatures are considered as Yes votes)")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgCreatePollCmd creates a CLI command for Msg/CreatePoll.
func MsgCreatePollCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-poll [creator] [group-id] [title] [option[,option]*] [vote-limit] [timeout] [metadata]",
		Short: "Submit a new poll",
		Long: `Submit a new poll.

Parameters:
			group-id: unique id of the group
			title: title of the poll
			option: comma separated (no spaces) list of options. Example: "option1,option2"
			vote-limit: number of options each voter can choose
			time-out: deadline for voting the poll
			Metadata: metadata for the poll
`,
		Args: cobra.ExactArgs(7),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			groupID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			title := args[2]

			optionTitles := strings.Split(args[3], ",")
			for i := range optionTitles {
				optionTitles[i] = strings.TrimSpace(optionTitles[i])
			}
			options := group.Options{Titles: optionTitles}

			voteLimit, err := strconv.ParseInt(args[4], 10, 32)
			if err != nil {
				return err
			}

			timeString := fmt.Sprintf("\"%s\"", args[5])
			var timeout gogotypes.Timestamp
			err = clientCtx.JSONMarshaler.UnmarshalJSON([]byte(timeString), &timeout)
			if err != nil {
				return err
			}
			timeNow := gogotypes.TimestampNow()
			if timeout.Compare(timeNow) <= 0 {
				return fmt.Errorf("deadline of the poll has passed")
			}

			b, err := base64.StdEncoding.DecodeString(args[6])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "metadata is malformed, proper base64 string is required")
			}

			msg := &group.MsgCreatePoll{
				GroupId:   groupID,
				Title:     title,
				Options:   options,
				Creator:   args[0],
				VoteLimit: int32(voteLimit),
				Metadata:  b,
				Timeout:   timeout,
			}
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagExec, "", "Set to 1 to try to execute proposal immediately after creation (proposers signatures are considered as Yes votes)")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgVoteCmd creates a CLI command for Msg/Vote.
func MsgVoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote [proposal-id] [voter] [choice] [metadata]",
		Short: "Vote on a proposal",
		Long: `Vote on a proposal.

Parameters:
			proposal-id: unique ID of the proposal
			voter: voter account addresses.
			choice: choice of the voter(s)
				CHOICE_UNSPECIFIED: no-op
				CHOICE_NO: no
				CHOICE_YES: yes
				CHOICE_ABSTAIN: abstain
				CHOICE_VETO: veto
			Metadata: metadata for the vote
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[1])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			choice, err := group.ChoiceFromString(args[2])
			if err != nil {
				return err
			}

			b, err := base64.StdEncoding.DecodeString(args[3])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "metadata is malformed, proper base64 string is required")
			}

			execStr, err := cmd.Flags().GetString(FlagExec)
			if err != nil {
				return err
			}

			msg := &group.MsgVote{
				ProposalId: proposalID,
				Voter:      args[1],
				Choice:     choice,
				Metadata:   b,
				Exec:       execFromString(execStr),
			}
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagExec, "", "Set to 1 to try to execute proposal immediately after voting")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgExecCmd creates a CLI command for Msg/MsgExec.
func MsgExecCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec [proposal-id]",
		Short: "Execute a proposal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := &group.MsgExec{
				ProposalId: proposalID,
				Signer:     clientCtx.GetFromAddress().String(),
			}
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func MsgVoteAggCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote-agg [sender] [proposal_id] [timeout] [group-members-json-file] [[vote-json-file]...]",
		Short: "Aggregate signatures of basic votes into aggregated signature and submit the combined votes",
		Long: `Aggregate signatures of basic votes into aggregated signature and submit the combined votes.

Parameters:
			sender: sender's account address
			proposal-id: unique ID of the proposal
			timeout: UTC time for the submission deadline of the aggregated vote, e.g., 2021-08-15T12:00:00Z
			group-members-json-file: path to json file that contains group members
			vote-json-file: path to json file that contains a basic vote with a verified signature
`,
		Args: cobra.MinimumNArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			timeString := fmt.Sprintf("\"%s\"", args[2])
			var timeout gogotypes.Timestamp
			err = clientCtx.JSONMarshaler.UnmarshalJSON([]byte(timeString), &timeout)
			if err != nil {
				return err
			}
			timeNow := gogotypes.TimestampNow()
			if timeout.Compare(timeNow) <= 0 {
				return fmt.Errorf("deadline for submitting the vote has passed")
			}

			groupMembers, err := parseGroupMembers(clientCtx, args[3])
			if err != nil {
				return err
			}
			for _, mem := range groupMembers {
				if err = mem.ValidateBasic(); err != nil {
					return err
				}
			}

			// make sure group members are sorted by their addresses
			sorted := sort.SliceIsSorted(groupMembers, func(i, j int) bool {
				addri, err := sdk.AccAddressFromBech32(groupMembers[i].Member.Address)
				if err != nil {
					panic(err)
				}
				addrj, err := sdk.AccAddressFromBech32(groupMembers[j].Member.Address)
				if err != nil {
					panic(err)
				}
				return bytes.Compare(addri, addrj) < 0
			})
			if !sorted {
				sort.Slice(groupMembers, func(i, j int) bool {
					addri, err := sdk.AccAddressFromBech32(groupMembers[i].Member.Address)
					if err != nil {
						panic(err)
					}
					addrj, err := sdk.AccAddressFromBech32(groupMembers[j].Member.Address)
					if err != nil {
						panic(err)
					}
					return bytes.Compare(addri, addrj) < 0
				})
			}

			index := make(map[string]int, len(groupMembers))
			for i, mem := range groupMembers {
				addr := mem.Member.Address
				if _, exists := index[addr]; exists {
					return fmt.Errorf("duplicate address: %s", addr)
				}
				index[addr] = i
			}

			votes := make([]group.Choice, len(groupMembers))
			for i := range votes {
				votes[i] = group.Choice_CHOICE_UNSPECIFIED
			}

			var sigs [][]byte
			for i := 4; i < len(args); i++ {
				vote, err := parseVoteBasic(clientCtx, args[i])
				if err != nil {
					return err
				}

				if vote.ProposalId != proposalID || !vote.Expiry.Equal(timeout) {
					return fmt.Errorf("invalid vote from %s: expected proposal id %d and timeout %s", vote.Voter, proposalID, timeout.String())
				}

				memIndex, ok := index[vote.Voter]
				if !ok {
					return fmt.Errorf("invalid voter")
				}

				votes[memIndex] = vote.Choice
				sigs = append(sigs, vote.Sig)
			}

			sigma, err := bls12381.AggregateSignature(sigs)
			if err != nil {
				return err
			}

			execStr, err := cmd.Flags().GetString(FlagExec)
			if err != nil {
				return err
			}

			msg := &group.MsgVoteAgg{
				Sender:     args[0],
				ProposalId: proposalID,
				Votes:      votes,
				Expiry:     timeout,
				AggSig:     sigma,
				Metadata:   nil,
				Exec:       execFromString(execStr),
			}

			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagExec, "", "Set to 1 to try to execute proposal immediately after voting")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// MsgVotePollCmd creates a CLI command for Msg/Vote.
func MsgVotePollCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote-poll [poll-id] [voter] [option[,option]*] [metadata]",
		Short: "Vote on a poll",
		Long: `Vote on a poll.

Parameters:
			poll-id: unique ID of the poll
			voter: voter account addresses.
			option: options chosen by the voter
			Metadata: metadata for the vote
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[1])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			pollID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			optionTitles := strings.Split(args[2], ",")
			for i := range optionTitles {
				optionTitles[i] = strings.TrimSpace(optionTitles[i])
			}
			options := group.Options{Titles: optionTitles}

			b, err := base64.StdEncoding.DecodeString(args[3])
			if err != nil {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "metadata is malformed, proper base64 string is required")
			}

			msg := &group.MsgVotePoll{
				PollId:   pollID,
				Voter:    args[1],
				Options:  options,
				Metadata: b,
			}
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func MsgVotePollAggCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote-poll-agg [sender] [poll_id] [expiry] [group-members-json-file] [[vote-json-file]...]",
		Short: "Aggregate signatures of basic votes into aggregated signature and submit the combined votes",
		Long: `Aggregate signatures of basic votes into aggregated signature and submit the combined votes.

Parameters:
			sender: sender's account address
			poll-id: unique ID of the poll
			timeout: UTC time for the submission deadline of the aggregated vote, e.g., 2021-08-15T12:00:00Z
			group-members-json-file: path to json file that contains group members
			vote-json-file: path to json file that contains a basic vote with a verified signature
`,
		Args: cobra.MinimumNArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			pollID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			timeString := fmt.Sprintf("\"%s\"", args[2])
			var timeout gogotypes.Timestamp
			err = clientCtx.JSONMarshaler.UnmarshalJSON([]byte(timeString), &timeout)
			if err != nil {
				return err
			}
			timeNow := gogotypes.TimestampNow()
			if timeout.Compare(timeNow) <= 0 {
				return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "deadline for submitting the vote has passed")
			}

			groupMembers, err := parseGroupMembers(clientCtx, args[3])
			if err != nil {
				return err
			}
			for _, mem := range groupMembers {
				if err = mem.ValidateBasic(); err != nil {
					return err
				}
			}

			// make sure group members are sorted by their addresses
			sorted := sort.SliceIsSorted(groupMembers, func(i, j int) bool {
				addri, err := sdk.AccAddressFromBech32(groupMembers[i].Member.Address)
				if err != nil {
					panic(err)
				}
				addrj, err := sdk.AccAddressFromBech32(groupMembers[j].Member.Address)
				if err != nil {
					panic(err)
				}
				return bytes.Compare(addri, addrj) < 0
			})
			if !sorted {
				sort.Slice(groupMembers, func(i, j int) bool {
					addri, err := sdk.AccAddressFromBech32(groupMembers[i].Member.Address)
					if err != nil {
						panic(err)
					}
					addrj, err := sdk.AccAddressFromBech32(groupMembers[j].Member.Address)
					if err != nil {
						panic(err)
					}
					return bytes.Compare(addri, addrj) < 0
				})
			}

			index := make(map[string]int, len(groupMembers))
			for i, mem := range groupMembers {
				addr := mem.Member.Address
				if _, exists := index[addr]; exists {
					return fmt.Errorf("duplicate address: %s", addr)
				}
				index[addr] = i
			}

			votes := make([]group.Options, len(groupMembers))

			var sigs [][]byte
			for i := 4; i < len(args); i++ {
				vote, err := parseVotePollBasic(clientCtx, args[i])
				if err != nil {
					return err
				}

				if vote.PollId != pollID || !vote.Expiry.Equal(timeout) {
					return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest,
						"invalid vote from %s: expect poll id %d and timeout %s", vote.Voter, pollID, timeout.String())
				}

				memIndex, ok := index[vote.Voter]
				if !ok {
					return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "voter %s", vote.Voter)
				}

				votes[memIndex] = vote.Options
				sigs = append(sigs, vote.Sig)
			}

			sigma, err := bls12381.AggregateSignature(sigs)
			if err != nil {
				return err
			}

			msg := &group.MsgVotePollAgg{
				Sender:   args[0],
				PollId:   pollID,
				Votes:    votes,
				Expiry:   timeout,
				AggSig:   sigma,
				Metadata: []byte(fmt.Sprintf("submitted as aggregated vote by account %s", args[0])),
			}

			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// GetVoteBasicCmd creates a CLI command for Msg/VoteBasic.
func GetVoteBasicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote-basic [voter] [proposal-id] [expiry] [choice]",
		Short: "Vote on a proposal",
		Long: `Vote on a proposal and the vote will be aggregated with other votes.

Parameters:
			voter: voter account addresses.
			proposal-id: unique ID of the proposal
			choice: choice of the voter(s)
				CHOICE_UNSPECIFIED: no-op
				CHOICE_NO: no
				CHOICE_YES: yes
				CHOICE_ABSTAIN: abstain
				CHOICE_VETO: veto
			timeout: UTC time for the submission deadline of the vote, e.g., 2021-08-15T12:00:00Z
`,
		Args: cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			timeString := fmt.Sprintf("\"%s\"", args[2])
			var expiry gogotypes.Timestamp
			err = clientCtx.JSONMarshaler.UnmarshalJSON([]byte(timeString), &expiry)
			if err != nil {
				return err
			}

			timeNow := gogotypes.TimestampNow()
			if expiry.Compare(timeNow) <= 0 {
				return fmt.Errorf("deadline for submitting the vote has passed")
			}

			choice, err := group.ChoiceFromString(args[3])
			if err != nil {
				return err
			}

			msg := &group.MsgVoteBasic{
				ProposalId: proposalID,
				Choice:     choice,
				Expiry:     expiry,
			}

			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			bytesToSign := msg.GetSignBytes()
			sigBytes, pubKey, err := clientCtx.Keyring.Sign(clientCtx.GetFromName(), bytesToSign)
			if err != nil {
				return fmt.Errorf("signature failed: %w", err)
			}

			pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
			if err != nil {
				return err
			}

			vote := &group.MsgVoteBasicResponse{
				ProposalId: proposalID,
				Choice:     choice,
				Expiry:     expiry,
				Voter:      args[0],
				PubKey:     pubKeyAny,
				Sig:        sigBytes,
			}

			return clientCtx.PrintProto(vote)
		},
	}

	cmd.Flags().String(flags.FlagFrom, "", "Name or address of private key with which to sign")
	cmd.Flags().StringP(tmcli.OutputFlag, "o", "text", "Output format (text|json)")

	return cmd
}

// GetVerifyVoteBasicCmd creates a CLI command for aggregating basic votes.
func GetVerifyVoteBasicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify-vote-basic [file]",
		Short: "Verify signature for a basic vote",
		Long: `Verify signature for a basic vote.

Parameters:
			file: a basic vote with signature
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			vote, err := parseVoteBasic(clientCtx, args[0])
			if err != nil {
				return err
			}

			if err = vote.ValidateBasic(); err != nil {
				return err
			}

			if err = vote.VerifySignature(); err != nil {
				return err
			}

			cmd.Println("Verification Successful!")

			return nil
		},
	}

	return cmd
}

// GetVotePollBasicCmd creates a CLI command for Msg/VotePollBasic.
func GetVotePollBasicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote-poll-basic [voter] [poll-id] [expiry] [[option]...] ",
		Short: "Vote on a poll",
		Long: `Vote on a proposal and the vote will be aggregated with other votes.

Parameters:
			voter: voter account addresses.
			proposal-id: unique ID of the proposal
			timeout: UTC time for the submission deadline of the vote, e.g., 2021-08-15T12:00:00Z
			options: options chosen by the voter(s)
`,
		Args: cobra.MinimumNArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := cmd.Flags().Set(flags.FlagFrom, args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			pollID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			timeString := fmt.Sprintf("\"%s\"", args[2])
			var timeout gogotypes.Timestamp
			err = clientCtx.JSONMarshaler.UnmarshalJSON([]byte(timeString), &timeout)
			if err != nil {
				return err
			}

			timeNow := gogotypes.TimestampNow()
			if timeout.Compare(timeNow) <= 0 {
				return fmt.Errorf("deadline for submitting the vote has passed")
			}

			keyInfo, err := clientCtx.Keyring.Key(clientCtx.GetFromName())
			if err != nil {
				return err
			}
			pubKey := keyInfo.GetPubKey()
			pubKeyAny, err := codectypes.NewAnyWithValue(pubKey)
			if err != nil {
				return err
			}

			var optionTitles []string
			var sigs [][]byte
			for i := 3; i < len(args); i++ {
				option := args[i]
				optionTitles = append(optionTitles, option)

				msg := group.MsgVotePollBasic{
					PollId: pollID,
					Option: option,
					Expiry: timeout,
				}
				if err = msg.ValidateBasic(); err != nil {
					return err
				}

				bytesToSign := msg.GetSignBytes()
				sigBytes, _, err := clientCtx.Keyring.Sign(clientCtx.GetFromName(), bytesToSign)
				if err != nil {
					return err
				}

				sigs = append(sigs, sigBytes)
			}

			sigBytes, err := bls12381.AggregateSignature(sigs)
			if err != nil {
				return err
			}

			options := group.Options{Titles: optionTitles}
			if err = options.ValidateBasic(); err != nil {
				return err
			}

			vote := &group.MsgVotePollBasicResponse{
				PollId:  pollID,
				Options: options,
				Expiry:  timeout,
				Voter:   args[0],
				PubKey:  pubKeyAny,
				Sig:     sigBytes,
			}

			return clientCtx.PrintProto(vote)
		},
	}

	cmd.Flags().String(flags.FlagFrom, "", "Name or address of private key with which to sign")
	cmd.Flags().StringP(tmcli.OutputFlag, "o", "text", "Output format (text|json)")

	return cmd
}

// GetVerifyVotePollBasicCmd creates a CLI command for aggregating basic votes.
func GetVerifyVotePollBasicCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify-vote-poll-basic [file]",
		Short: "Verify signature for a basic vote for poll",
		Long: `Verify signature for a basic vote for poll.

Parameters:
			file: a basic vote for poll with signature
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			vote, err := parseVotePollBasic(clientCtx, args[0])
			if err != nil {
				return err
			}

			if err = vote.ValidateBasic(); err != nil {
				return err
			}

			if err = vote.VerifySignature(); err != nil {
				return err
			}

			cmd.Println("Verification Successful!")

			return nil
		},
	}

	return cmd
}
