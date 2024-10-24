package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/bech32"
	bech32btc "github.com/cosmos/btcutil/bech32"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/cosmos/cosmos-sdk/types/errors"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/tendermint/tendermint/types"
)

func loadAppStateFromGenesis(genesisPath string) (genDoc *types.GenesisDoc, appState genutiltypes.AppMap, err error) {
	genDoc, err = types.GenesisDocFromFile(genesisPath)
	if err != nil {
		return nil, genutiltypes.AppMap{}, fmt.Errorf("failed to load genesis file at %q: %w", genesisPath, err)
	}
	if err := json.Unmarshal(genDoc.AppState, &appState); err != nil {
		return nil, genutiltypes.AppMap{}, errors.Wrap(err, "failed to JSON unmarshal initial genesis state")
	}
	return genDoc, appState, nil
}

func utilCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "util",
		Aliases:                    []string{"u"},
		Short:                      "Utility subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		utilJsonCommand(),
		utilAddressCommand(),
		utilNetworkMergeCommand(),
	)

	return cmd
}

// ReadJSONFile reads the content of a JSON file and returns it as a []byte.
func ReadJSONFile(filePath string) (string, error) {
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}
	return string(fileBytes), nil
}

func NormaliseJsonFileContent(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("provided file path is empty string")
	}

	var err error
	var jsonStr string
	if jsonStr, err = ReadJSONFile(path); err != nil {
		return "", err
	}

	var jsonNormalised string
	if jsonNormalised, err = NormalizeJSONString(jsonStr); err != nil {
		return "", err
	}

	return jsonNormalised, nil
}

func NormaliseJsonFileContentSha256Hex(path string) (string, error) {
	var err error
	var jsonNormalised string
	if jsonNormalised, err = NormaliseJsonFileContent(path); err != nil {
		return "", err
	}

	hash := sha256.Sum256([]byte(jsonNormalised))
	hashHex := hex.EncodeToString(hash[:])
	return hashHex, nil
}

func utilJsonCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "json",
		Aliases:                    []string{"a"},
		Short:                      "json subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmdHash := &cobra.Command{
		Use:   "normalised-hash [json_file_path]",
		Short: "Generates sha256 hash from *normalised* content of json file",
		Long:  "Command normalises json before generating the sha256 out of it (see description of the `normalise` sub-command)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			var err error
			var hashHex string
			if hashHex, err = NormaliseJsonFileContentSha256Hex(path); err != nil {
				return err
			}

			ctx := client.GetClientContextFromCmd(cmd)
			return ctx.PrintString(fmt.Sprintf("%s\n", hashHex))
		},
	}

	cmdNormalise := &cobra.Command{
		Use:   "normalise [json_file_path]",
		Short: "Normalises content of json file and prints it to stdout",
		Long:  "Normalisation means sorting all dictionaries in the json structure by their keys, and then removing all non-active white space characters (= characters which do *not* participate on value of the json =  they can be removed or added and value of json will remain the same)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := args[0]

			var err error
			var jsonNormalised string
			if jsonNormalised, err = NormaliseJsonFileContent(path); err != nil {
				return err
			}

			ctx := client.GetClientContextFromCmd(cmd)
			return ctx.PrintString(fmt.Sprintf("%s\n", jsonNormalised))
		},
	}

	cmd.AddCommand(
		cmdNormalise,
		cmdHash,
	)

	return cmd
}

func bech32BytesToBech32Str(beech32Bytes []byte) (string, error) {
	var bldr strings.Builder
	// Write the data part, using the bech32 charset.
	const bech32Charset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
	for _, b := range beech32Bytes {
		if int(b) >= len(bech32Charset) {
			return "", bech32.ErrInvalidDataByte(b)
		}
		bldr.WriteByte(bech32Charset[b])
	}
	return bldr.String(), nil
}

func utilAddressCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "address",
		Aliases:                    []string{"a"},
		Short:                      "Address subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmdParseAddress := &cobra.Command{
		Use:   "parse [bech32_address]",
		Short: "Parses bech32 human readable address and prints its components - prefix, raw address and checksum",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			bech32Address := args[0]

			if len(bech32Address) == 0 {
				return fmt.Errorf("empty address string")
			}

			extractedPrefix, rawAddress5bits, checksum5bits, err := bech32btc.DecodeUnsafe(bech32Address)
			if err != nil {
				return err
			}

			var rawAddress []byte
			rawAddress, err = bech32.ConvertBits(rawAddress5bits, 5, 8, false)
			if err != nil {
				return fmt.Errorf("decoding bech32 failed: %w", err)
			}

			var checksum []byte
			checksum, err = bech32.ConvertBits(checksum5bits, 5, 8, true)
			if err != nil {
				return fmt.Errorf("decoding checksum failed: %w", err)
			}

			var checksum5bitsVerif []byte
			checksum5bitsVerif, err = bech32.ConvertBits(checksum, 8, 5, false)
			if err != nil {
				return fmt.Errorf("encoding checksum failed: %w", err)
			}

			var checksumBech32Verif string
			checksumBech32Verif, err = bech32BytesToBech32Str(checksum5bitsVerif)
			if err != nil {
				return fmt.Errorf("encoding checksum failed: %w", err)
			}

			var rawAddressBech32Verif string
			rawAddressBech32Verif, err = bech32BytesToBech32Str(rawAddress5bits)
			if err != nil {
				return fmt.Errorf("encoding checksum failed: %w", err)
			}

			rawAddressHex := hex.EncodeToString(rawAddress)
			checksumHex := hex.EncodeToString(checksum)

			ctx := client.GetClientContextFromCmd(cmd)
			return ctx.PrintString(fmt.Sprintf("prefix: %v\nraw address [bech32]: %v\nraw address [hex]: %v\nchecksum [bech32]: %v\nchecksum [hex]: %v\n", extractedPrefix, rawAddressBech32Verif, rawAddressHex, checksumBech32Verif, checksumHex))
		},
	}

	cmdChangePrefix := &cobra.Command{
		Use:   "change-prefix [bech32_address] [prefix]",
		Short: "Changes prefix of the address",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			bech32Address := args[0]
			newPrefix := args[1]

			if len(bech32Address) == 0 {
				return fmt.Errorf("empty address string")
			}

			_ /*extractedPrefix*/, rawAddress5bits, _ /*checksum*/, err := bech32btc.DecodeUnsafe(bech32Address)
			//extractedPrefix, rawAddress, err := bech32.DecodeAndConvert(bech32Address)
			if err != nil {
				return err
			}

			//var rawAddress []byte
			//rawAddress, err = bech32.ConvertBits(rawAddress5bits, 5, 8, false)
			//if err != nil {
			//	return fmt.Errorf("decoding bech32 failed: %w", err)
			//}

			var newAddress string
			newAddress, err = bech32btc.Encode(newPrefix, rawAddress5bits)
			if err != nil {
				return err
			}

			ctx := client.GetClientContextFromCmd(cmd)
			return ctx.PrintString(fmt.Sprintf("%s\n", newAddress))
		},
	}

	cmd.AddCommand(
		cmdParseAddress,
		cmdChangePrefix,
	)
	return cmd
}

// NormalizeJSON recursively normalizes the JSON by sorting keys in maps and removing whitespace.
func NormalizeJSON(data interface{}) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		// Sort the map keys and normalize the values recursively
		sortedMap := make(map[string]interface{})
		keys := make([]string, 0, len(v))
		for key := range v {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			sortedMap[key] = NormalizeJSON(v[key])
		}
		return sortedMap
	case []interface{}:
		// Normalize each item in the list recursively
		for i := range v {
			v[i] = NormalizeJSON(v[i])
		}
	}
	return data
}

// NormalizeJSONString takes a JSON string, normalizes it, and returns the result as a minified string.
func NormalizeJSONString(input string) (string, error) {
	// Unmarshal the input JSON string into an interface{}
	var data interface{}
	err := json.Unmarshal([]byte(input), &data)
	if err != nil {
		return "", fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	// Recursively normalize the JSON
	normalizedData := NormalizeJSON(data)

	// Marshal the normalized data into minified JSON format (no extra whitespace)
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "") // no indentation for minified output
	err = encoder.Encode(normalizedData)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %v", err)
	}

	// Return the normalized, minified JSON string
	return buf.String(), nil
}
