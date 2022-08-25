package cli

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/spf13/cobra"

	"github.com/fetchai/fetchd/x/verifiable-credential/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group verifiable-credential queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// this line is used by starport scaffolding # 1
	cmd.AddCommand(
		GetCmdQueryVerifiableCredentials(),
		GetCmdQueryVerifiableCredential(),
		GetCmdQueryValidateVerifiableCredential(),
	)

	return cmd
}

func GetCmdQueryVerifiableCredentials() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verifiable-credentials",
		Short: "Query for all verifiable credentials",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			result, err := queryClient.VerifiableCredentials(
				context.Background(),
				&types.QueryVerifiableCredentialsRequest{
					Pagination: pageReq,
				},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(result)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryVerifiableCredential implements the VerifiableCredential query command.
func GetCmdQueryVerifiableCredential() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verifiable-credential [verifiable-credential-id]",
		Short: "Query a verifiable-credential",
		Long:  `Query details about an individual verifiable-credential.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryVerifiableCredentialRequest{VerifiableCredentialId: args[0]}
			res, err := queryClient.VerifiableCredential(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryValidateVerifiableCredential implements the VerifiableCredential query command.
func GetCmdQueryValidateVerifiableCredential() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate-verifiable-credential [verifiable-credential-id] [pubkey]",
		Short: "Validate a verifiable-credential",
		Long:  `Validate proof for an individual verifiable-credential.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			// query to get the verifiable credential
			params := &types.QueryVerifiableCredentialRequest{VerifiableCredentialId: args[0]}
			res, err := queryClient.VerifiableCredential(cmd.Context(), params)
			if err != nil {
				return err
			}

			// check the returned credential is signed by the provided pubkey
			var pk cryptotypes.PubKey
			err = clientCtx.Codec.UnmarshalInterfaceJSON([]byte(args[1]), &pk)
			if err != nil {
				return err
			}

			err = res.VerifiableCredential.Validate(pk)
			isValid := false
			if err == nil {
				isValid = true
			}

			result := &types.QueryValidateVerifiableCredentialResponse{
				IsValid: isValid,
			}

			return clientCtx.PrintProto(result)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
