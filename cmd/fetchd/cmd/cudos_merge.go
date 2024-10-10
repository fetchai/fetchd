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

func utilNetworkMergeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "network-merge",
		Short:                      "Network merge commands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmdVerify := &cobra.Command{
		Use:   "verify-config [network_merge_config_json_file_path] [source_chain_genesis_json_file_path]",
		Short: "Verifies the configuration JSON file of the network merge",
		Long: `This command verifies the structure and content of the network merge config JSON file. 
It checks whether the network merge config file conforms to expected schema - presence of all required fields and validates their values against predefined rules.
`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := client.GetClientContextFromCmd(cmd)

			configFilePath := args[0]
			GenesisFilePath := args[1]

			// Read and verify the JSON file
			var err error
			if err = VerifyConfigFile(configFilePath, GenesisFilePath, ctx); err != nil {
				return err
			}

			return ctx.PrintString("Configuration is valid.\n")
		},
	}

	cmd.AddCommand(cmdVerify)
	return cmd
}

// VerifyConfigFile validates the content of a JSON configuration file.
func VerifyConfigFile(configFilePath string, GenesisFilePath string, ctx client.Context) error {
	manifest := app.NewUpgradeManifest()

	destinationChainPrefix := "fetch"

	networkInfo, configBytes, err := app.LoadNetworkConfigFromFile(configFilePath)
	if err != nil {
		return err
	}
	err = ctx.PrintString(fmt.Sprintf("Config hash: %s\n", app.GenerateSha256Hex(*configBytes)))
	if err != nil {
		return err
	}

	_, GenDoc, err := genutiltypes.GenesisStateFromGenFile(GenesisFilePath)
	if err != nil {
		return fmt.Errorf("failed to unmarshal genesis state: %w", err)
	}

	// unmarshal the app state
	var cudosJsonData map[string]interface{}
	if err = json.Unmarshal(GenDoc.AppState, &cudosJsonData); err != nil {
		return fmt.Errorf("failed to unmarshal app state: %w", err)
	}
	cudosConfig := app.NewCudosMergeConfig(networkInfo.CudosMerge)

	genesisData, err := app.ParseGenesisData(cudosJsonData, GenDoc, cudosConfig, manifest)
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
	err = app.VerifyAddressPrefix(networkInfo.CudosMerge.IbcTargetAddr, genesisData.Prefix)
	if err != nil {
		return fmt.Errorf("ibc targer address error: %v", err)
	}

	err = app.VerifyAddressPrefix(networkInfo.CudosMerge.RemainingStakingBalanceAddr, genesisData.Prefix)
	if err != nil {
		return fmt.Errorf("remaining staking balance address error: %v", err)
	}

	err = app.VerifyAddressPrefix(networkInfo.CudosMerge.RemainingGravityBalanceAddr, genesisData.Prefix)
	if err != nil {
		return fmt.Errorf("remaining gravity balance address error: %v", err)
	}

	err = app.VerifyAddressPrefix(networkInfo.CudosMerge.RemainingDistributionBalanceAddr, genesisData.Prefix)
	if err != nil {
		return fmt.Errorf("remaining distribution balance address error: %v", err)
	}

	err = app.VerifyAddressPrefix(networkInfo.CudosMerge.ContractDestinationFallbackAddr, genesisData.Prefix)
	if err != nil {
		return fmt.Errorf("contract destination fallback address error: %v", err)
	}

	// Community pool address is optional
	if networkInfo.CudosMerge.CommunityPoolBalanceDestAddr != "" {
		err = app.VerifyAddressPrefix(networkInfo.CudosMerge.CommunityPoolBalanceDestAddr, genesisData.Prefix)
		if err != nil {
			return fmt.Errorf("community pool balance destination address error: %v", err)
		}
	}

	err = app.VerifyAddressPrefix(networkInfo.CudosMerge.CommissionFetchAddr, destinationChainPrefix)
	if err != nil {
		return fmt.Errorf("comission address error: %v", err)
	}

	err = app.VerifyAddressPrefix(networkInfo.CudosMerge.ExtraSupplyFetchAddr, destinationChainPrefix)
	if err != nil {
		return fmt.Errorf("extra supply address error: %v", err)
	}

	err = app.VerifyAddressPrefix(networkInfo.CudosMerge.VestingCollisionDestAddr, genesisData.Prefix)
	if err != nil {
		return fmt.Errorf("vesting collision destination address error: %v", err)
	}

	// Verify extra supply
	bondDenomSourceTotalSupply := genesisData.TotalSupply.AmountOf(genesisData.BondDenom)
	if networkInfo.CudosMerge.TotalCudosSupply.LT(bondDenomSourceTotalSupply) {
		return fmt.Errorf("total supply %s from config is smaller than total supply %s in genesis", networkInfo.CudosMerge.TotalCudosSupply.String(), bondDenomSourceTotalSupply.String())
	}

	if len(networkInfo.CudosMerge.BalanceConversionConstants) == 0 {
		return fmt.Errorf("list of conversion constants is empty")
	}

	if len(networkInfo.CudosMerge.BackupValidators) == 0 {
		return fmt.Errorf("list of backup validators is empty")
	}

	return nil
}
