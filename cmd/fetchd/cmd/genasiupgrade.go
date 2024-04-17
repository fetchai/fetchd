package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/bech32"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/types"
	"log"
	"os"
)

const (
	ReconcilliationDataPath  = "reconciliation_data.csv"
	ReconciliationAccAddress = "fetch1rhrlzsx9z865dqen8t4v47r99dw6y4va4uph0x" /* "asi1rhrlzsx9z865dqen8t4v47r99dw6y4vaw76rd9" */

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

			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			// replace chain-id
			ASIGenesisUpgradeReplaceChainID(genDoc)

			if err = ASIGenesisUpgradeWithdrawReconciliationBalances(clientCtx.Codec, &appState); err != nil {
				return fmt.Errorf("failed to withdraw reconciliation balances: %w", err)
			}

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

func ASIGenesisUpgradeReplaceDenom() {}

func ASIGenesisUpgradeReplaceAddresses() {}

func ASIGenesisUpgradeWithdrawIBCChannelsBalances() {}

func getGenesisAccountsMap(authGenState *authtypes.GenesisState) (map[string]*authtypes.GenesisAccount, *authtypes.GenesisAccounts, error) {
	accountMap := make(map[string]*authtypes.GenesisAccount)

	accounts, err := authtypes.UnpackAccounts(authGenState.Accounts)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unpack accounts from authtypes.GenState: %w", err)
	}

	for _, account := range accounts {
		// ensure account is valid
		err := account.Validate()
		if err != nil {
			return nil, nil, err
		}

		accountMap[account.GetAddress().String()] = &account
	}

	return accountMap, &accounts, nil
}

func getGenesisBalancesMap(bankGenState *banktypes.GenesisState) (*map[string]int, error) {
	balanceMap := make(map[string]int)

	for i, balance := range bankGenState.Balances {
		balanceMap[balance.Address] = i
	}

	return &balanceMap, nil
}

func ASIGenesisUpgradeWithdrawReconciliationBalances(cdc codec.Codec, appState *map[string]json.RawMessage) error {
	bankGenState := banktypes.GetGenesisStateFromAppState(cdc, *appState)
	balances := bankGenState.Balances
	balanceMap, err := getGenesisBalancesMap(bankGenState)
	if err != nil {
		return err
	}

	authGenState := authtypes.GetGenesisStateFromAppState(cdc, *appState)
	accountMap, _, err := getGenesisAccountsMap(&authGenState)
	if err != nil {
		return err
	}

	file, err := os.Open(ReconcilliationDataPath)
	if err != nil {
		log.Fatalf("Error opening reconciliation data: %s", err)
	}
	defer file.Close()

	r := csv.NewReader(file)
	items, err := r.ReadAll()
	if err != nil {
		log.Fatalf("Error reading reconciliation data: %s", err)
	}

	//reconciliationAccount, ok := accountMap[ReconciliationAccAddress]
	//if !ok {
	//	return fmt.Errorf("no genesis match for reconciliation address: %s", ReconciliationAccAddress)
	//}

	reconciliationBalanceIdx, ok := (*balanceMap)[ReconciliationAccAddress]
	if !ok {
		return fmt.Errorf("no genesis match for reconciliation address: %s", ReconciliationAccAddress)
	}

	for _, row := range items {
		addr := row[2]

		//_ = row[3] balance from CSV

		// TODO: use this conversion when merging with address replacement branch
		//addr, err := convertAddressToASI(oldAddr, AccAddressPrefix)
		//if err != nil {
		//	return fmt.Errorf("failed to convert address: %w", err)
		//}
		//if _, err = sdk.AccAddressFromBech32(addr); err != nil {
		//	return fmt.Errorf("converted address is invalid: %w", err)
		//}

		acc, ok := accountMap[addr]
		if !ok {
			return fmt.Errorf("no genesis match for reconciliation address: %s", addr)
		}

		balanceIdx, ok := (*balanceMap)[addr]
		if !ok {
			continue
		}

		accBalance := balances[balanceIdx]

		// check if the reconciliation address is still dormant and contains funds
		if (*acc).GetSequence() != 0 && !accBalance.Coins.IsAllPositive() {
			fmt.Println("Reconciliation address is not dormant or has no funds, skipping withdrawal")
			continue
		}

		// withdraw funds from the reconciliation address
		balances[reconciliationBalanceIdx].Coins = balances[reconciliationBalanceIdx].Coins.Add(accBalance.Coins...)

		// zero out the other account's balance
		balances[balanceIdx].Coins = sdk.NewCoins()
	}

	// update the bank genesis state
	(*bankGenState).Balances = balances

	bankGenStateBytes, err := cdc.MarshalJSON(bankGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal auth genesis state: %w", err)
	}

	(*appState)[banktypes.ModuleName] = bankGenStateBytes

	return nil
}

func convertAddressToASI(addr string, addressPrefix string) (string, error) {
	_, decodedAddrData, err := bech32.Decode(addr)
	if err != nil {
		return "", err
	}

	newAddress, err := bech32.Encode(NewAddrPrefix+addressPrefix, decodedAddrData)
	if err != nil {
		return "", err
	}

	err = sdk.VerifyAddressFormat(decodedAddrData)
	if err != nil {
		return "", err
	}

	return newAddress, nil
}
