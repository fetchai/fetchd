package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	authz "github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	legacyibc "github.com/cosmos/ibc-go/v2/modules/core/legacy/v100"
	"github.com/fetchai/fetchd/app"
)

const (
	flagGenesisTime                = "genesis-time"
	flagInitialHeight              = "initial-height"
	flagIBCMaxExpectedTimePerBlock = "max-expected-time-per-block"
)

// AddDoradoMigrateCmd returns a command to migrate genesis to Dorado version.
func AddDoradoMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dorado-migrate <genesis-file>",
		Short: "Migrate fetchAI mainnet genesis from the Capricorn version to the Dorado version",
		Long:  ``,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			importGenesis := args[0]

			genDoc, err := tmtypes.GenesisDocFromFile(importGenesis)
			if err != nil {
				return fmt.Errorf("failed to load genesis file at %q: %w", importGenesis, err)
			}

			// set genesis time
			genesisTime, err := cmd.Flags().GetString(flagGenesisTime)
			if err != nil {
				return fmt.Errorf("failed to read %q flag: %w", flagGenesisTime, err)
			}
			if genesisTime != "" {
				var t time.Time

				err := t.UnmarshalText([]byte(genesisTime))
				if err != nil {
					return errors.Wrap(err, "failed to unmarshal genesis time")
				}

				genDoc.GenesisTime = t
			}

			// set initial height
			initialHeight, err := cmd.Flags().GetInt64(flagInitialHeight)
			if err != nil {
				return fmt.Errorf("failed to retrieve flag %q: %w", flagInitialHeight, err)
			}
			// only set initial height if it was given
			// otherwise, keep the initial_height from exported genesis
			// that should already be set to last committed block+1
			if initialHeight > 0 {
				genDoc.InitialHeight = initialHeight
			}

			// set new chain-id
			chainID, err := cmd.Flags().GetString(flags.FlagChainID)
			if err != nil {
				return fmt.Errorf("failed to read %q flag: %w", flags.FlagChainID, err)
			}
			if chainID != "" {
				genDoc.ChainID = chainID
			}

			var appState types.AppMap
			if err := json.Unmarshal(genDoc.AppState, &appState); err != nil {
				return errors.Wrap(err, "failed to JSON unmarshal initial genesis state")
			}

			// ------------ Start of custom migration operations ------------

			// Migrate ibc state from v1 to v2
			maxExpectedTimePerBlock, err := cmd.Flags().GetUint64(flagIBCMaxExpectedTimePerBlock)
			if err != nil {
				return fmt.Errorf("failed to read %q flag: %w", flagIBCMaxExpectedTimePerBlock, err)
			}
			appState, err = legacyibc.MigrateGenesis(appState, clientCtx, *genDoc, maxExpectedTimePerBlock)
			if err != nil {
				return errors.Wrap(err, "failed to IBC migrate genesis")
			}

			// Init Authz module state
			authzbz, err := clientCtx.Codec.MarshalJSON(authz.DefaultGenesisState())
			if err != nil {
				return errors.Wrap(err, "failed to marshal authz state")
			}
			appState[authz.ModuleName] = authzbz

			// ------------ End of custom migration operations ------------

			// Validate state (same as fetchd validate-genesis cmd)
			if err := app.ModuleBasics.ValidateGenesis(clientCtx.Codec, clientCtx.TxConfig, appState); err != nil {
				return fmt.Errorf("failed to validate state: %w", err)
			}

			// build and print the new genesis state
			genDoc.AppState, err = json.Marshal(appState)
			if err != nil {
				return errors.Wrap(err, "failed to JSON marshal migrated genesis state")
			}
			bz, err := tmjson.Marshal(genDoc)
			if err != nil {
				return errors.Wrap(err, "failed to marshal genesis doc")
			}
			sortedBz, err := sdk.SortJSON(bz)
			if err != nil {
				return errors.Wrap(err, "failed to sort JSON genesis doc")
			}

			fmt.Println(string(sortedBz))

			return nil
		},
	}

	cmd.Flags().String(flagGenesisTime, "", "override genesis_time with this flag")
	cmd.Flags().Int64(flagInitialHeight, 0, "override initial_height with this flag")
	cmd.Flags().String(flags.FlagChainID, "fetchhub-4", "override chain_id with this flag")
	// see https://github.com/cosmos/ibc-go/blob/v2.0.3/modules/core/03-connection/types/connection.pb.go#L359-L362
	cmd.Flags().Uint64(flagIBCMaxExpectedTimePerBlock, 30000000000, "value for ibc.connection_genesis.params.max_expected_time_per_block")

	return cmd
}
