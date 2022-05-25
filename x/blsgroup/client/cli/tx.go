package cli

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupcli "github.com/cosmos/cosmos-sdk/x/group/client/cli"
	"github.com/spf13/cobra"

	"github.com/fetchai/fetchd/crypto/keys/bls12381"
	"github.com/fetchai/fetchd/x/blsgroup"
)

// TxCmd returns a root CLI command handler for all x/group transaction commands.
func TxCmd(name string) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        name,
		Short:                      "BLS Group transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		MsgVoteCmd(),
		MsgVoteAggCmd(),
	)

	return txCmd
}

func MsgVoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote [proposal-id] [voter] [vote-option]",
		Short: "Vote on a proposal",
		Long: `Vote on a proposal and the vote will be printed, so it can be aggregated with other votes.

Parameters:
			proposal-id: unique ID of the proposal
			voter: voter account addresses.
			vote-option: choice of the voter(s)
				VOTE_OPTION_UNSPECIFIED: no-op
				VOTE_OPTION_NO: no
				VOTE_OPTION_YES: yes
				VOTE_OPTION_ABSTAIN: abstain
				VOTE_OPTION_NO_WITH_VETO: no-with-veto
`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			err = cmd.Flags().Set(flags.FlagFrom, args[1])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			choice, err := group.VoteOptionFromString(args[2])
			if err != nil {
				return err
			}

			rec, err := clientCtx.Keyring.KeyByAddress(clientCtx.GetFromAddress())
			if err != nil {
				return err
			}
			pub, err := rec.GetPubKey()
			if err != nil {
				return err
			}
			if _, ok := pub.(*bls12381.PubKey); !ok {
				return errors.New("a bls12381 key is required")
			}

			msg := &group.MsgVote{
				ProposalId: proposalID,
				Voter:      clientCtx.GetFromAddress().String(),
				Option:     choice,
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

			vote := &blsgroup.MsgVoteResponse{
				ProposalId: proposalID,
				Voter:      clientCtx.GetFromAddress().String(),
				Option:     choice,
				PubKey:     pubKeyAny,
				Sig:        sigBytes,
			}

			// Force json format here to ease parsing later
			return clientCtx.WithOutputFormat("json").PrintProto(vote)
		},
	}

	cmd.Flags().String(flags.FlagFrom, "", "Name or address of private key with which to sign")

	return cmd
}

func MsgVoteAggCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote-agg [proposal_id] [group-members-json-file] [[vote-json-file]...]",
		Short: "Aggregate signatures of basic votes into aggregated signature and submit the combined votes",
		Long: `Aggregate signatures of basic votes into aggregated signature and submit the combined votes.

Parameters:
			proposal-id: unique ID of the proposal
			group-members-json-file: path to json file that contains group members
			vote-json-file: path to json file that contains a basic vote with a verified signature
`,
		Args: cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			groupMembers, err := parseGroupMembers(clientCtx, args[1])
			if err != nil {
				return err
			}
			for _, mem := range groupMembers {
				if err = mem.ValidateBasic(); err != nil {
					return err
				}
			}

			// make sure group members are sorted by their addresses
			if !sort.SliceIsSorted(groupMembers, sortGroupMembersFunc(groupMembers)) {
				sort.SliceStable(groupMembers, sortGroupMembersFunc(groupMembers))
			}

			groupMembersByAddr := make(map[string]int, len(groupMembers))
			for i, mem := range groupMembers {
				addr := mem.Member.Address
				if _, exists := groupMembersByAddr[addr]; exists {
					return fmt.Errorf("duplicate address: %s", addr)
				}
				groupMembersByAddr[addr] = i
			}

			votes := make([]group.VoteOption, len(groupMembers))
			for i := range votes {
				votes[i] = group.VOTE_OPTION_UNSPECIFIED
			}

			var sigs [][]byte
			for i := 2; i < len(args); i++ {
				vote, err := parseBlsVote(clientCtx, args[i])
				if err != nil {
					return err
				}

				if vote.ProposalId != proposalID {
					return fmt.Errorf("invalid vote from %s: expected proposal id %d", vote.Voter, proposalID)
				}

				memIndex, ok := groupMembersByAddr[vote.Voter]
				if !ok {
					return fmt.Errorf("invalid voter")
				}

				votes[memIndex] = vote.Option
				sigs = append(sigs, vote.Sig)
			}

			sigma, err := bls12381.AggregateSignature(sigs)
			if err != nil {
				return err
			}

			execStr, err := cmd.Flags().GetString(groupcli.FlagExec)
			if err != nil {
				return err
			}

			msg := &blsgroup.MsgVoteAgg{
				Sender:     clientCtx.GetFromAddress().String(),
				ProposalId: proposalID,
				Votes:      votes,
				AggSig:     sigma,
				Exec:       execFromString(execStr),
			}

			if err = msg.ValidateBasic(); err != nil {
				return fmt.Errorf("message validation failed: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(groupcli.FlagExec, "", "Set to 'try' to try to execute proposal immediately after voting")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
