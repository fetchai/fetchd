package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/bech32"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/types"
	"regexp"
	"strings"
)

const (
	BridgeContractAddress  = "fetch1qxxlalvsdjd07p07y3rc5fu6ll8k4tmetpha8n"
	NewBridgeContractAdmin = "fetch15p3rl5aavw9rtu86tna5lgxfkz67zzr6ed4yhw"

	flagNewDescription = "new-description"
	Bech32Chars        = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
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

			var jsonData map[string]interface{}
			if err = json.Unmarshal(genDoc.AppState, &jsonData); err != nil {
				return fmt.Errorf("failed to unmarshal app state: %w", err)
			}

			// replace bridge contract admin
			ASIGenesisUpgradeReplaceBridgeAdmin(jsonData)

			// replace addresses across the genesis file
			ASIGenesisUpgradeReplaceAddresses(jsonData)

			// set denom metadata in bank module
			err = ASIGenesisUpgradeReplaceDenomMetadata(jsonData)
			if err != nil {
				return fmt.Errorf("failed to replace denom metadata: %w", err)
			}

			// replace denom across the genesis file
			ASIGenesisUpgradeReplaceDenom(jsonData)

			// replace chain-id
			ASIGenesisUpgradeReplaceChainID(genDoc)

			var encodedAppState []byte
			if encodedAppState, err = json.Marshal(jsonData); err != nil {
				return err
			}

			genDoc.AppState = encodedAppState
			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")
	cmd.Flags().String(flagNewDescription, "", "The new description for the native coin in the genesis file")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func ASIGenesisUpgradeReplaceDenomMetadata(jsonData map[string]interface{}) error {
	type jsonMap map[string]interface{}

	NewBaseDenomUpper := strings.ToUpper(NewBaseDenom)

	newMetadata := jsonMap{
		"base": NewDenom,
		"denom_units": []jsonMap{
			{
				"denom":    NewBaseDenomUpper,
				"exponent": 18,
			},
			{
				"denom":    fmt.Sprintf("m%s", NewBaseDenom),
				"exponent": 15,
			},
			{
				"denom":    fmt.Sprintf("u%s", NewBaseDenom),
				"exponent": 12,
			},
			{
				"denom":    fmt.Sprintf("n%s", NewBaseDenom),
				"exponent": 9,
			},
			{
				"denom":    fmt.Sprintf("p%s", NewBaseDenom),
				"exponent": 6,
			},
			{
				"denom":    fmt.Sprintf("f%s", NewBaseDenom),
				"exponent": 3,
			},
			{
				"denom":    fmt.Sprintf("a%s", NewBaseDenom),
				"exponent": 0,
			},
		},
		"description": NewDescription,
		"display":     NewBaseDenomUpper,
		"name":        NewBaseDenomUpper,
		"symbol":      NewBaseDenomUpper,
	}

	bank := jsonData["bank"].(map[string]interface{})
	denomMetadata := bank["denom_metadata"].([]interface{})

	for i, metadata := range denomMetadata {
		denomUnit := metadata.(map[string]interface{})
		if denomUnit["base"] == OldDenom {
			denomMetadata[i] = newMetadata
			break
		}
	}
	return nil
}

func ASIGenesisUpgradeReplaceChainID(genesisData *types.GenesisDoc) {
	genesisData.ChainID = NewChainId
}

func ASIGenesisUpgradeReplaceBridgeAdmin(jsonData map[string]interface{}) {
	contracts := jsonData["wasm"].(map[string]interface{})["contracts"].([]interface{})

	for i, contract := range contracts {
		c := contract.(map[string]interface{})
		if c["contract_address"] == BridgeContractAddress {
			contractInfo := c["contract_info"].(map[string]interface{})
			contractInfo["admin"] = NewBridgeContractAdmin
			contracts[i] = c
			break
		}
	}
}

func ASIGenesisUpgradeReplaceDenom(jsonData map[string]interface{}) {
	targets := map[string]struct{}{"denom": {}, "bond_denom": {}, "mint_denom": {}, "base_denom": {}, "base": {}}

	crawlJson("", jsonData, -1, func(key string, value interface{}, idx int) interface{} {
		if str, ok := value.(string); ok {
			_, isInTargets := targets[key]
			if str == OldDenom && isInTargets {
				return NewDenom
			}
		}
		return value
	})
}

func ASIGenesisUpgradeReplaceAddresses(jsonData map[string]interface{}) {
	// account addresses
	replaceAddresses(AccAddressPrefix, jsonData, AddrDataLength+AddrChecksumLength)
	// validator addresses
	replaceAddresses(ValAddressPrefix, jsonData, AddrDataLength+AddrChecksumLength)
	// consensus addresses
	replaceAddresses(ConsAddressPrefix, jsonData, AddrDataLength+AddrChecksumLength)
	// contract addresses
	replaceAddresses(AccAddressPrefix, jsonData, WasmDataLength+AddrChecksumLength)
}

func replaceAddresses(addressTypePrefix string, jsonData map[string]interface{}, dataLength int) {
	re := regexp.MustCompile(fmt.Sprintf(`^%s%s1([%s]{%d})$`, OldAddrPrefix, addressTypePrefix, Bech32Chars, dataLength))

	crawlJson("", jsonData, -1, func(key string, value interface{}, idx int) interface{} {
		if str, ok := value.(string); ok {
			if !re.MatchString(str) {
				return value
			}
			newAddress, err := convertAddressToASI(str, addressTypePrefix)
			if err != nil {
				panic(err)
			}

			return newAddress
		}
		return value
	})
}

func ASIGenesisUpgradeWithdrawIBCChannelsBalances() {}

func ASIGenesisUpgradeWithdrawReconciliationBalances() {}

func crawlJson(key string, value interface{}, idx int, strHandler func(string, interface{}, int) interface{}) interface{} {
	switch val := value.(type) {
	case string:
		if strHandler != nil {
			return strHandler(key, val, idx)
		}
	case []interface{}:
		for i := range val {
			val[i] = crawlJson("", val[i], i, strHandler)
		}
	case map[string]interface{}:
		for k, v := range val {
			val[k] = crawlJson(k, v, -1, strHandler)
		}
	}
	return value
}

func convertAddressToASI(addr string, addressTypePrefix string) (string, error) {
	_, decodedAddrData, err := bech32.Decode(addr)
	if err != nil {
		return "", fmt.Errorf("failed to decode address: %w", err)
	}

	err = sdk.VerifyAddressFormat(decodedAddrData)
	if err != nil {
		return "", fmt.Errorf("failed to verify address format: %w", err)
	}

	newAddress, err := bech32.Encode(NewAddrPrefix+addressTypePrefix, decodedAddrData)
	if err != nil {
		return "", fmt.Errorf("failed to encode new address: %w", err)
	}

	return newAddress, nil
}
