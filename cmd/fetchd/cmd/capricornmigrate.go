package cmd

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	ibctransfer "github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
	ibchost "github.com/cosmos/cosmos-sdk/x/ibc/core/24-host"
	ibccoretypes "github.com/cosmos/cosmos-sdk/x/ibc/core/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/fetchai/fetchd/app"
)

const (
	flagGenesisTime                    = "genesis-time"
	flagInitialHeight                  = "initial-height"
	flagConsensusBlockMaxBytes         = "consensus-block-max-bytes"
	flagConsensusBlockMaxGas           = "consensus-block-max-gas"
	flagFoundationAddress              = "foundation-address"
	flagFoundationTokensToBurn         = "foundation-tokens-to-burn"
	flagStakingParamsHistoricalEntries = "staking-historical-entries"
	flagBridgeNewContractPath          = "bridge-new-contract-path"
	flagMobixNewContractPath           = "mobix-new-contract-path"
	flagNewSlashFractionDoubleSign     = "new-slash-fraction-double-sign"
)

const (
	bridgeContractCodeID = 2
	mobixContractCodeID  = 3
)

var (
	// contractCodeIDsToRemove list the codeIDs of the contracts to be removed by the migration
	contractCodeIDsToRemove = []string{
		"1", // Unused bridge contract (effective one use codeID 2)
		"4", // Reconciliation contract v0 (not needed anymore)
		"5", // Reconciliation contract v1 (not needed anymore)
	}
)

