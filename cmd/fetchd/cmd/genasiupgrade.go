package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/bech32"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	ibc_core "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	ibctypes "github.com/cosmos/ibc-go/v3/modules/core/types"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/types"
)

const (
	IbcWithdrawAddress = "fetch1rhrlzsx9z865dqen8t4v47r99dw6y4va4uph0x" /* "asi1rhrlzsx9z865dqen8t4v47r99dw6y4vaw76rd9" */

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
			cdc := clientCtx.Codec

			config.SetRoot(clientCtx.HomeDir)

			genFile := config.GenesisFile()

			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			// replace chain-id
			ASIGenesisUpgradeReplaceChainID(genDoc)

			if err = ASIGenesisUpgradeWithdrawIBCChannelsBalances(&cdc, &appState); err != nil {
				return fmt.Errorf("failed to withdraw IBC channels balances: %w", err)
			}

			// reflect changes in the genesis file
			var modifiedGenState json.RawMessage
			if modifiedGenState, err = json.Marshal(appState); err != nil {
				return fmt.Errorf("failed to marshal app state: %w", err)
			}

			(*genDoc).AppState = modifiedGenState
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

func ASIGenesisUpgradeWithdrawIBCChannelsBalances(cdc *codec.Codec, appState *map[string]json.RawMessage) error {
	bankGenState := banktypes.GetGenesisStateFromAppState(*cdc, *appState)
	balances := bankGenState.Balances
	balanceMap, err := getGenesisBalancesMap(bankGenState)
	if err != nil {
		return err
	}

	var ibcAppState ibctypes.GenesisState
	if err := (*cdc).UnmarshalJSON((*appState)[ibc_core.ModuleName], &ibcAppState); err != nil {
		return fmt.Errorf("failed to unmarshal IBC genesis state: %w", err)
	}

	withdrawalBalanceIdx, ok := (*balanceMap)[IbcWithdrawAddress]
	if !ok {
		return fmt.Errorf("failed to find withdrawal address in genesis balances")
	}

	for _, channel := range ibcAppState.ChannelGenesis.Channels {
		rawAddr := ibctransfertypes.GetEscrowAddress(channel.PortId, channel.ChannelId)
		addr, err := sdk.Bech32ifyAddressBytes(OldAddrPrefix+AccAddressPrefix, rawAddr)
		if err != nil {
			return fmt.Errorf("failed to bech32ify address: %w", err)
		}

		// TODO: use this conversion when merging with address replacement branch
		//addr, err := convertAddressToASI(oldAddr, AccAddressPrefix)
		//if err != nil {
		//	return fmt.Errorf("failed to convert address: %w", err)
		//}
		//if _, err = sdk.AccAddressFromBech32(addr); err != nil {
		//	return fmt.Errorf("converted address is invalid: %w", err)
		//}

		balanceIdx, ok := (*balanceMap)[addr]
		if !ok {
			// channel address not found in genesis balances
			continue
		}

		accBalance := balances[balanceIdx]

		// withdraw funds from the channel balance
		balances[withdrawalBalanceIdx].Coins = balances[withdrawalBalanceIdx].Coins.Add(accBalance.Coins...)

		// zero out the channel balance
		balances[balanceIdx].Coins = sdk.NewCoins()
	}

	// update the bank genesis state
	(*bankGenState).Balances = balances

	bankGenStateBytes, err := (*cdc).MarshalJSON(bankGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal auth genesis state: %w", err)
	}

	(*appState)[banktypes.ModuleName] = bankGenStateBytes
	return nil
}

func ASIGenesisUpgradeWithdrawReconciliationBalances() {}

func getGenesisBalancesMap(bankGenState *banktypes.GenesisState) (*map[string]int, error) {
	balanceMap := make(map[string]int)

	for i, balance := range bankGenState.Balances {
		balanceMap[balance.Address] = i
	}

	return &balanceMap, nil
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
