package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/fetchai/fetchd/app"
	"github.com/spf13/cobra"
)

// Module init related flags
const (
	FlagCudosGenesisPath           = "cudos-genesis-path"
	FlagCudosGenesisSha256         = "cudos-genesis-sha256"
	FlagCudosMigrationConfigPath   = "cudos-migration-config-path"
	FlagCudosMigrationConfigSha256 = "cudos-migration-config-sha256"
)

func AddCudosFlags(startCmd *cobra.Command) {
	startCmd.Flags().String(FlagCudosGenesisPath, "", "Cudos genesis file path. Required to be provided *exclusively* during cudos migration upgrade node start, *ignored* on all subsequent node starts.")
	startCmd.Flags().String(FlagCudosMigrationConfigPath, "", "Upgrade config file path. Required to be provided *exclusively* during cudos migration upgrade node start, *ignored* on all subsequent node starts.")
	startCmd.Flags().String(FlagCudosGenesisSha256, "", "Sha256 of the cudos genesis file. Optional to be provided *exclusively* during cudos migration upgrade node start, *ignored* on all subsequent node starts.")
	startCmd.Flags().String(FlagCudosMigrationConfigSha256, "", fmt.Sprintf("Sha256 of the upgrade config file. Required if to be provided *exclusively* during cudos migration upgrade node start and *only IF* \"%v\" flag has been provided, *ignored* on all subsequent node starts.", FlagCudosMigrationConfigPath))

	// Capture the existing PreRunE function
	existingPreRunE := startCmd.PreRunE

	// Set a new PreRunE function that includes the old one
	startCmd.PreRunE = func(cmd *cobra.Command, args []string) error {
		// Check for positional arguments
		if len(args) > 0 {
			return fmt.Errorf("no positional arguments are allowed")
		}

		// Call the existing PreRunE function if it exists
		if existingPreRunE != nil {
			if err := existingPreRunE(cmd, args); err != nil {
				return err
			}
		}

		return nil
	}

}

func utilCudosCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "cudos",
		Short:                      "Cudos commands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmdVerify := &cobra.Command{
		Use:   "verify-config [config_json_file_path]",
		Short: "Verifies the configuration JSON file",
		Long:  "This command verifies the structure and content of the configuration JSON file. It checks if all required fields are present and validates their values against predefined rules.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := client.GetClientContextFromCmd(cmd)

			configFilePath := args[0]
			cudosGenesisFilePath := args[1]

			// Read and verify the JSON file
			var err error
			if err = VerifyConfigFile(configFilePath, cudosGenesisFilePath, ctx); err != nil {
				return err
			}

			return ctx.PrintString("Configuration is valid.")
		},
	}

	cmd.AddCommand(cmdVerify)
	return cmd
}

// VerifyConfigFile validates the content of a JSON configuration file.
func VerifyConfigFile(configFilePath string, cudosGenesisFilePath string, ctx client.Context) error {
	manifest := app.NewUpgradeManifest()

	networkInfo, configBytes, err := app.LoadNetworkConfigFromFile(configFilePath)
	if err != nil {
		return err
	}

	_, cudosGenDoc, err := genutiltypes.GenesisStateFromGenFile(cudosGenesisFilePath)
	if err != nil {
		return fmt.Errorf("cudos merge: failed to unmarshal genesis state: %w", err)
	}

	// unmarshal the app state
	var cudosJsonData map[string]interface{}
	if err = json.Unmarshal(cudosGenDoc.AppState, &cudosJsonData); err != nil {
		return fmt.Errorf("cudos merge: failed to unmarshal app state: %w", err)
	}
	cudosConfig := app.NewCudosMergeConfig(networkInfo.CudosMerge)

	genesisData, err := app.ParseGenesisData(cudosJsonData, cudosGenDoc, cudosConfig, manifest)
	if err != nil {
		return fmt.Errorf("failed to parse genesis data: %w", err)
	}

	if networkInfo.MergeSourceChainID != genesisData.ChainId {
		return fmt.Errorf("source chain id %s is different from config chain id %s", networkInfo.MergeSourceChainID, genesisData.ChainId)
	}

	// We don't have access to home folder here so we can't check
	if networkInfo.DestinationChainID == "" {
		return fmt.Errorf("destination chain id is empty")
	}

	// Verify addresses
	err = app.VerifyAddressPrefix(networkInfo.CudosMerge.RemainingDistributionBalanceAddr, genesisData.Prefix)
	if err != nil {
		return fmt.Errorf("remaining distribution balance address prefix error: %v", err)
	}

	err = ctx.PrintString(fmt.Sprintf("Config hash: %s", app.GenerateSha256Hex(*configBytes)))
	if err != nil {
		return err
	}

	println(networkInfo)

	return nil
}
