package cmd

import (
	"cosmossdk.io/math"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	// defaultMinDelegatedAmount defines a minimum amount required to add a delegation.
	defaultMinDelegatedAmount = "2000000000000000000afet"

	// defaultAccountReservedAmount defines the amount kept on the user account (not delegated)
	defaultAccountReservedAmount = "1000000000000000000afet"
)

const (
	flagMinDelegatedAmount    = "min-delegated-amount"
	flagAccountReservedAmount = "account-reserved-amount"
)

// AddGenesisDelegationCmd returns a command to add delegations to genesis.
func AddGenesisDelegationCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add-genesis-delegation [address_or_key_name] [validator_address] [amount]",
		Short: "Create a genesis account and try to create a genesis delegation.",
		Long: `Create a genesis account and try to create a genesis delegation.
> when amount is greater than or equal to <min-delegated-amount>, <account-reserved-amount>
will be subtracted and added to the account balance 
(to allow users to pay transaction fees to redelegate or unbond...). 
Remaining tokens will be delegated to the chosen validator.
> when amount is lower than <min-delegated-amount>, no delegation is created, 
and the full amount is stored on the account balance.`,
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			cdc := clientCtx.Codec

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			minDelegatedAmountStr, err := cmd.Flags().GetString(flagMinDelegatedAmount)
			if err != nil {
				return fmt.Errorf("failed to get flag %q: %w", flagMinDelegatedAmount, err)
			}
			accountReservedAmountStr, err := cmd.Flags().GetString(flagAccountReservedAmount)
			if err != nil {
				return fmt.Errorf("failed to get flag %q: %w", flagAccountReservedAmount, err)
			}

			minDelegatedCoin, err := sdk.ParseCoinNormalized(minDelegatedAmountStr)
			if err != nil {
				return fmt.Errorf("failed to parse coin from minDelegatedAmount: %w", err)
			}
			accountReservedCoin, err := sdk.ParseCoinNormalized(accountReservedAmountStr)
			if err != nil {
				return fmt.Errorf("failed to parse coin from accountReservedAmount: %w", err)
			}

			totalAmount, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return fmt.Errorf("failed to parse amount: %w", err)
			}

			if totalAmount.Denom != minDelegatedCoin.Denom ||
				totalAmount.Denom != accountReservedCoin.Denom ||
				minDelegatedCoin.Denom != accountReservedCoin.Denom {
				return errors.New("amount denom mismatch, all amounts must share same denom")
			}

			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return fmt.Errorf("failed to parse bech32 account address: %w", err)
			}

			valAddr, err := sdk.ValAddressFromBech32(args[1])
			if err != nil {
				return fmt.Errorf("failed to parse bech32 validator address: %w", err)
			}

			var delegatedCoin sdk.Coin
			var accountCoin sdk.Coin
			// determine if amount is enough to create a validation and leave some
			// tokens on the user account.
			// Otherwise just send it all to the user account and skip delegation
			if totalAmount.IsGTE(minDelegatedCoin) {
				delegatedCoin = totalAmount.Sub(accountReservedCoin)
				accountCoin = accountReservedCoin
			} else {
				delegatedCoin = sdk.NewCoin(totalAmount.Denom, math.NewInt(0))
				accountCoin = totalAmount
			}

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			// update auth

			authGenState := authtypes.GetGenesisStateFromAppState(cdc, appState)

			accs, err := authtypes.UnpackAccounts(authGenState.Accounts)
			if err != nil {
				return fmt.Errorf("failed to get accounts from any: %w", err)
			}

			// add the delegator account to the authState when it is not in it yet
			if !accs.Contains(addr) {
				accs = append(accs, authtypes.NewBaseAccount(addr, nil, 0, 0))
				accs = authtypes.SanitizeGenesisAccounts(accs)
			}

			genAccs, err := authtypes.PackAccounts(accs)
			if err != nil {
				return fmt.Errorf("failed to convert accounts into any's: %w", err)
			}
			authGenState.Accounts = genAccs

			authGenStateBz, err := cdc.MarshalJSON(&authGenState)
			if err != nil {
				return fmt.Errorf("failed to marshal auth genesis state: %w", err)
			}
			appState[authtypes.ModuleName] = authGenStateBz

			// add a delegation if amount allows it, otherwise just send it all to the user account.
			if !delegatedCoin.IsZero() {
				appState, err = addDelegation(cdc, appState, addr, valAddr, delegatedCoin, uint64(genDoc.InitialHeight))
				if err != nil {
					return fmt.Errorf("failed to add delegation: %w", err)
				}
			}

			// update bank

			bankState := banktypes.GetGenesisStateFromAppState(cdc, appState)

			// increment bonded pool account
			successUpdatingBondedPool := false
			updatedUserBank := false
			bondedPoolAddr := authtypes.NewModuleAddress(stakingtypes.BondedPoolName)
			for i, balance := range bankState.Balances {
				switch balance.Address {
				case bondedPoolAddr.String():
					// add delegatedAmount to the bondedPool balance - might be zero.
					bankState.Balances[i].Coins = bankState.Balances[i].Coins.Add(delegatedCoin)
					successUpdatingBondedPool = true
				case addr.String():
					bankState.Balances[i].Coins = bankState.Balances[i].Coins.Add(accountCoin)
					updatedUserBank = true
				}
			}
			if !successUpdatingBondedPool {
				return fmt.Errorf("failed to update bonded pool balance: cannot find account %q", bondedPoolAddr.String())
			}

			// user does not have a balance yet, so we create it
			if !updatedUserBank {
				bankState.Balances = append(bankState.Balances, banktypes.Balance{
					Address: addr.String(),
					Coins:   sdk.NewCoins(accountCoin),
				})
			}

			// increment total supply by the total amount of new tokens
			bankState.Supply = bankState.Supply.Add(totalAmount)

			bankStateBz, err := cdc.MarshalJSON(bankState)
			if err != nil {
				return fmt.Errorf("failed to marshal bank genesis state: %w", err)
			}
			appState[banktypes.ModuleName] = bankStateBz

			// Encode back the genesis state to json
			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flagMinDelegatedAmount, defaultMinDelegatedAmount, "minimum amount required to create a delegation")
	cmd.Flags().String(flagAccountReservedAmount, defaultAccountReservedAmount, "amount subtracted from the delegated amount and transferred on the user account when a delegation is created")

	return cmd
}

