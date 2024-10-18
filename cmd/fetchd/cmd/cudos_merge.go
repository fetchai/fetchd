package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/fetchai/fetchd/app"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/log"
	"os"
)

// Module init related flags
const (
	FlagCudosGenesisPath           = "cudos-genesis-path"
	FlagCudosGenesisSha256         = "cudos-genesis-sha256"
	FlagCudosMigrationConfigPath   = "cudos-migration-config-path"
	FlagCudosMigrationConfigSha256 = "cudos-migration-config-sha256"

	FlagManifestDestinationPath = "manifest-destination-path"

	DestinationChainPrefix = "fetch"
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

func AddCommandVerify(networkMergeCmd *cobra.Command) {

	cmd := &cobra.Command{
		Use:   "verify-config [network_merge_config_json_file_path] [source_chain_genesis_json_file_path]",
		Short: "Verifies the configuration JSON file of the network merge",
		Long: `This command verifies the structure and content of the network merge config JSON file. 
It checks whether the network merge config file conforms to expected schema - presence of all required fields and validates their values against predefined rules.
` +
			"Verification fully executes the front-end of the upgrade procedure using source chain genesis json and " +
			"network config files as inputs, constructing front-end cache containing all necessary structures and " +
			"derived data, exactly as it would be executed during the real upgrade.",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := client.GetClientContextFromCmd(cmd)

			configFilePath := args[0]
			GenesisFilePath := args[1]

			manifestFilePath, err := cmd.Flags().GetString(FlagManifestDestinationPath)
			if err != nil {
				return err
			}

			// Read and verify the JSON file
			if err = VerifyConfigFile(configFilePath, GenesisFilePath, ctx, manifestFilePath); err != nil {
				return err
			}

			return ctx.PrintString("Configuration is valid.\n")
		},
	}
	cmd.Flags().String(FlagManifestDestinationPath, "", "Save manifest to specified file if set")
	flags.AddQueryFlagsToCmd(cmd)

	networkMergeCmd.AddCommand(cmd)
}

func AddCommandExtractAddressInfo(networkMergeCmd *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "extract-address-info [network_merge_config_json_file_path] [source_chain_genesis_json_file_path] [address]",
		Short: "Extracts balance information for a specific address",
		Long: `This command retrieves all balance information for a given address, including the amount delegated to validators, rewards, and other relevant data.
It utilizes the provided network merge config and genesis JSON files to perform the extraction and display the results.`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := client.GetClientContextFromCmd(cmd)

			configFilePath := args[0]
			GenesisFilePath := args[1]
			address := args[2]

			// Call a function to extract address info
			err := ExtractAddressInfo(configFilePath, GenesisFilePath, address, ctx)
			if err != nil {
				return err
			}
			return nil
		},
	}

	networkMergeCmd.AddCommand(cmd)
}