// AddCapricornMigrateCmd returns a command to migrate genesis to capricorn version.
func AddCapricornMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "capricorn-migrate <genesis-file> --bridge-new-contract-path /path/to/bridge.wasm --mobix-new-contract-path /path/to/mobix.wasm",
		Short: "Migrate fetchAI mainnet genesis from the Stargate version to the Capricorn version",
		Long: `Migrate fetchAI mainnet genesis from the Stargate version to the Capricorn version.
It does the following operations:
	- set new chain-id and genesis-time
	- increase consensus block max_bytes & max_gas & evidence max_bytes
	- burn some foundation tokens (ERC20 stake migration cleanup)
	- enable IBC transfers and set staking historical entries parameter (required by IBC module).
	- delete unused contract codes and states
	- update bridge and mobix contracts to cosmwasm v1.0.0
	- adjust mobix config unbonding_period from nanoseconds to seconds
	- increase slashing.params.slash_fraction_double_sign value
`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			depCdc := clientCtx.JSONMarshaler
			cdc := depCdc.(codec.Marshaler)

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

			// increase consensus block max_bytes & max_gas
			maxBytes, err := cmd.Flags().GetInt64(flagConsensusBlockMaxBytes)
			if err != nil {
				return fmt.Errorf("failed to retrieve flag %q: %w", flagConsensusBlockMaxBytes, err)
			}
			maxGas, err := cmd.Flags().GetInt64(flagConsensusBlockMaxGas)
			if err != nil {
				return fmt.Errorf("failed to retrieve flag %q: %w", flagConsensusBlockMaxGas, err)
			}

			genDoc.ConsensusParams.Block.MaxBytes = maxBytes
			genDoc.ConsensusParams.Block.MaxGas = maxGas
			genDoc.ConsensusParams.Evidence.MaxBytes = maxBytes

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

			// Burn foundation tokens
			foundationAddressStr, err := cmd.Flags().GetString(flagFoundationAddress)
			if err != nil {
				return fmt.Errorf("failed to retrieve flag %q: %w", flagFoundationAddress, err)
			}
			foundationAddr, err := sdk.AccAddressFromBech32(foundationAddressStr)
			if err != nil {
				return fmt.Errorf("failed to parse bech32 foundation address: %w", err)
			}

			foundationTokensToBurnStr, err := cmd.Flags().GetString(flagFoundationTokensToBurn)
			if err != nil {
				return fmt.Errorf("failed to retrieve flag %q: %w", flagFoundationTokensToBurn, err)
			}
			foundationCoinsToBurn, err := sdk.ParseCoinsNormalized(foundationTokensToBurnStr)
			if err != nil {
				return fmt.Errorf("failed to parse coins to burn: %w", err)
			}

			appState, err = burnTokens(appState, cdc, foundationAddr, foundationCoinsToBurn)
			if err != nil {
				return fmt.Errorf("failed to burn tokens: %w", err)
			}

			// Enable IBC
			numHistoricalEntries, err := cmd.Flags().GetUint32(flagStakingParamsHistoricalEntries)
			if err != nil {
				return fmt.Errorf("failed to retrieve flag %q: %w", flagStakingParamsHistoricalEntries, err)
			}

			appState, err = enableIBC(appState, cdc, numHistoricalEntries)
			if err != nil {
				return fmt.Errorf("failed to enable IBC: %w", err)
			}

			// removes unused bridge contract
			appState, err = deleteWrongBridgeContract(appState, cdc)
			if err != nil {
				return fmt.Errorf("failed to delete wrong bridge contract: %w", err)
			}

			// update bridge and mobix contract codes
			newBridgeContractPath, err := cmd.Flags().GetString(flagBridgeNewContractPath)
			if err != nil {
				return fmt.Errorf("failed to retrieve flag %q: %w", flagBridgeNewContractPath, err)
			}
			newMobixContractPath, err := cmd.Flags().GetString(flagMobixNewContractPath)
			if err != nil {
				return fmt.Errorf("failed to retrieve flag %q: %w", flagMobixNewContractPath, err)
			}

			appState, err = migrateWasmContracts(appState, cdc, newBridgeContractPath, newMobixContractPath)
			if err != nil {
				return fmt.Errorf("failed to migrate wasm contracts: %w", err)
			}

			appState, err = updateMobixContractState(appState, cdc)
			if err != nil {
				return fmt.Errorf("failed to migrate wasm contracts: %w", err)
			}

			slashingParamDoubleSignFractionStr, err := cmd.Flags().GetString(flagNewSlashFractionDoubleSign)
			if err != nil {
				return fmt.Errorf("failed to retrieve flag %q: %w", flagNewSlashFractionDoubleSign, err)
			}

			slashingParamDoubleSignFraction, err := sdk.NewDecFromStr(slashingParamDoubleSignFractionStr)
			if err != nil {
				return fmt.Errorf("invalid flag %q value: %w", flagNewSlashFractionDoubleSign, err)
			}

			appState, err = setNewSlashFractionDoubleSign(appState, cdc, slashingParamDoubleSignFraction)
			if err != nil {
				return fmt.Errorf("failed to set new slashing double sign fraction param: %w", err)
			}

			// Validate state (same as fetchd validate-genesis cmd)
			if err := app.ModuleBasics.ValidateGenesis(cdc, clientCtx.TxConfig, appState); err != nil {
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
	cmd.Flags().String(flags.FlagChainID, "fetchhub-3", "override chain_id with this flag")
	cmd.Flags().Int64(flagConsensusBlockMaxBytes, 200_000, "override consensus.block.max_bytes with this flag")
	cmd.Flags().Int64(flagConsensusBlockMaxGas, 2_000_000, "override consensus.block.max_gas with this flag")
	cmd.Flags().String(flagFoundationAddress, "fetch1c2wlfqn6eqqknpwcr0na43m9k6hux94dp6fx4y", "fetch.ai foundation address")
	cmd.Flags().String(flagFoundationTokensToBurn, "80616475991676767023462315afet", "fetch.ai foundation tokens to burn")
	cmd.Flags().Uint32(flagStakingParamsHistoricalEntries, 10000, "override staking.params.historical_entries with this flag")
	cmd.Flags().String(flagBridgeNewContractPath, "", "path to cosmwasm 1.0.0 bridge.wasm contract file")
	cmd.Flags().String(flagMobixNewContractPath, "", "path to cosmwasm 1.0.0 mobix.wasm contract file")
	cmd.Flags().String(flagNewSlashFractionDoubleSign, "0.1", "new slashing fraction params for double sign")

	return cmd
}

