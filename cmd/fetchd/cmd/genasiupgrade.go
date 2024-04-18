package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/types"
	"strings"
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
			cdc := clientCtx.Codec

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

			// set denom metadata in bank module
			err = ASIGenesisUpgradeReplaceDenomMetadata(cdc, &appState)
			if err != nil {
				return fmt.Errorf("failed to replace denom metadata: %w", err)
			}

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			appStateStr := string(appStateJSON)

			// replace denom across the genesis file
			ASIGenesisUpgradeReplaceDenom(&appStateStr)

			// save the updated genesis file
			genDoc.AppState = []byte(appStateStr)
			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")
	cmd.Flags().String(flagNewDescription, "", "The new description for the native coin in the genesis file")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func ASIGenesisUpgradeReplaceDenomMetadata(cdc codec.Codec, appState *map[string]json.RawMessage) error {
	bankGenState := banktypes.GetGenesisStateFromAppState(cdc, *appState)

	OldBaseDenomUpper := strings.ToUpper(OldBaseDenom)
	NewBaseDenomUpper := strings.ToUpper(NewBaseDenom)

	newMetadata := banktypes.Metadata{
		Base: NewDenom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    NewBaseDenomUpper,
				Exponent: 18,
			},
			{
				Denom:    fmt.Sprintf("m%s", NewBaseDenom),
				Exponent: 15,
			},
			{
				Denom:    fmt.Sprintf("u%s", NewBaseDenom),
				Exponent: 12,
			},
			{
				Denom:    fmt.Sprintf("n%s", NewBaseDenom),
				Exponent: 9,
			},
			{
				Denom:    fmt.Sprintf("p%s", NewBaseDenom),
				Exponent: 6,
			},
			{
				Denom:    fmt.Sprintf("f%s", NewBaseDenom),
				Exponent: 3,
			},
			{
				Denom:    fmt.Sprintf("a%s", NewBaseDenom),
				Exponent: 0,
			},
		},
		Description: NewDescription,
		Display:     NewBaseDenomUpper,
		Name:        NewBaseDenomUpper,
		Symbol:      NewBaseDenomUpper,
	}

	for i, metadata := range bankGenState.DenomMetadata {
		if metadata.Name == OldBaseDenomUpper {
			(*bankGenState).DenomMetadata[i] = newMetadata
			break
		}
	}

	bankGenStateBytes, err := cdc.MarshalJSON(bankGenState)
	if err != nil {
		return fmt.Errorf("failed to marshal auth genesis state: %w", err)
	}

	(*appState)[banktypes.ModuleName] = bankGenStateBytes
	return nil
}

func ASIGenesisUpgradeReplaceChainID(genesisData *types.GenesisDoc) {
	genesisData.ChainID = NewChainId
}

func ASIGenesisUpgradeReplaceDenom(jsonString *string) {
	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(*jsonString), &jsonData)
	if err != nil {
		panic(err)
	}

	keyIsInTarget := func(target string) bool {
		for _, key := range []string{"denom", "bond_denom", "mint_denom", "base_denom", "base"} {
			if target == key {
				return true
			}
		}
		return false
	}

	modifiedGenesisJson := crawlJson(nil, jsonData, func(key interface{}, value interface{}) interface{} {
		if str, ok := value.(string); ok {
			if str == OldDenom && keyIsInTarget(key.(string)) {
				return NewDenom
			}
		}
		return value
	})

	modifiedJSON, err := json.Marshal(modifiedGenesisJson)
	if err != nil {
		panic(err)
	}
	*jsonString = string(modifiedJSON)
}

func ASIGenesisUpgradeReplaceAddresses() {}

func ASIGenesisUpgradeWithdrawIBCChannelsBalances() {}

func ASIGenesisUpgradeWithdrawReconciliationBalances() {}

func crawlJson(key interface{}, value interface{}, strHandler func(interface{}, interface{}) interface{}) interface{} {
	switch val := value.(type) {
	case string:
		if strHandler != nil {
			return strHandler(key, val)
		}
	case []interface{}:
		for i := range val {
			val[i] = crawlJson(nil, val[i], strHandler)
		}
	case map[string]interface{}:
		for k := range val {
			val[k] = crawlJson(k, val[k], strHandler)
		}
	default:
	}
	return value
}