func AddCommandManifestAddressInfo(networkMergeCmd *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "manifest-address-info [manifest_file_path] [address]",
		Short: "Extracts balance information for a specific address",
		Long:  `This command retrieves all balance information for a given address from almanac, including the amount delegated to validators, and rewards.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := client.GetClientContextFromCmd(cmd)

			manifestFilePath := args[0]
			address := args[1]

			// Call a function to extract address info
			err := AlmanacAddressInfo(manifestFilePath, address, ctx)
			if err != nil {
				return err
			}
			return nil
		},
	}

	networkMergeCmd.AddCommand(cmd)
}

func utilNetworkMergeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "network-merge",
		Short:                      "Network merge commands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	AddCommandVerify(cmd)
	AddCommandExtractAddressInfo(cmd)
	AddCommandManifestAddressInfo(cmd)

	return cmd
}

func LoadGenesisDataFromFile(GenesisFilePath string, cudosConfig *app.CudosMergeConfig, manifest *app.UpgradeManifest) (*app.GenesisData, error) {
	_, GenDoc, err := genutiltypes.GenesisStateFromGenFile(GenesisFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal genesis state: %w", err)
	}

	// unmarshal the app state
	var cudosJsonData map[string]interface{}
	if err = json.Unmarshal(GenDoc.AppState, &cudosJsonData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal app state: %w", err)
	}

	genesisData, err := app.ParseGenesisData(cudosJsonData, GenDoc, cudosConfig, manifest)
	if err != nil {
		return nil, fmt.Errorf("failed to parse genesis data: %w", err)
	}
	return genesisData, nil
}

// VerifyConfigFile validates the content of a JSON configuration file.
func VerifyConfigFile(configFilePath string, GenesisFilePath string, ctx client.Context, manifestFilePath string) error {
	manifest := app.NewUpgradeManifest()

	networkInfo, configBytes, err := app.LoadNetworkConfigFromFile(configFilePath)
	if err != nil {
		return err
	}
	configHashHex := app.GenerateSha256Hex(*configBytes)
	err = ctx.PrintString(fmt.Sprintf("Config hash: %s\n", configHashHex))
	if err != nil {
		return err
	}
	manifest.NetworkConfigFileSha256 = configHashHex

	genesisHashHex, err := app.GenerateSHA256FromFile(GenesisFilePath)
	if err != nil {
		return err
	}
	manifest.GenesisFileSha256 = genesisHashHex

	cudosConfig := app.NewCudosMergeConfig(networkInfo.CudosMerge)
	genesisData, err := LoadGenesisDataFromFile(GenesisFilePath, cudosConfig, manifest)

	if err != nil {
		return fmt.Errorf("failed to load genesis data: %w", err)
	}

	if networkInfo.MergeSourceChainID != genesisData.ChainId {
		return fmt.Errorf("source chain id %s is different from config chain id %s", networkInfo.MergeSourceChainID, genesisData.ChainId)
	}

	// We don't have access to home folder here so we can't check
	if networkInfo.DestinationChainID == "" {
		return fmt.Errorf("destination chain id is empty")
	}

	// Verify addresses
	err = app.VerifyAddressPrefix(cudosConfig.Config.IbcTargetAddr, genesisData.Prefix)
	if err != nil {
		return fmt.Errorf("ibc targer address error: %v", err)
	}

	err = app.VerifyAddressPrefix(cudosConfig.Config.RemainingStakingBalanceAddr, genesisData.Prefix)
	if err != nil {
		return fmt.Errorf("remaining staking balance address error: %v", err)
	}

	err = app.VerifyAddressPrefix(cudosConfig.Config.RemainingGravityBalanceAddr, genesisData.Prefix)
	if err != nil {
		return fmt.Errorf("remaining gravity balance address error: %v", err)
	}

	err = app.VerifyAddressPrefix(cudosConfig.Config.RemainingDistributionBalanceAddr, genesisData.Prefix)
	if err != nil {
		return fmt.Errorf("remaining distribution balance address error: %v", err)
	}

	err = app.VerifyAddressPrefix(cudosConfig.Config.ContractDestinationFallbackAddr, genesisData.Prefix)
	if err != nil {
		return fmt.Errorf("contract destination fallback address error: %v", err)
	}

	// Community pool address is optional
	if cudosConfig.Config.CommunityPoolBalanceDestAddr != "" {
		err = app.VerifyAddressPrefix(cudosConfig.Config.CommunityPoolBalanceDestAddr, genesisData.Prefix)
		if err != nil {
			return fmt.Errorf("community pool balance destination address error: %v", err)
		}
	}

	err = app.VerifyAddressPrefix(cudosConfig.Config.CommissionFetchAddr, DestinationChainPrefix)
	if err != nil {
		return fmt.Errorf("comission address error: %v", err)
	}

	err = app.VerifyAddressPrefix(cudosConfig.Config.ExtraSupplyFetchAddr, DestinationChainPrefix)
	if err != nil {
		return fmt.Errorf("extra supply address error: %v", err)
	}

	err = app.VerifyAddressPrefix(cudosConfig.Config.VestingCollisionDestAddr, genesisData.Prefix)
	if err != nil {
		return fmt.Errorf("vesting collision destination address error: %v", err)
	}

	// Verify extra supply
	bondDenomSourceTotalSupply := genesisData.TotalSupply.AmountOf(genesisData.BondDenom)
	if cudosConfig.Config.TotalCudosSupply.LT(bondDenomSourceTotalSupply) {
		return fmt.Errorf("total supply %s from config is smaller than total supply %s in genesis", cudosConfig.Config.TotalCudosSupply.String(), bondDenomSourceTotalSupply.String())
	}

	if len(cudosConfig.Config.BalanceConversionConstants) == 0 {
		return fmt.Errorf("list of conversion constants is empty")
	}

	if len(cudosConfig.Config.BackupValidators) == 0 {
		return fmt.Errorf("list of backup validators is empty")
	}

	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	err = app.ProcessSourceNetworkGenesis(logger, cudosConfig, genesisData, manifest)
	if err != nil {
		return err
	}

	if manifestFilePath != "" {
		err = app.SaveManifestToPath(manifest, manifestFilePath)
		if err != nil {
			return err
		}

	}

	return nil
}

func ExtractAddressInfo(configFilePath string, GenesisFilePath string, address string, ctx client.Context) error {
	manifest := app.NewUpgradeManifest()

	networkInfo, _, err := app.LoadNetworkConfigFromFile(configFilePath)
	if err != nil {
		return err
	}

	cudosConfig := app.NewCudosMergeConfig(networkInfo.CudosMerge)
	genesisData, err := LoadGenesisDataFromFile(GenesisFilePath, cudosConfig, manifest)
	if err != nil {
		return err
	}

	err = printAccInfo(genesisData, address, ctx)
	if err != nil {
		return err
	}

	/*
		logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
		err = app.ProcessSourceNetworkGenesis(logger, cudosConfig, genesisData, manifest)
		if err != nil {
			return err
		}
	*/

	return nil
}

func printAccInfo(genesisData *app.GenesisData, address string, ctx client.Context) error {
	totalAvailableBalance := sdk.NewCoins()

	accountInfo, exists := genesisData.Accounts.Get(address)
	if !exists {
		err := ctx.PrintString(fmt.Sprintf("Account %s doesn't exist\n", address))
		if err != nil {
			return err
		}
		return nil
	}

	err := ctx.PrintString(fmt.Sprintf("Account type: %s\n", accountInfo.AccountType))
	if err != nil {
		return err
	}

	if accountInfo.Name != "" {
		err = ctx.PrintString(fmt.Sprintf("Account name: %s\n", accountInfo.Name))
		if err != nil {
			return err
		}
	}

	if !accountInfo.OriginalVesting.IsZero() {
		err = ctx.PrintString(fmt.Sprintf("Vested balance: %s\n", accountInfo.OriginalVesting))
		if err != nil {
			return err
		}
	}

	// Get bank balance
	totalAvailableBalance = totalAvailableBalance.Add(accountInfo.Balance...)
	err = ctx.PrintString(fmt.Sprintf("Bank balance: %s\n", accountInfo.Balance))
	if err != nil {
		return err
	}

	// Bonded tokens
	err = ctx.PrintString("Balance in delegations:\n")
	if err != nil {
		return err
	}
	if delegations, exists := genesisData.Delegations.Get(address); exists {
		for i := range delegations.Iterate() {
			validatorAddress, delegatedAmount := i.Key, i.Value
			delegatedBalance := sdk.NewCoin(genesisData.BondDenom, delegatedAmount)
			totalAvailableBalance = totalAvailableBalance.Add(delegatedBalance)
			err = ctx.PrintString(fmt.Sprintf("%s, %s\n", validatorAddress, delegatedBalance))
			if err != nil {
				return err
			}
		}
	}

	// Unbonding tokens
	err = ctx.PrintString("Balance in unbonding delegations:\n")
	if err != nil {
		return err
	}
	if delegations, exists := genesisData.UnbondingDelegations.Get(address); exists {
		for i := range delegations.Iterate() {
			validatorAddress, delegatedAmount := i.Key, i.Value
			delegatedBalance := sdk.NewCoin(genesisData.BondDenom, delegatedAmount)
			totalAvailableBalance = totalAvailableBalance.Add(delegatedBalance)
			err = ctx.PrintString(fmt.Sprintf("%s, %s\n", validatorAddress, delegatedBalance))
			if err != nil {
				return err
			}
		}
	}

	// Unbonded tokens
	err = ctx.PrintString("Balance in unbonded delegations:\n")
	if err != nil {
		return err
	}
	if delegations, exists := genesisData.UnbondedDelegations.Get(address); exists {
		for i := range delegations.Iterate() {
			validatorAddress, delegatedAmount := i.Key, i.Value
			delegatedBalance := sdk.NewCoin(genesisData.BondDenom, delegatedAmount)
			totalAvailableBalance = totalAvailableBalance.Add(delegatedBalance)
			err = ctx.PrintString(fmt.Sprintf("%s, %s\n", validatorAddress, delegatedBalance))
			if err != nil {
				return err
			}
		}
	}

	// Get distribution module rewards
	err = ctx.PrintString("Rewards:\n")
	if err != nil {
		return err
	}

	if DelegatorRewards, exists := genesisData.DistributionInfo.Rewards.Get(address); exists {
		for j := range DelegatorRewards.Iterate() {
			validatorOperatorAddr, rewardDecAmount := j.Key, j.Value
			rewardAmount, _ := rewardDecAmount.TruncateDecimal()
			if !rewardAmount.IsZero() {
				totalAvailableBalance = totalAvailableBalance.Add(rewardAmount...)

				err = ctx.PrintString(fmt.Sprintf("%s, %s\n", validatorOperatorAddr, rewardAmount))
				if err != nil {
					return err
				}
			}

		}
	}

	err = ctx.PrintString(fmt.Sprintf("Total available balance: %s\n", totalAvailableBalance))
	if err != nil {
		return err
	}

	return nil
}

type ManifestData struct {
	InitialBalances *app.OrderedMap[string, app.UpgradeBalances]
	MovedBalances   *app.OrderedMap[string, app.UpgradeBalances]
}

func parseManifestData(manifest *app.UpgradeManifest) (*ManifestData, error) {
	manifestData := ManifestData{}

	manifestData.InitialBalances = app.NewOrderedMap[string, app.UpgradeBalances]()
	for _, initialBalanceEntry := range manifest.InitialBalances {
		manifestData.InitialBalances.SetNew(initialBalanceEntry.Address, initialBalanceEntry)
	}

	manifestData.MovedBalances = app.NewOrderedMap[string, app.UpgradeBalances]()
	for _, movedBalanceEntry := range manifest.MovedBalances {
		manifestData.MovedBalances.SetNew(movedBalanceEntry.Address, movedBalanceEntry)
	}

	return &manifestData, nil
}

func printBalancesEntry(upgradeBalances app.UpgradeBalances, ctx client.Context) error {
	if !upgradeBalances.BankBalance.IsZero() {
		err := ctx.PrintString(fmt.Sprintf("Bank balance: %s\n", upgradeBalances.BankBalance))
		if err != nil {
			return err
		}
	}

	if !upgradeBalances.VestedBalance.IsZero() {
		err := ctx.PrintString(fmt.Sprintf("Vested balance: %s\n", upgradeBalances.VestedBalance))
		if err != nil {
			return err
		}
	}

	if !upgradeBalances.BondedStakingBalance.IsZero() {
		err := ctx.PrintString(fmt.Sprintf("Bonded staking balance: %s\n", upgradeBalances.BondedStakingBalance))
		if err != nil {
			return err
		}
	}

	if !upgradeBalances.UnbondedStakingBalance.IsZero() {
		err := ctx.PrintString(fmt.Sprintf("Unbonded staking balance: %s\n", upgradeBalances.UnbondedStakingBalance))
		if err != nil {
			return err
		}
	}

	if !upgradeBalances.UnbondingStakingBalance.IsZero() {
		err := ctx.PrintString(fmt.Sprintf("Unbonding staking balance: %s\n", upgradeBalances.UnbondingStakingBalance))
		if err != nil {
			return err
		}
	}

	if !upgradeBalances.DistributionRewards.IsZero() {
		err := ctx.PrintString(fmt.Sprintf("Distribution rewards: %s\n", upgradeBalances.DistributionRewards))
		if err != nil {
			return err
		}
	}

	return nil
}

func AlmanacAddressInfo(almanacFilePath string, address string, ctx client.Context) error {
	manifest := app.NewUpgradeManifest()

	manifest, err := app.LoadManifestFromPath(almanacFilePath)
	if err != nil {
		return err
	}

	manifestData, err := parseManifestData(manifest)
	if err != nil {
		return err
	}

	err = ctx.PrintString("Initial balances:\n")
	if err != nil {
		return err
	}

	if InitialBalances, exists := manifestData.InitialBalances.Get(address); exists {
		err = printBalancesEntry(InitialBalances, ctx)
		if err != nil {
			return err
		}
	}

	err = ctx.PrintString("Moved balances:\n")
	if err != nil {
		return err
	}

	if MovedBalances, exists := manifestData.MovedBalances.Get(address); exists {
		err = printBalancesEntry(MovedBalances, ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
