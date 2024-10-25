package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
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
		Long:  `This command retrieves all balance information for a given address from manifest, including the amount delegated to validators, and rewards.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := client.GetClientContextFromCmd(cmd)

			manifestFilePath := args[0]
			address := args[1]

			// Call a function to extract address info
			err := ManifestAddressInfo(manifestFilePath, address, ctx)
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

	err = app.VerifyConfig(cudosConfig, genesisData.Prefix, app.AccountAddressPrefix)
	if err != nil {
		return err
	}

	// Verify extra supply
	bondDenomSourceTotalSupply := genesisData.TotalSupply.AmountOf(genesisData.BondDenom)
	if cudosConfig.Config.TotalCudosSupply.LT(bondDenomSourceTotalSupply) {
		return fmt.Errorf("total supply %s from config is smaller than total supply %s in genesis", cudosConfig.Config.TotalCudosSupply.String(), bondDenomSourceTotalSupply.String())
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

type MigratedBalance struct {
	Address       string    `json:"address"`
	SourceBalance sdk.Coins `json:"src_balance,omitempty"`
	DestBalance   sdk.Coins `json:"dest_balance,omitempty"`
}

type DelegationsAggrEntry struct {
	Address      string  `json:"address"`
	SourceTokens sdk.Int `json:"src_amount,omitempty"`
	DestTokens   sdk.Int `json:"dest_amount,omitempty"`
}

type ManifestData struct {
	InitialBalances *app.OrderedMap[string, app.UpgradeBalances]
	MovedBalances   *app.OrderedMap[string, app.UpgradeBalances]

	MigratedBalances     *app.OrderedMap[string, []app.UpgradeBalanceMovement]
	MigratedBalancesAggr *app.OrderedMap[string, MigratedBalance]

	Delegations     *app.OrderedMap[string, []app.UpgradeDelegation]
	DelegationsAggr *app.OrderedMap[string, DelegationsAggrEntry]

	sourcePrefix string
}

func parseManifestData(manifest *app.UpgradeManifest) (*ManifestData, error) {
	manifestData := ManifestData{}

	manifestData.InitialBalances = app.NewOrderedMap[string, app.UpgradeBalances]()
	for _, initialBalanceEntry := range manifest.InitialBalances {
		manifestData.InitialBalances.SetNew(initialBalanceEntry.Address, initialBalanceEntry)

		if manifestData.sourcePrefix == "" {
			prefix, _, err := bech32.DecodeAndConvert(initialBalanceEntry.Address)
			if err != nil {
				return nil, err
			}
			manifestData.sourcePrefix = prefix
		}
	}

	manifestData.MovedBalances = app.NewOrderedMap[string, app.UpgradeBalances]()
	for _, movedBalanceEntry := range manifest.MovedBalances {
		manifestData.MovedBalances.SetNew(movedBalanceEntry.Address, movedBalanceEntry)
	}

	// Make map of minted destination chain balances
	manifestData.MigratedBalances = app.NewOrderedMap[string, []app.UpgradeBalanceMovement]()
	manifestData.MigratedBalancesAggr = app.NewOrderedMap[string, MigratedBalance]()
	for _, migrationEntry := range manifest.Migration.Migrations {
		convertedAddr, err := app.ConvertAddressPrefix(migrationEntry.To, manifestData.sourcePrefix)
		if err != nil {
			return nil, err
		}

		MigratedBalances, _ := manifestData.MigratedBalances.GetOrSetDefault(convertedAddr, []app.UpgradeBalanceMovement{})
		manifestData.MigratedBalances.Set(convertedAddr, append(MigratedBalances, migrationEntry))

		MigratedBalancesAggr, _ := manifestData.MigratedBalancesAggr.GetOrSetDefault(convertedAddr, MigratedBalance{})
		MigratedBalancesAggr.Address = convertedAddr
		MigratedBalancesAggr.SourceBalance = MigratedBalancesAggr.SourceBalance.Add(migrationEntry.SourceBalance...)
		MigratedBalancesAggr.DestBalance = MigratedBalancesAggr.DestBalance.Add(migrationEntry.DestBalance...)
		manifestData.MigratedBalancesAggr.Set(convertedAddr, MigratedBalancesAggr)
	}

	manifestData.Delegations = app.NewOrderedMap[string, []app.UpgradeDelegation]()
	manifestData.DelegationsAggr = app.NewOrderedMap[string, DelegationsAggrEntry]()
	for _, DelegationsEntry := range manifest.Delegate.Delegations {
		convertedAddr, err := app.ConvertAddressPrefix(DelegationsEntry.NewDelegator, manifestData.sourcePrefix)
		if err != nil {
			return nil, err
		}
		Delegations, _ := manifestData.Delegations.GetOrSetDefault(convertedAddr, []app.UpgradeDelegation{})
		manifestData.Delegations.Set(convertedAddr, append(Delegations, DelegationsEntry))

		DelegationsAggr, _ := manifestData.DelegationsAggr.GetOrSetDefault(convertedAddr, DelegationsAggrEntry{SourceTokens: sdk.NewIntFromUint64(0), DestTokens: sdk.NewIntFromUint64(0)})
		DelegationsAggr.Address = convertedAddr
		DelegationsAggr.SourceTokens = DelegationsAggr.SourceTokens.Add(DelegationsEntry.OriginalTokens)
		DelegationsAggr.DestTokens = DelegationsAggr.DestTokens.Add(DelegationsEntry.NewTokens)
		manifestData.DelegationsAggr.Set(convertedAddr, DelegationsAggr)
	}

	return &manifestData, nil
}

func printJSONEntry(upgradeBalances any, ctx client.Context) error {
	data, err := json.MarshalIndent(upgradeBalances, "", "  ")
	if err != nil {
		return err
	}
	err = ctx.PrintString(fmt.Sprintf("%s\n", string(data)))
	if err != nil {
		return err
	}
	return nil
}

func ManifestAddressInfo(manifestFilePath string, address string, ctx client.Context) error {
	manifest, err := app.LoadManifestFromPath(manifestFilePath)
	if err != nil {
		return err
	}

	manifestData, err := parseManifestData(manifest)
	if err != nil {
		return err
	}

	err = ctx.PrintString("Frontend records:\n")
	if err != nil {
		return err
	}

	err = ctx.PrintString("Initial balances:\n")
	if err != nil {
		return err
	}

	address, err = app.ConvertAddressPrefix(address, manifestData.sourcePrefix)
	if err != nil {
		return err
	}

	if InitialBalances, exists := manifestData.InitialBalances.Get(address); exists {
		err = printJSONEntry(InitialBalances, ctx)
		if err != nil {
			return err
		}
	}

	err = ctx.PrintString("Moved balances:\n")
	if err != nil {
		return err
	}

	if MovedBalances, exists := manifestData.MovedBalances.Get(address); exists {
		err = printJSONEntry(MovedBalances, ctx)
		if err != nil {
			return err
		}
	}

	err = ctx.PrintString("Backend records:\n")
	if err != nil {
		return err
	}

	err = ctx.PrintString("Migrated balances:\n")
	if err != nil {
		return err
	}

	if MigratedBalances, exists := manifestData.MigratedBalances.Get(address); exists {
		err = printJSONEntry(MigratedBalances, ctx)
		if err != nil {
			return err
		}
	}

	err = ctx.PrintString("Aggregated migrations:\n")
	if err != nil {
		return err
	}
	if MigratedBalancesAggr, exists := manifestData.MigratedBalancesAggr.Get(address); exists {
		err = printJSONEntry(MigratedBalancesAggr, ctx)
		if err != nil {
			return err
		}
	}

	err = ctx.PrintString("Created delegations:\n")
	if err != nil {
		return err
	}

	if Delegations, exists := manifestData.Delegations.Get(address); exists {
		err = printJSONEntry(Delegations, ctx)
		if err != nil {
			return err
		}
	}

	err = ctx.PrintString("Aggregated delegations:\n")
	if err != nil {
		return err
	}
	if DelegationsAggr, exists := manifestData.DelegationsAggr.Get(address); exists {
		err = printJSONEntry(DelegationsAggr, ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