func mustGetStakingGenesis(cdc codec.JSONCodec, appState map[string]json.RawMessage) (stakingtypes.GenesisState, error) {
	var gs stakingtypes.GenesisState
	if bz := appState[stakingtypes.ModuleName]; len(bz) > 0 {
		if err := cdc.UnmarshalJSON(bz, &gs); err != nil {
			return stakingtypes.GenesisState{}, fmt.Errorf("unmarshal staking genesis: %w", err)
		}
	} else {
		gs = *stakingtypes.DefaultGenesisState()
	}
	return gs, nil
}

func mustGetDistrGenesis(cdc codec.JSONCodec, appState map[string]json.RawMessage) (distributiontypes.GenesisState, error) {
	var gs distributiontypes.GenesisState
	if bz := appState[distributiontypes.ModuleName]; len(bz) > 0 {
		if err := cdc.UnmarshalJSON(bz, &gs); err != nil {
			return distributiontypes.GenesisState{}, fmt.Errorf("unmarshal distribution genesis: %w", err)
		}
	} else {
		gs = *distributiontypes.DefaultGenesisState()
	}
	return gs, nil
}

func addDelegation(
	cdc codec.JSONCodec,
	appState map[string]json.RawMessage,
	userAddr sdk.AccAddress,
	valAddr sdk.ValAddress,
	delegatedCoin sdk.Coin,
	currentHeight uint64,
) (map[string]json.RawMessage, error) {

	// --- STAKING: load & mutate genesis ---
	stakingState, err := mustGetStakingGenesis(cdc, appState)
	if err != nil {
		return nil, err
	}

	// find validator
	var currentValidator *stakingtypes.Validator
	for i := range stakingState.Validators {
		if stakingState.Validators[i].OperatorAddress == valAddr.String() {
			currentValidator = &stakingState.Validators[i]
			break
		}
	}
	if currentValidator == nil {
		return nil, fmt.Errorf("failed to update validator: could not find validator %q", valAddr.String())
	}

	// compute shares: if validator has existing tokens+shares, use exchange rate S/T; otherwise 1:1
	var shares math.LegacyDec
	amountDec := math.LegacyNewDecFromInt(delegatedCoin.Amount)
	if currentValidator.Tokens.IsZero() || currentValidator.DelegatorShares.IsZero() {
		shares = amountDec
	} else {
		// shares = amount * totalShares / totalTokens
		shares = amountDec.Mul(currentValidator.DelegatorShares).QuoInt(currentValidator.Tokens)
	}

	// check existing delegation
	var currentDelegation *stakingtypes.Delegation
	for i := range stakingState.Delegations {
		d := &stakingState.Delegations[i]
		if d.DelegatorAddress == userAddr.String() && d.ValidatorAddress == valAddr.String() {
			currentDelegation = d
			break
		}
	}

	if currentDelegation == nil {
		// NOTE: NewDelegation takes bech32 strings + LegacyDec shares in new SDK.
		delegation := stakingtypes.NewDelegation(userAddr.String(), valAddr.String(), shares)
		stakingState.Delegations = append(stakingState.Delegations, delegation)
	} else {
		currentDelegation.Shares = currentDelegation.Shares.Add(shares)
	}

	// bump validator totals
	currentValidator.DelegatorShares = currentValidator.DelegatorShares.Add(shares)
	currentValidator.Tokens = currentValidator.Tokens.Add(delegatedCoin.Amount)

	// write back staking
	if bz, err := cdc.MarshalJSON(&stakingState); err != nil {
		return nil, fmt.Errorf("failed to marshal staking genesis state: %w", err)
	} else {
		appState[stakingtypes.ModuleName] = bz
	}

	// --- DISTRIBUTION: load & mutate genesis ---
	distributionState, err := mustGetDistrGenesis(cdc, appState)
	if err != nil {
		return nil, err
	}

	// find validator current rewards
	var currentValidatorRewards *distributiontypes.ValidatorCurrentRewardsRecord
	for i := range distributionState.ValidatorCurrentRewards {
		if distributionState.ValidatorCurrentRewards[i].ValidatorAddress == valAddr.String() {
			currentValidatorRewards = &distributionState.ValidatorCurrentRewards[i]
			break
		}
	}
	if currentValidatorRewards == nil {
		return nil, fmt.Errorf("failed to retrieve validator current reward for %q", valAddr.String())
	}

	currentPeriod := currentValidatorRewards.Rewards.Period

	// ensure/adjust DelegatorStartingInfo
	var (
		startingInfosExists     = false
		startingInfosPrevPeriod uint64
	)
	for i := range distributionState.DelegatorStartingInfos {
		r := &distributionState.DelegatorStartingInfos[i]
		if r.DelegatorAddress == userAddr.String() && r.ValidatorAddress == valAddr.String() {
			startingInfosPrevPeriod = r.StartingInfo.PreviousPeriod
			r.StartingInfo.PreviousPeriod = currentPeriod
			r.StartingInfo.Stake = r.StartingInfo.Stake.Add(shares)
			r.StartingInfo.Height = currentHeight
			startingInfosExists = true
			break
		}
	}
	if !startingInfosExists {
		distributionState.DelegatorStartingInfos = append(
			distributionState.DelegatorStartingInfos,
			distributiontypes.DelegatorStartingInfoRecord{
				DelegatorAddress: userAddr.String(),
				ValidatorAddress: valAddr.String(),
				StartingInfo: distributiontypes.DelegatorStartingInfo{
					PreviousPeriod: currentPeriod,
					Stake:          shares,
					Height:         currentHeight,
				},
			},
		)
	}

	// update validator historical rewards
	var lastHistoricalRecord *distributiontypes.ValidatorHistoricalRewardsRecord
	deleteHistoricalRecordIndex := -1
	for i := range distributionState.ValidatorHistoricalRewards {
		rec := &distributionState.ValidatorHistoricalRewards[i]
		if rec.ValidatorAddress != valAddr.String() {
			continue
		}
		if lastHistoricalRecord == nil || lastHistoricalRecord.Period < rec.Period {
			lastHistoricalRecord = rec
		}
		if startingInfosExists && rec.Period == startingInfosPrevPeriod {
			rec.Rewards.ReferenceCount--
			if rec.Rewards.ReferenceCount == 0 {
				deleteHistoricalRecordIndex = i
			}
		}
	}
	if lastHistoricalRecord == nil {
		return nil, fmt.Errorf("failed to retrieve validator historical reward records for %q", valAddr.String())
	}

	// remove the "validator" reference on the last historical record (will be added to the new one)
	if lastHistoricalRecord.Rewards.ReferenceCount > 0 {
		lastHistoricalRecord.Rewards.ReferenceCount--
	}

	if deleteHistoricalRecordIndex >= 0 {
		distributionState.ValidatorHistoricalRewards = append(
			distributionState.ValidatorHistoricalRewards[:deleteHistoricalRecordIndex],
			distributionState.ValidatorHistoricalRewards[deleteHistoricalRecordIndex+1:]...,
		)
	}

	denom := currentValidator.Tokens.Sub(delegatedCoin.Amount) // tokens BEFORE this delegation
	if denom.IsZero() {
		denom = math.OneInt() // avoid div-by-zero; ratio will be "all to new period"
	}
	currentRatio := currentValidatorRewards.Rewards.Rewards.QuoDecTruncate(
		math.LegacyNewDecFromInt(denom),
	)
	newRatio := lastHistoricalRecord.Rewards.CumulativeRewardRatio.Add(currentRatio...)

	// add new historical record; 2 refs => 1 delegator + 1 validator
	distributionState.ValidatorHistoricalRewards = append(
		distributionState.ValidatorHistoricalRewards,
		distributiontypes.ValidatorHistoricalRewardsRecord{
			ValidatorAddress: valAddr.String(),
			Period:           currentPeriod,
			Rewards:          distributiontypes.NewValidatorHistoricalRewards(newRatio, 2),
		},
	)

	// bump current rewards period
	currentValidatorRewards.Rewards = distributiontypes.NewValidatorCurrentRewards(sdk.DecCoins{}, currentPeriod+1)

	// write back distribution
	if bz, err := cdc.MarshalJSON(&distributionState); err != nil {
		return nil, fmt.Errorf("failed to marshal distribution genesis state: %w", err)
	} else {
		appState[distributiontypes.ModuleName] = bz
	}

	return appState, nil
}
