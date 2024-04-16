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
	"regexp"
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

			genDoc.AppState = []byte(appStateStr)
      
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

func ASIGenesisUpgradeReplaceDenomMetadata(cdc codec.Codec, appState *map[string]json.RawMessage) error {
	bankGenState := banktypes.GetGenesisStateFromAppState(cdc, *appState)

	OldBaseDenomUpper := strings.ToUpper(OldBaseDenom)
	NewBaseDenomUpper := strings.ToUpper(NewBaseDenom)

	denomRegex := getRegex(OldDenom, NewDenom)
	upperDenomRegex := getRegex(OldBaseDenomUpper, NewBaseDenomUpper)
	exponentDenomRegex := getPartialRegexLeft(OldBaseDenom, NewBaseDenom)

	for _, metadata := range bankGenState.DenomMetadata {
		replaceString(&metadata.Base, []*regexPair{denomRegex})
		if metadata.Name == OldBaseDenomUpper {
			metadata.Description = NewDescription
			metadata.Display = NewBaseDenomUpper
			metadata.Name = NewBaseDenomUpper
			metadata.Symbol = NewBaseDenomUpper
		}
		for _, unit := range metadata.DenomUnits {
			replaceString(&unit.Denom, []*regexPair{upperDenomRegex})
			replaceString(&unit.Denom, []*regexPair{exponentDenomRegex})
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
	for _, target := range []string{"denom", "bond_denom", "mint_denom", "base_denom", "base"} {
		re := regexp.MustCompile(fmt.Sprintf(`("%s"\s*:\s*)"%s"`, target, OldDenom))
		if re.MatchString(*jsonString) {
			*jsonString = re.ReplaceAllString(*jsonString, fmt.Sprintf(`${1}"%s"`, NewDenom))
		}
	}
}

func ASIGenesisUpgradeReplaceAddresses() {}

func ASIGenesisUpgradeWithdrawIBCChannelsBalances() {}

func ASIGenesisUpgradeWithdrawReconciliationBalances() {}

func getRegex(oldValue string, newValue string) *regexPair {
	return &regexPair{
		pattern:     fmt.Sprintf(`^%s$`, oldValue),
		replacement: fmt.Sprintf(`%s`, newValue),
	}
}

func getPartialRegexLeft(oldValue string, newValue string) *regexPair {
	return &regexPair{
		pattern:     fmt.Sprintf(`(.*?)%s"`, oldValue),
		replacement: fmt.Sprintf(`${1}%s"`, newValue),
	}
}

func replaceString(s *string, replacements []*regexPair) {
	for _, pair := range replacements {
		re := regexp.MustCompile(pair.pattern)
		if re.MatchString(*s) {
			*s = re.ReplaceAllString(*s, pair.replacement)
		}
	}
}

type regexPair struct {
	pattern     string
	replacement string
}
