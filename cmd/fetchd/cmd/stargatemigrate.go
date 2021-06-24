package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/CosmWasm/wasmd/x/wasm"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	v038auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v038"
	v039 "github.com/cosmos/cosmos-sdk/x/genutil/legacy/v039"
	v040 "github.com/cosmos/cosmos-sdk/x/genutil/legacy/v040"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"
)

const flagGenesisTime = "genesis-time"
const flagConsensusEvidenceMaxBytes = "consensus-evidence-max-bytes"

// AddStargateMigrateCmd returns a command to migrate genesis to stargate version.
func AddStargateMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stargate-migrate [genesis-file]",
		Short: "Migrate fetch.ai mainnet genesis to the stargate Cosmos SDK version",
		Long: `Set the consensus.params.evidence.max_bytes value to 150Kb,
removes multisig public keys that fail to decode, reset wasm module to its genesis state (as existing contracts are not backward compatible),
and then migrate the given genesis to version v0.39, and then v0.40 of the cosmos-sdk.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			importGenesis := args[0]
			genDoc, err := tmtypes.GenesisDocFromFile(importGenesis)
			if err != nil {
				return fmt.Errorf("failed to load genesis file at %q: %w", importGenesis, err)
			}

			// Set consensus_params.evidence.max_bytes to avoid warning from 0 value
			maxBytes, err := cmd.Flags().GetInt64(flagConsensusEvidenceMaxBytes)
			if err != nil {
				return fmt.Errorf("failed to retrieve flag %q: %w", flagConsensusEvidenceMaxBytes, err)
			}
			genDoc.ConsensusParams.Evidence.MaxBytes = maxBytes

			var v038GenState types.AppMap
			if err := json.Unmarshal(genDoc.AppState, &v038GenState); err != nil {
				return errors.Wrap(err, "failed to JSON unmarshal initial genesis state")
			}

			v038Codec := codec.NewLegacyAmino()
			v038auth.RegisterLegacyAminoCodec(v038Codec)

			// Drop any multisig.LegacyAminoPubKey we have as v039 migration crash on them.
			var authGenState v038auth.GenesisState
			v038Codec.MustUnmarshalJSON(v038GenState[v038auth.ModuleName], &authGenState)
			for i, acc := range authGenState.Accounts {
				switch t := acc.(type) {
				case *v038auth.BaseAccount:
					switch t.PubKey.(type) {
					case *multisig.LegacyAminoPubKey:
						t.PubKey = nil
						authGenState.Accounts[i] = t
					default:
						continue
					}
				default:
					continue
				}
			}
			v038GenState[v038auth.ModuleName] = v038Codec.MustMarshalJSON(authGenState)

			// v039 migration
			v039GenState := v039.Migrate(v038GenState, clientCtx)
			// sanity check that the state can still be marhsalled to json
			_, err = json.Marshal(v039GenState)
			if err != nil {
				return errors.Wrap(err, "failed to JSON marshal migrated genesis state")
			}

			// v040 migration
			v040GenState := v040.Migrate(v039GenState, clientCtx)

			// Reset wasm module to genesis
			v040WasmDefaultState, err := json.Marshal(&wasm.GenesisState{
				Params: wasm.DefaultParams(),
			})
			if err != nil {
				return errors.Wrap(err, "failed to marshal wasm default genesis state")
			}
			v040GenState[wasm.ModuleName] = v040WasmDefaultState

			// Update genesis with migrated state
			genDoc.AppState, err = json.Marshal(v040GenState)
			if err != nil {
				return errors.Wrap(err, "failed to JSON marshal migrated genesis state")
			}

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

			chainID, err := cmd.Flags().GetString(flags.FlagChainID)
			if err != nil {
				return fmt.Errorf("failed to read %q flag: %w", flags.FlagChainID, err)
			}
			if chainID != "" {
				genDoc.ChainID = chainID
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
	cmd.Flags().Int64(flagConsensusEvidenceMaxBytes, 150000, "override consensus.evidence.max_bytes with this flag")
	cmd.Flags().String(flags.FlagChainID, "", "override chain_id with this flag")

	return cmd
}
