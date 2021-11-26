package client

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/fetchai/fetchd/x/group"
	"github.com/spf13/cobra"
)

// QueryCmd returns the cli query commands for the group module.
func QueryCmd(name string) *cobra.Command {
	queryCmd := &cobra.Command{
		Use:                        name,
		Short:                      "Querying commands for the group module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	queryCmd.AddCommand(
		QueryGroupInfoCmd(),
		QueryGroupAccountInfoCmd(),
		QueryGroupMembersCmd(),
		QueryGroupsByAdminCmd(),
		QueryGroupAccountsByGroupCmd(),
		QueryGroupAccountsByAdminCmd(),
		QueryProposalCmd(),
		QueryProposalsByGroupAccountCmd(),
		QueryVoteByProposalVoterCmd(),
		QueryVotesByProposalCmd(),
		QueryVotesByVoterCmd(),
		QueryPollCmd(),
		QueryPollsByGroupCmd(),
		QueryPollsByCreatorCmd(),
		QueryVoteForPollByPollVoterCmd(),
		QueryVotesForPollByPollCmd(),
		QueryVotesForPollByVoterCmd(),
	)

	return queryCmd
}

// QueryGroupInfoCmd creates a CLI command for Query/GroupInfo.
func QueryGroupInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group-info [id]",
		Short: "Query for group info by group id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			groupID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := group.NewQueryClient(clientCtx)

			res, err := queryClient.GroupInfo(cmd.Context(), &group.QueryGroupInfoRequest{
				GroupId: groupID,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res.Info)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryGroupAccountInfoCmd creates a CLI command for Query/GroupAccountInfo.
func QueryGroupAccountInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group-account-info [group-account]",
		Short: "Query for group account info by group account address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := group.NewQueryClient(clientCtx)

			res, err := queryClient.GroupAccountInfo(cmd.Context(), &group.QueryGroupAccountInfoRequest{
				Address: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res.Info)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryGroupMembersCmd creates a CLI command for Query/GroupMembers.
func QueryGroupMembersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group-members [id]",
		Short: "Query for group members by group id with pagination flags",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			groupID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := group.NewQueryClient(clientCtx)

			res, err := queryClient.GroupMembers(cmd.Context(), &group.QueryGroupMembersRequest{
				GroupId:    groupID,
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryGroupsByAdminCmd creates a CLI command for Query/GroupsByAdmin.
func QueryGroupsByAdminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "groups-by-admin [admin]",
		Short: "Query for groups by admin account address with pagination flags",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := group.NewQueryClient(clientCtx)

			res, err := queryClient.GroupsByAdmin(cmd.Context(), &group.QueryGroupsByAdminRequest{
				Admin:      args[0],
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryGroupAccountsByGroupCmd creates a CLI command for Query/GroupAccountsByGroup.
func QueryGroupAccountsByGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group-accounts-by-group [group-id]",
		Short: "Query for group accounts by group id with pagination flags",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			groupID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := group.NewQueryClient(clientCtx)

			res, err := queryClient.GroupAccountsByGroup(cmd.Context(), &group.QueryGroupAccountsByGroupRequest{
				GroupId:    groupID,
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryGroupAccountsByAdminCmd creates a CLI command for Query/GroupAccountsByAdmin.
func QueryGroupAccountsByAdminCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "group-accounts-by-admin [admin]",
		Short: "Query for group accounts by admin account address with pagination flags",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := group.NewQueryClient(clientCtx)

			res, err := queryClient.GroupAccountsByAdmin(cmd.Context(), &group.QueryGroupAccountsByAdminRequest{
				Admin:      args[0],
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryProposalCmd creates a CLI command for Query/Proposal.
func QueryProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposal [id]",
		Short: "Query for proposal by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := group.NewQueryClient(clientCtx)

			res, err := queryClient.Proposal(cmd.Context(), &group.QueryProposalRequest{
				ProposalId: proposalID,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryProposalsByGroupAccountCmd creates a CLI command for Query/ProposalsByGroupAccount.
func QueryProposalsByGroupAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposals-by-group-account [group-account]",
		Short: "Query for proposals by group account address with pagination flags",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := group.NewQueryClient(clientCtx)

			res, err := queryClient.ProposalsByGroupAccount(cmd.Context(), &group.QueryProposalsByGroupAccountRequest{
				Address:    args[0],
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryVoteByProposalVoterCmd creates a CLI command for Query/VoteByProposalVoter.
func QueryVoteByProposalVoterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote [proposal-id] [voter]",
		Short: "Query for vote by proposal id and voter account address",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := group.NewQueryClient(clientCtx)

			res, err := queryClient.VoteByProposalVoter(cmd.Context(), &group.QueryVoteByProposalVoterRequest{
				ProposalId: proposalID,
				Voter:      args[1],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryVotesByProposalCmd creates a CLI command for Query/VotesByProposal.
func QueryVotesByProposalCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "votes-by-proposal [proposal-id]",
		Short: "Query for votes by proposal id with pagination flags",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			proposalID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := group.NewQueryClient(clientCtx)

			res, err := queryClient.VotesByProposal(cmd.Context(), &group.QueryVotesByProposalRequest{
				ProposalId: proposalID,
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryVotesByVoterCmd creates a CLI command for Query/VotesByVoter.
func QueryVotesByVoterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "votes-by-voter [voter]",
		Short: "Query for votes by voter account address with pagination flags",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := group.NewQueryClient(clientCtx)

			res, err := queryClient.VotesByVoter(cmd.Context(), &group.QueryVotesByVoterRequest{
				Voter:      args[0],
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryPollCmd creates a CLI command for Query/Poll.
func QueryPollCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "poll [id]",
		Short: "Query for poll by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pollID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := group.NewQueryClient(clientCtx)

			res, err := queryClient.Poll(cmd.Context(), &group.QueryPollRequest{
				PollId: pollID,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryPollsByGroupCmd creates a CLI command for Query/PollsByGroup.
func QueryPollsByGroupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "polls-by-group [group-id]",
		Short: "Query for polls by group id with pagination flags",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			groupID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := group.NewQueryClient(clientCtx)

			res, err := queryClient.PollsByGroup(cmd.Context(), &group.QueryPollsByGroupRequest{
				GroupId:    groupID,
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryPollsByCreatorCmd creates a CLI command for Query/PollsByCreator.
func QueryPollsByCreatorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "polls-by-creator [creator]",
		Short: "Query for polls by creator address with pagination flags",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := group.NewQueryClient(clientCtx)

			res, err := queryClient.PollsByCreator(cmd.Context(), &group.QueryPollsByCreatorRequest{
				Creator:    args[0],
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryVoteForPollByPollVoterCmd creates a CLI command for Query/VoteForPollByPollVoter.
func QueryVoteForPollByPollVoterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote-poll [poll-id] [voter]",
		Short: "Query for vote by poll id and voter account address",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pollID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			queryClient := group.NewQueryClient(clientCtx)

			res, err := queryClient.VoteForPollByPollVoter(cmd.Context(), &group.QueryVoteForPollByPollVoterRequest{
				PollId: pollID,
				Voter:  args[1],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryVotesForPollByPollCmd creates a CLI command for Query/VotesByProposal.
func QueryVotesForPollByPollCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "votes-by-poll [poll-id]",
		Short: "Query for votes by poll id with pagination flags",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pollID, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := group.NewQueryClient(clientCtx)

			res, err := queryClient.VotesForPollByPoll(cmd.Context(), &group.QueryVotesForPollByPollRequest{
				PollId:     pollID,
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryVotesForPollByVoterCmd creates a CLI command for Query/VotesForPollByVoter.
func QueryVotesForPollByVoterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "votes-poll-by-voter [voter]",
		Short: "Query for votes by voter account address with pagination flags",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := group.NewQueryClient(clientCtx)

			res, err := queryClient.VotesForPollByVoter(cmd.Context(), &group.QueryVotesForPollByVoterRequest{
				Voter:      args[0],
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
