package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/bech32"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	cosmostypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/spf13/cobra"
	"regexp"
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

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			genFile := config.GenesisFile()

			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			appStateStr := string(appStateJSON)
			// replace addresses across the genesis file
			ASIGenesisUpgradeReplaceAddresses(&appStateStr)

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

func ASIGenesisUpgradeReplaceChainID() {}

func ASIGenesisUpgradeReplaceDenom() {}

func ASIGenesisUpgradeReplaceAddresses(jsonString *string) {
	// account addresses
	replaceAddresses(AccAddressPrefix, jsonString, AddrDataLength+AddrChecksumLength)
	// validator addresses
	replaceAddresses(ValAddressPrefix, jsonString, AddrDataLength+AddrChecksumLength)
	// consensus addresses
	replaceAddresses(ConsAddressPrefix, jsonString, AddrDataLength+AddrChecksumLength)
	// contract addresses
	replaceAddresses(AccAddressPrefix, jsonString, WasmDataLength+AddrChecksumLength)
}

func replaceAddresses(addressTypePrefix string, jsonString *string, dataLength int) {
	re := regexp.MustCompile(fmt.Sprintf(`"%s%s1([%s]{%d})"`, OldAddrPrefix, addressTypePrefix, Bech32Chars, dataLength))
	matches := re.FindAllString(*jsonString, -1)

	replacements := make(map[string]string, len(matches))
	for _, match := range matches {
		matchedAddr := strings.ReplaceAll(match, `"`, "")
		_, decodedAddrData, err := bech32.Decode(matchedAddr)
		if err != nil {
			panic(err)
		}

		newAddress, err := bech32.Encode(NewAddrPrefix+addressTypePrefix, decodedAddrData)
		if err != nil {
			panic(err)
		}

		err = cosmostypes.VerifyAddressFormat(decodedAddrData)
		if err != nil {
			panic(err)
		}

		switch addressTypePrefix {
		case AccAddressPrefix:
			_, err = cosmostypes.AccAddressFromBech32(newAddress)
		case ValAddressPrefix:
			_, err = cosmostypes.ValAddressFromBech32(newAddress)
		case ConsAddressPrefix:
			_, err = cosmostypes.ConsAddressFromBech32(newAddress)
		default:
			panic("invalid address type prefix")
		}
		if err != nil {
			panic(err)
		}
		replacements[matchedAddr] = newAddress
	}

	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(*jsonString), &jsonData)
	if err != nil {
		panic(err)
	}

	modified := crawlJson(jsonData, func(data interface{}) interface{} {
		if str, ok := data.(string); ok {
			if !re.MatchString(fmt.Sprintf(`"%s"`, str)) || len(str) > 200 {
				return data
			}

			return replacements[str]
		}
		return data
	})

	modifiedJSON, err := json.Marshal(modified)
	if err != nil {
		panic(err)
	}
	*jsonString = string(modifiedJSON)
}

func ASIGenesisUpgradeWithdrawIBCChannelsBalances() {}

func ASIGenesisUpgradeWithdrawReconciliationBalances() {}

func crawlJson(data interface{}, strHandler func(interface{}) interface{}) interface{} {
	switch value := data.(type) {
	case string:
		if strHandler != nil {
			return strHandler(data)
		}
	case []interface{}:
		for i := range value {
			value[i] = crawlJson(value[i], strHandler)
		}
	case map[string]interface{}:
		for k := range value {
			value[k] = crawlJson(value[k], strHandler)
		}
	default:
	}
	return data
}