// burnTokens removes given coins from given address and decrease total supply accordingly.
func burnTokens(state types.AppMap, cdc codec.JSONMarshaler, addr sdk.AccAddress, coins sdk.Coins) (types.AppMap, error) {
	bankState := banktypes.GetGenesisStateFromAppState(cdc, state)

	var updated bool
	for i, balance := range bankState.Balances {
		if balance.GetAddress().Equals(addr) {
			// sanity check
			if balance.GetCoins().IsAllLT(coins) {
				return nil, fmt.Errorf(
					"insufficient balance from %s, got %s, want at least %s",
					addr.String(),
					balance.GetCoins().String(),
					coins.String(),
				)
			}

			bankState.Balances[i].Coins = bankState.Balances[i].Coins.Sub(coins)
			bankState.Supply = bankState.Supply.Sub(coins)

			updated = true
			break
		}
	}

	if !updated {
		return nil, fmt.Errorf("account %s not found", addr.String())
	}

	bankStateBz, err := cdc.MarshalJSON(bankState)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal bank genesis state: %w", err)
	}
	state[banktypes.ModuleName] = bankStateBz

	return state, nil
}

// enableIBC enable IBC transfer send & receive and set staking historical entries parameter (required by IBC module).
func enableIBC(appState types.AppMap, cdc codec.JSONMarshaler, numHistoricalEntries uint32) (types.AppMap, error) {
	// Enable transfer send & receive
	appState[ibctransfer.ModuleName] = cdc.MustMarshalJSON(ibctransfer.NewGenesisState(
		ibctransfer.PortID,
		ibctransfer.Traces{},
		ibctransfer.NewParams(true, true),
	))
	appState[ibchost.ModuleName] = cdc.MustMarshalJSON(ibccoretypes.DefaultGenesisState())

	// Increase staking params historical entries (required by IBC module)
	stakingState := stakingtypes.GetGenesisStateFromAppState(cdc, appState)
	stakingState.Params.HistoricalEntries = numHistoricalEntries

	stakingStateBz, err := cdc.MarshalJSON(stakingState)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal staking genesis state: %w", err)
	}
	appState[stakingtypes.ModuleName] = stakingStateBz

	return appState, nil
}

// deleteWrongBridgeContract removes unused and duplicate token bridge contract.
func deleteWrongBridgeContract(appState types.AppMap, cdc codec.JSONMarshaler) (types.AppMap, error) {
	// Unset wasm.codes[].code_info source and builder fields from wasm state (from https://github.com/CosmWasm/wasmd/pull/564)
	// we work with raw json as the updated wasm GenesisState fail to unmarshal because of these removed fields.
	var s map[string]interface{}
	if err := json.Unmarshal(appState[wasmtypes.ModuleName], &s); err != nil {
		panic(err)
	}

	// filter codes
	codes := s["codes"].([]interface{})
	var keptCodes []interface{} // nolint: prealloc
	for _, c := range codes {
		code := c.(map[string]interface{})
		// skip over the CodeID we want to remove
		if isCodeIDToRemove(code["code_id"]) {
			continue
		}

		codeInfo := code["code_info"].(map[string]interface{})
		delete(codeInfo, "builder")
		delete(codeInfo, "source")
		keptCodes = append(keptCodes, code)
	}
	s["codes"] = keptCodes

	// filter instanciated contracts
	contracts := s["contracts"].([]interface{})
	var keptContracts []interface{} // nolint: prealloc
	for _, c := range contracts {
		contract := c.(map[string]interface{})
		contractInfos := contract["contract_info"].(map[string]interface{})
		if isCodeIDToRemove(contractInfos["code_id"]) {
			continue
		}

		keptContracts = append(keptContracts, contract)
	}
	s["contracts"] = keptContracts

	statebz, err := json.Marshal(s)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal wasm json state: %w", err)
	}
	appState[wasmtypes.ModuleName] = statebz

	return appState, nil
}

func isCodeIDToRemove(codeID interface{}) bool {
	for _, id := range contractCodeIDsToRemove {
		if id == codeID {
			return true
		}
	}

	return false
}

