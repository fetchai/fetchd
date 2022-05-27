package cmd

import (
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
				delegatedCoin = sdk.NewCoin(totalAmount.Denom, sdk.NewInt(0))
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

func addDelegation(cdc codec.JSONCodec, appState map[string]json.RawMessage, userAddr sdk.AccAddress, valAddr sdk.ValAddress, delegatedCoin sdk.Coin, currentHeight uint64) (map[string]json.RawMessage, error) {
	stakingState := stakingtypes.GetGenesisStateFromAppState(cdc, appState)
	shares := sdk.Dec(delegatedCoin.Amount.Mul(sdk.DefaultPowerReduction))

	var currentDelegation *stakingtypes.Delegation
	// check if user already delegated to this validator
	for i, delegation := range stakingState.Delegations {
		if delegation.GetDelegatorAddr().Equals(userAddr) &&
			delegation.GetValidatorAddr().Equals(valAddr) {
			currentDelegation = &stakingState.Delegations[i]
			break
		}
	}

	if currentDelegation == nil {
		// create a new delegation
		delegation := stakingtypes.NewDelegation(userAddr, valAddr, shares)
		stakingState.Delegations = append(stakingState.Delegations, delegation)
	} else {
		// increment existing delegation shares
		currentDelegation.Shares = currentDelegation.Shares.Add(shares)
	}

	// increment validator delegator_shares and token amount
	var currentValidator *stakingtypes.Validator
	for i, v := range stakingState.Validators {
		if v.OperatorAddress == valAddr.String() {
			currentValidator = &stakingState.Validators[i]

			break
		}
	}
	if currentValidator == nil {
		return nil, fmt.Errorf("failed to update validator: could not find validator %q", valAddr.String())
	}

	currentValidator.DelegatorShares = currentValidator.DelegatorShares.Add(shares)
	currentValidator.Tokens = currentValidator.Tokens.Add(delegatedCoin.Amount)

	stakingStateBz, err := cdc.MarshalJSON(stakingState)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal staking genesis state: %w", err)
	}
	appState[stakingtypes.ModuleName] = stakingStateBz

	// update distribution

	distributionState := distributiontypes.GetGenesisStateFromAppState(cdc, appState)

	var currentValidatorRewards *distributiontypes.ValidatorCurrentRewardsRecord
	for i, cur := range distributionState.ValidatorCurrentRewards {
		if cur.ValidatorAddress == valAddr.String() {
			currentValidatorRewards = &distributionState.ValidatorCurrentRewards[i]
			break
		}
	}
	if currentValidatorRewards == nil {
		return nil, fmt.Errorf("failed to retrieve validator current reward: cannt find current reward for %q", valAddr.String())
	}

	currentPeriod := currentValidatorRewards.Rewards.Period

	startingInfosExists := false
	var startingInfosPrevPeriod uint64
	// retrieve existing distribution info for the delegator / validator couple if any
	// otherwise just append a new one
	for i, r := range distributionState.DelegatorStartingInfos {
		if r.DelegatorAddress == userAddr.String() && r.ValidatorAddress == valAddr.String() {
			// keep the previous period to decrement the historical record reference_count
			startingInfosPrevPeriod = distributionState.DelegatorStartingInfos[i].StartingInfo.PreviousPeriod
			distributionState.DelegatorStartingInfos[i].StartingInfo.PreviousPeriod = currentPeriod
			distributionState.DelegatorStartingInfos[i].StartingInfo.Stake = r.StartingInfo.Stake.Add(shares)
			distributionState.DelegatorStartingInfos[i].StartingInfo.Height = currentHeight
			startingInfosExists = true
			break
		}
	}
	if !startingInfosExists {
		distributionState.DelegatorStartingInfos = append(distributionState.DelegatorStartingInfos, distributiontypes.DelegatorStartingInfoRecord{
			DelegatorAddress: userAddr.String(),
			ValidatorAddress: valAddr.String(),
			StartingInfo: distributiontypes.DelegatorStartingInfo{
				PreviousPeriod: currentPeriod,
				Stake:          shares,
				Height:         currentHeight,
			},
		})
	}

	// update validator historical rewards
	// same logic as in the keeper: https://github.com/fetchai/cosmos-sdk/blob/83a838df248ec012904c5ede1ff6381045f689ea/x/distribution/keeper/validator.go#L28
	var lastHistoricalRecord *distributiontypes.ValidatorHistoricalRewardsRecord
	deleteHistoricalRecordIndex := -1
	for i, rec := range distributionState.ValidatorHistoricalRewards {
		if rec.ValidatorAddress != valAddr.String() {
			continue // ignore records of other than current validator
		}

		if lastHistoricalRecord == nil || lastHistoricalRecord.Period < rec.Period {
			lastHistoricalRecord = &distributionState.ValidatorHistoricalRewards[i]
		}

		// when startingInfo already existed, this means its previous period was updated to the current period
		// so we must decrement the number of references held by the historical records
		// on this period.
		if startingInfosExists && rec.Period == startingInfosPrevPeriod {
			distributionState.ValidatorHistoricalRewards[i].Rewards.ReferenceCount--
			if distributionState.ValidatorHistoricalRewards[i].Rewards.ReferenceCount == 0 {
				// mark the historical record when we have no more reference to it for deletion
				deleteHistoricalRecordIndex = i
			}
		}
	}
	if lastHistoricalRecord == nil {
		return nil, fmt.Errorf("failed to retrieve validator historical reward records: cannot find historical reward records for %q", valAddr.String())
	}

	// removes the "validator" reference on the last historical record
	// it will get added back to the new record we'll insert
	lastHistoricalRecord.Rewards.ReferenceCount--

	if deleteHistoricalRecordIndex >= 0 {
		distributionState.ValidatorHistoricalRewards = append(
			distributionState.ValidatorHistoricalRewards[:deleteHistoricalRecordIndex],
			distributionState.ValidatorHistoricalRewards[deleteHistoricalRecordIndex+1:]...,
		)
	}

	currentRatio := currentValidatorRewards.Rewards.Rewards.QuoDecTruncate(currentValidator.Tokens.Sub(delegatedCoin.Amount).ToDec())
	newRatio := lastHistoricalRecord.Rewards.CumulativeRewardRatio.Add(currentRatio...)

	distributionState.ValidatorHistoricalRewards = append(distributionState.ValidatorHistoricalRewards, distributiontypes.ValidatorHistoricalRewardsRecord{
		ValidatorAddress: valAddr.String(),
		Period:           currentPeriod,
		Rewards:          distributiontypes.NewValidatorHistoricalRewards(newRatio, 2), // 2 referenceCount => 1 delegator + 1 validator
	})

	currentValidatorRewards.Rewards = distributiontypes.NewValidatorCurrentRewards(sdk.DecCoins{}, currentPeriod+1)

	distributionStateBz, err := cdc.MarshalJSON(distributionState)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal distribution genesis state: %w", err)
	}
	appState[distributiontypes.ModuleName] = distributionStateBz

	return appState, nil
}
