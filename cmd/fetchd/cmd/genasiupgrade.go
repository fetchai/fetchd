package cmd

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/types"
)

const (
	flagNewDescription = "new-description"
	Bech32Chars        = "023456789acdefghjklmnpqrstuvwxyz"
	AddrDataLength     = 32
	WasmDataLength     = 52
	AddrChecksumLength = 6
	AccAddressPrefix   = ""
	ValAddressPrefix   = "valoper"
	ConsAddressPrefix  = "valcons"

	NewBaseDenom   = "asi"
	NewDenom       = "aasi"
	NewAddrPrefix  = "asi"
	NewChainId     = "asi-1"
	NewDescription = "ASI Token"

	OldBaseDenom  = "fet"
	OldDenom      = "afet"
	OldAddrPrefix = "fetch"
)

// ASIGenesisUpgradeCmd returns replace-genesis-values cobra Command.
func ASIGenesisUpgradeCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asi-genesis-upgrade",
		Short: "This command carries out a full upgrade of the genesis file to the new ASI chain parameters.",
		Long: `The following command will upgrade the current genesis file to the new ASI chain parameters. The following changes will be made:
              - Chain ID will be updated to "asi-1"
              - The native coin denom will be updated to "asi"
              - The address prefix will be updated to "asi"
              - The old fetch addresses will be updated to the new asi addresses`,

		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			genFile := config.GenesisFile()

			_, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			// replace chain-id
			ASIGenesisUpgradeReplaceChainID(genDoc)
			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")
	cmd.Flags().String(flagNewDescription, "", "The new description for the native coin in the genesis file")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func ASIGenesisUpgradeReplaceDenomMetadata() {}

func ASIGenesisUpgradeReplaceChainID(genesisData *types.GenesisDoc) {
	genesisData.ChainID = NewChainId
}

func ASIGenesisUpgradeReplaceBridgeAdmin() {}

func ASIGenesisUpgradeReplaceDenom() {}

func ASIGenesisUpgradeReplaceAddresses() {}

func ASIGenesisUpgradeWithdrawIBCChannelsBalances() {}

func ASIGenesisUpgradeWithdrawReconciliationBalances() {}
