package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/fetchai/fetchd/x/did/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// this line is used by starport scaffolding # 1
	cmd.AddCommand(
		NewCreateDidDocumentCmd(),
		NewAddVerificationCmd(),
		NewAddServiceCmd(),
		NewRevokeVerificationCmd(),
		NewDeleteServiceCmd(),
		NewSetVerificationRelationshipCmd(),
		NewLinkAriesAgentCmd(),
		NewAddControllerCmd(),
		NewDeleteControllerCmd(),
	)

	return cmd
}

// deriveVMType derive the verification method type from a public key
func deriveVMType(pubKey cryptotypes.PubKey) (vmType types.VerificationMaterialType, err error) {
	switch pubKey.(type) {
	case *ed25519.PubKey:
		vmType = types.DIDVMethodTypeEd25519VerificationKey2018
	case *secp256k1.PubKey:
		vmType = types.DIDVMethodTypeEcdsaSecp256k1VerificationKey2019
	default:
		err = types.ErrKeyFormatNotSupported
	}
	return
}

// NewCreateDidDocumentCmd defines the command to create a new IBC light client.
func NewCreateDidDocumentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create-did [id]",
		Short:   "create decentralized did (did) document",
		Example: "creates a did document for users",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			// did
			did := types.NewChainDID(clientCtx.ChainID, args[0])
			// verification
			signer := clientCtx.GetFromAddress()
			// pubkey
			info, err := clientCtx.Keyring.KeyByAddress(signer)
			if err != nil {
				return err
			}
			pubKey, err := info.GetPubKey()
			if err != nil {
				return err
			}
			// verification method id
			vmID := did.NewVerificationMethodID(signer.String())
			// understand the vmType
			vmType, err := deriveVMType(pubKey)
			if err != nil {
				return err
			}
			auth := types.NewVerification(
				types.NewVerificationMethod(
					vmID,
					did,
					types.NewPublicKeyMultibase(pubKey.Bytes(), vmType),
				),
				[]string{types.Authentication},
				nil,
			)
			// create the message
			msg := types.NewMsgCreateDidDocument(
				did.String(),
				types.Verifications{auth},
				types.Services{},
				signer.String(),
			)
			// validate
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			// execute
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewAddVerificationCmd define the command to add a verification message
func NewAddVerificationCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "add-verification-method [id] [pubkey]",
		Short:   "add an verification method to a decentralized did (did) document",
		Example: "adds an verification method for a did document",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			// signer address
			signer := clientCtx.GetFromAddress()
			// public key
			var pk cryptotypes.PubKey
			err = clientCtx.Codec.UnmarshalInterfaceJSON([]byte(args[1]), &pk)
			if err != nil {
				return err
			}
			// derive the public key type
			vmType, err := deriveVMType(pk)
			if err != nil {
				return err
			}
			// document did
			did := types.NewChainDID(clientCtx.ChainID, args[0])
			// verification method id
			vmID := did.NewVerificationMethodID(sdk.MustBech32ifyAddressBytes(
				sdk.GetConfig().GetBech32AccountAddrPrefix(),
				pk.Address().Bytes(),
			))

			verification := types.NewVerification(
				types.NewVerificationMethod(
					vmID,
					did,
					types.NewPublicKeyMultibase(pk.Bytes(), vmType),
				),
				[]string{types.Authentication},
				nil,
			)
			// add verification
			msg := types.NewMsgAddVerification(
				did.String(),
				verification,
				signer.String(),
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewAddServiceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-service [id] [service_id] [type] [endpoint]",
		Short:   "add a service to a decentralized did (did) document",
		Example: "adds a service to a did document",
		Args:    cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// tx signer
			signer := clientCtx.GetFromAddress()
			// service parameters
			serviceID, serviceType, endpoint := args[1], args[2], args[3]
			// document did
			did := types.NewChainDID(clientCtx.ChainID, args[0])

			service := types.NewService(
				serviceID,
				serviceType,
				endpoint,
			)

			msg := types.NewMsgAddService(
				did.String(),
				service,
				signer.String(),
			)
			// broadcast
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewRevokeVerificationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revoke-verification-method [did_id] [verification_method_id_fragment]",
		Short: "revoke a verification method from a decentralized did (did) document",
		Example: `cosmos-cashd tx did revoke-verification-method 575d062c-d110-42a9-9c04-cb1ff8c01f06 \
 Z46DAL1MrJlVW_WmJ19WY8AeIpGeFOWl49Qwhvsnn2M \
 --from alice \
 --node https://rpc.cosmos-cash.app.beta.starport.cloud:443 --chain-id cosmoscash-testnet`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			// document did
			did := types.NewChainDID(clientCtx.ChainID, args[0])
			// signer
			signer := clientCtx.GetFromAddress()
			// verification method id
			vmID := did.NewVerificationMethodID(args[1])
			// build the message
			msg := types.NewMsgRevokeVerification(
				did.String(),
				vmID,
				signer.String(),
			)
			// validate
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			// broadcast
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewDeleteServiceCmd deletes a service from a DID Document
func NewDeleteServiceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete-service [id] [service-id]",
		Short:   "deletes a service from a decentralized did (did) document",
		Example: "delete a service for a did document",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			// document did
			did := types.NewChainDID(clientCtx.ChainID, args[0])
			// signer
			signer := clientCtx.GetFromAddress()
			// service id
			sID := args[1]

			msg := types.NewMsgDeleteService(
				did.String(),
				sID,
				signer.String(),
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// NewAddControllerCmd adds a controller to a did document
func NewAddControllerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add-controller [id] [controllerAddress]",
		Short:   "updates a decentralized identifier (did) document to contain a controller",
		Example: "add-controller vasp cosmos1kslgpxklq75aj96cz3qwsczr95vdtrd3p0fslp",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			// document did
			did := types.NewChainDID(clientCtx.ChainID, args[0])

			// did key to use as the controller
			didKey := types.NewKeyDID(args[1])

			// signer
			signer := clientCtx.GetFromAddress()

			msg := types.NewMsgAddController(
				did.String(),
				didKey.String(),
				signer.String(),
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewDeleteControllerCmd adds a controller to a did document
func NewDeleteControllerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete-controller [id] [controllerAddress]",
		Short:   "updates a decentralized identifier (did) document removing a controller",
		Example: "delete-controller vasp cosmos1kslgpxklq75aj96cz3qwsczr95vdtrd3p0fslp",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			// document did
			did := types.NewChainDID(clientCtx.ChainID, args[0])

			// did key to use as the controller
			didKey := types.NewKeyDID(args[1])

			// signer
			signer := clientCtx.GetFromAddress()

			msg := types.NewMsgDeleteController(
				did.String(),
				didKey.String(),
				signer.String(),
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewSetVerificationRelationshipCmd adds a verification relationship to a verification method
func NewSetVerificationRelationshipCmd() *cobra.Command {

	// relationships
	var relationships []string
	// if true do not add the default authentication relationship
	var unsafe bool

	cmd := &cobra.Command{
		Use:     "set-verification-relationship [did_id] [verification_method_id_fragment] --relationship NAME [--relationship NAME ...]",
		Short:   "sets one or more verification relationships to a key on a decentralized identifier (did) document.",
		Example: "set-verification-relationship vasp 6f1e0700-6c86-41b6-9e05-ae3cf839cdd0 --relationship capabilityInvocation",
		Args:    cobra.ExactArgs(2),

		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			// document did
			did := types.NewChainDID(clientCtx.ChainID, args[0])

			// method id
			vmID := did.NewVerificationMethodID(args[1])

			// signer
			signer := clientCtx.GetFromAddress()

			msg := types.NewMsgSetVerificationRelationships(
				did.String(),
				vmID,
				relationships,
				signer.String(),
			)

			// make sure that the authentication relationship is preserved
			if !unsafe {
				msg.Relationships = append(msg.Relationships, types.Authentication)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	// add flags to set did relationships
	cmd.Flags().StringSliceVarP(&relationships, "relationship", "r", []string{}, "the relationships to set for the verification method in the DID")
	cmd.Flags().BoolVar(&unsafe, "unsafe", false, fmt.Sprint("do not ensure that '", types.Authentication, "' relationship is set"))

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