// migrateWasmContracts replace contract codes and hash with their new and updated versions.
func migrateWasmContracts(appState types.AppMap, cdc codec.JSONMarshaler, newBridgeContractPath, newMobixContractPath string) (types.AppMap, error) {
	var wasmState wasmtypes.GenesisState
	if err := cdc.UnmarshalJSON(appState[wasmtypes.ModuleName], &wasmState); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wasm genesis state: %w", err)
	}

	newBridgeContractBytes, err := ioutil.ReadFile(newBridgeContractPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read contract file %q: %w", newBridgeContractPath, err)
	}
	newBridgeContractHash := sha256.Sum256(newBridgeContractBytes)

	newMobixContractBytes, err := ioutil.ReadFile(newMobixContractPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read contract file %q: %w", newMobixContractPath, err)
	}
	newMobixContractHash := sha256.Sum256(newMobixContractBytes)

	var updated []uint64
	for i, code := range wasmState.Codes {
		switch code.CodeID {
		case bridgeContractCodeID:
			wasmState.Codes[i].CodeBytes = newBridgeContractBytes
			wasmState.Codes[i].CodeInfo.CodeHash = newBridgeContractHash[:]
			updated = append(updated, code.CodeID)
		case mobixContractCodeID:
			wasmState.Codes[i].CodeBytes = newMobixContractBytes
			wasmState.Codes[i].CodeInfo.CodeHash = newMobixContractHash[:]
			updated = append(updated, code.CodeID)
		}
	}

	if got, want := len(updated), 2; got != want {
		return nil, fmt.Errorf("failed to update contract code, got %d updates, want %d: %v", got, want, updated)
	}

	wasmStateBz, err := cdc.MarshalJSON(&wasmState)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal wasm genesis state: %w", err)
	}
	appState[wasmtypes.ModuleName] = wasmStateBz

	return appState, nil
}

// updateMobixContractState updates the config.unbonding_period entry of the mobix contract states (both staging and production)
// because the new cosmwasm v1.0.0 contract have changed the unit of this field from nanosecond to seconds.
func updateMobixContractState(appState types.AppMap, cdc codec.JSONMarshaler) (types.AppMap, error) {
	var wasmState wasmtypes.GenesisState
	if err := cdc.UnmarshalJSON(appState[wasmtypes.ModuleName], &wasmState); err != nil {
		return nil, fmt.Errorf("failed to unmarshal wasm genesis state: %w", err)
	}

	for cid, contract := range wasmState.Contracts {
		if contract.ContractInfo.CodeID != mobixContractCodeID {
			continue
		}

		for eid, entry := range contract.ContractState {
			if !bytes.Equal(entry.Key.Bytes(), []byte("config")) {
				continue
			}

			configMap := make(map[string]interface{})
			if err := json.Unmarshal(entry.Value, &configMap); err != nil {
				return nil, err
			}

			unbondingPeriod, err := strconv.ParseInt(configMap["unbonding_period"].(string), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to read unbonding_period: %w", err)
			}
			configMap["unbonding_period"] = fmt.Sprintf("%d", int64(time.Duration(unbondingPeriod).Seconds()))

			configMapBz, err := json.Marshal(configMap)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal mobix config: %w", err)
			}

			wasmState.Contracts[cid].ContractState[eid].Value = configMapBz
		}
	}

	wasmStateBz, err := cdc.MarshalJSON(&wasmState)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal wasm genesis state: %w", err)
	}
	appState[wasmtypes.ModuleName] = wasmStateBz

	return appState, nil
}

func setNewSlashFractionDoubleSign(appState types.AppMap, cdc codec.JSONMarshaler, newValue sdk.Dec) (types.AppMap, error) {
	var slashingState slashingtypes.GenesisState
	if err := cdc.UnmarshalJSON(appState[slashingtypes.ModuleName], &slashingState); err != nil {
		return nil, fmt.Errorf("failed to unmarshal slashing genesis state: %w", err)
	}

	slashingState.Params.SlashFractionDoubleSign = newValue

	slashingStateBz, err := cdc.MarshalJSON(&slashingState)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal slashing genesis state: %w", err)
	}
	appState[slashingtypes.ModuleName] = slashingStateBz

	return appState, nil
}
