package app

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/spf13/cast"
	"github.com/tendermint/tendermint/libs/log"
	"math"
	"sort"
)

type DelegatorStartingInfo struct {
	height         uint64
	previousPeriod uint64
	stake          sdk.Dec
}

type ValidatorSlashEvent struct {
	period          uint64
	fraction        sdk.Dec
	validatorPeriod uint64
}

type ValidatorHistoricalReward struct {
	cumulativeRewardRatio sdk.DecCoins
}

type ValidatorCurrentReward struct {
	period uint64
	reward sdk.DecCoins
}

type FeePool struct {
	CommunityPool sdk.DecCoins
}

type DistributionInfo struct {
	FeePool            *FeePool
	OutstandingRewards *OrderedMap[string, sdk.DecCoins]

	// params
	// previousProposer string

	ValidatorAccumulatedCommissions *OrderedMap[string, sdk.DecCoins]                                    // validator_addr -> validator_accumulated_commissions
	ValidatorCurrentRewards         *OrderedMap[string, *ValidatorCurrentReward]                         // validator_addr -> validator_current_rewards
	ValidatorHistoricalRewards      *OrderedMap[string, *OrderedMap[uint64, *ValidatorHistoricalReward]] // validator_addr -> period -> validator_historical_reward

	DelegatorStartingInfos           *OrderedMap[string, *OrderedMap[string, *DelegatorStartingInfo]] // validator_addr -> delegator_addr -> starting_info
	DelegatorWithdrawInfos           *OrderedMap[string, string]                                      // delegator_address -> withdraw_address
	ValidatorSlashEvents             *OrderedMap[string, *OrderedMap[uint64, *ValidatorSlashEvent]]   // validator_address -> height -> validator_slash_event
	DistributionModuleAccountAddress string

	// Aggregated values
	Rewards          *OrderedMap[string, *OrderedMap[string, sdk.DecCoins]] // delegator_addr -> validator_addr -> reward
	ValidatorRewards *OrderedMap[string, sdk.DecCoins]                      // validator_addr -> validator_rewards
}

func parseDelegatorStartingInfos(distribution map[string]interface{}) (*OrderedMap[string, *OrderedMap[string, *DelegatorStartingInfo]], error) {

	delegatorStartingInfos := NewOrderedMap[string, *OrderedMap[string, *DelegatorStartingInfo]]()
	delegatorStartingInfosMap := distribution["delegator_starting_infos"].([]interface{})
	for _, info := range delegatorStartingInfosMap {
		delegatorStartingInfoMap := info.(map[string]interface{})

		validatorAddress := delegatorStartingInfoMap["validator_address"].(string)
		delegatorAddress := delegatorStartingInfoMap["delegator_address"].(string)
		startInfoMap := delegatorStartingInfoMap["starting_info"].(map[string]interface{})

		var delegatorStartingInfo DelegatorStartingInfo
		delegatorStartingInfo.height = cast.ToUint64(startInfoMap["height"].(string))
		delegatorStartingInfo.previousPeriod = cast.ToUint64(startInfoMap["previous_period"].(string))

		stake := startInfoMap["stake"].(string)
		stakeDec, err := sdk.NewDecFromStr(stake)
		if err != nil {
			return nil, err
		}
		delegatorStartingInfo.stake = stakeDec

		valStartingInfo, _ := delegatorStartingInfos.GetOrSetDefault(validatorAddress, NewOrderedMap[string, *DelegatorStartingInfo]())

		valStartingInfo.Set(delegatorAddress, &delegatorStartingInfo)
		delegatorStartingInfos.Set(validatorAddress, valStartingInfo)

	}
	return delegatorStartingInfos, nil
}

func parseValidatorHistoricalRewards(distribution map[string]interface{}) (*OrderedMap[string, *OrderedMap[uint64, *ValidatorHistoricalReward]], error) {

	validatorHistoricalRewards := NewOrderedMap[string, *OrderedMap[uint64, *ValidatorHistoricalReward]]()
	validatorHistoricalRewardsMap := distribution["validator_historical_rewards"].([]interface{})
	for _, info := range validatorHistoricalRewardsMap {
		var delegatorStartingInfo ValidatorHistoricalReward

		delegatorStartingInfoMap := info.(map[string]interface{})
		validatorAddress := delegatorStartingInfoMap["validator_address"].(string)
		period := cast.ToUint64(delegatorStartingInfoMap["period"].(string))
		rewards := delegatorStartingInfoMap["rewards"].(map[string]interface{})

		cumulativeRewardRatio, err := getDecCoinsFromInterfaceSlice(rewards["cumulative_reward_ratio"].([]interface{}))
		if err != nil {
			return nil, err
		}

		delegatorStartingInfo.cumulativeRewardRatio = cumulativeRewardRatio

		valRewards, _ := validatorHistoricalRewards.GetOrSetDefault(validatorAddress, NewOrderedMap[uint64, *ValidatorHistoricalReward]())

		valRewards.SetNew(period, &delegatorStartingInfo)
		//validatorHistoricalRewards.Set(validatorAddress, *valRewards)

	}
	return validatorHistoricalRewards, nil
}

func parseValidatorCurrentRewards(distribution map[string]interface{}) (*OrderedMap[string, *ValidatorCurrentReward], error) {

	validatorCurrentRewards := NewOrderedMap[string, *ValidatorCurrentReward]()
	validatorCurrentRewardsMap := distribution["validator_current_rewards"].([]interface{})
	for _, info := range validatorCurrentRewardsMap {
		var validatorCurrentReward ValidatorCurrentReward

		validatorCurrentRewardMap := info.(map[string]interface{})
		validatorAddress := validatorCurrentRewardMap["validator_address"].(string)
		rewardsMap := validatorCurrentRewardMap["rewards"].(map[string]interface{})

		period := cast.ToUint64(rewardsMap["period"].(string))

		rewards, err := getDecCoinsFromInterfaceSlice(rewardsMap["rewards"].([]interface{}))
		if err != nil {
			return nil, err
		}

		validatorCurrentReward.reward = rewards
		validatorCurrentReward.period = period

		validatorCurrentRewards.SetNew(validatorAddress, &validatorCurrentReward)

	}
	return validatorCurrentRewards, nil
}

func parseOutstandingRewards(distribution map[string]interface{}) (*OrderedMap[string, sdk.DecCoins], error) {

	OutstandingRewards := NewOrderedMap[string, sdk.DecCoins]()
	OutstandingRewardsMap := distribution["outstanding_rewards"].([]interface{})
	for _, info := range OutstandingRewardsMap {

		validatorCurrentRewardMap := info.(map[string]interface{})
		validatorAddress := validatorCurrentRewardMap["validator_address"].(string)

		outstandingRewardsCoins, err := getDecCoinsFromInterfaceSlice(validatorCurrentRewardMap["outstanding_rewards"].([]interface{}))
		if err != nil {
			return nil, err
		}

		OutstandingRewards.SetNew(validatorAddress, outstandingRewardsCoins)

	}
	return OutstandingRewards, nil
}

func parseValidatorAccumulatedCommissions(distribution map[string]interface{}) (*OrderedMap[string, sdk.DecCoins], error) {

	validatorAccumulatedCommissions := NewOrderedMap[string, sdk.DecCoins]()
	validatorAccumulatedCommissionsMap := distribution["validator_accumulated_commissions"].([]interface{})
	for _, info := range validatorAccumulatedCommissionsMap {

		validatorCurrentRewardMap := info.(map[string]interface{})
		validatorAddress := validatorCurrentRewardMap["validator_address"].(string)

		accumulatedCommissionsCoins, err := getDecCoinsFromInterfaceSlice(validatorCurrentRewardMap["accumulated"].(map[string]interface{})["commission"].([]interface{}))
		if err != nil {
			return nil, err
		}

		validatorAccumulatedCommissions.SetNew(validatorAddress, accumulatedCommissionsCoins)

	}
	return validatorAccumulatedCommissions, nil
}

func parseValidatorSlashEvents(distribution map[string]interface{}) (*OrderedMap[string, *OrderedMap[uint64, *ValidatorSlashEvent]], error) {

	validatorSlashEvents := NewOrderedMap[string, *OrderedMap[uint64, *ValidatorSlashEvent]]()
	validatorSlashEventsMap := distribution["validator_slash_events"].([]interface{})
	for _, info := range validatorSlashEventsMap {
		var delegatorStartingInfo ValidatorSlashEvent

		delegatorStartingInfoMap := info.(map[string]interface{})
		validatorAddress := delegatorStartingInfoMap["validator_address"].(string)
		period := cast.ToUint64(delegatorStartingInfoMap["period"].(string))
		height := cast.ToUint64(delegatorStartingInfoMap["height"].(string))

		slashEvent := delegatorStartingInfoMap["validator_slash_event"].(map[string]interface{})

		fraction, err := sdk.NewDecFromStr(slashEvent["fraction"].(string))
		if err != nil {
			return nil, err
		}

		delegatorStartingInfo.fraction = fraction
		delegatorStartingInfo.period = period
		delegatorStartingInfo.validatorPeriod = cast.ToUint64(slashEvent["validator_period"].(string))

		valEvents, _ := validatorSlashEvents.GetOrSetDefault(validatorAddress, NewOrderedMap[uint64, *ValidatorSlashEvent]())

		// TODO(pb): Is this necessary? It basically re-validates data in genesis file, which we should treat & take
		//           as granted to be correct by definition.
		if delegatorStartingInfo.validatorPeriod != delegatorStartingInfo.period {
			delegatorAddress := delegatorStartingInfoMap["delegator_address"].(string)
			return nil, fmt.Errorf("delegator %v period %v does not match associated validator %v period %v", delegatorAddress, delegatorStartingInfo.period, validatorAddress, delegatorStartingInfo.validatorPeriod)
		}

		valEvents.SetNew(height, &delegatorStartingInfo)
		//validatorSlashEvents.Set(validatorAddress, *valEvents)

	}
	return validatorSlashEvents, nil
}

func parseFeePool(distribution map[string]interface{}) (*FeePool, error) {
	feePool := distribution["fee_pool"].(map[string]interface{})
	communityPool, err := getDecCoinsFromInterfaceSlice(feePool["community_pool"].([]interface{}))
	if err != nil {
		return nil, err
	}

	return &FeePool{CommunityPool: communityPool}, nil
}

func parseDelegatorWithdrawInfos(distribution map[string]interface{}) (*OrderedMap[string, string], error) {
	delegatorWithdrawInfos := NewOrderedMap[string, string]()

	infos := distribution["delegator_withdraw_infos"].([]interface{})

	for _, info := range infos {
		delegatorWithdrawInfoMap := info.(map[string]interface{})
		delegatorAddress := delegatorWithdrawInfoMap["delegator_address"].(string)
		withdrawAddress := delegatorWithdrawInfoMap["withdraw_address"].(string)
		delegatorWithdrawInfos.Set(delegatorAddress, withdrawAddress)
	}

	return delegatorWithdrawInfos, nil
}

func aggregateRewards(distributionInfo *DistributionInfo, validators *OrderedMap[string, *ValidatorInfo]) error {
	blockHeight := uint64(math.MaxUint64)

	if distributionInfo.Rewards == nil {
		distributionInfo.Rewards = NewOrderedMap[string, *OrderedMap[string, sdk.DecCoins]]()
	}
	if distributionInfo.ValidatorRewards == nil {
		distributionInfo.ValidatorRewards = NewOrderedMap[string, sdk.DecCoins]()
	}

	// Withdraw all delegation rewards
	for i := range distributionInfo.DelegatorStartingInfos.Iterate() {
		validatorOperatorAddr, delegatorStartInfo := i.Key, i.Value
		validator := validators.MustGet(validatorOperatorAddr)

		// Get initial outstanding reward
		outstandingRewards := distributionInfo.OutstandingRewards.MustGet(validatorOperatorAddr)

		endingPeriod := UpdateValidatorData(distributionInfo, validator)

		for _, delegatorAddr := range delegatorStartInfo.Keys() {
			delegation := validator.Delegations.MustGet(delegatorAddr)

			rewardsRaw, err := CalculateDelegationRewards(blockHeight, distributionInfo, validator, delegation, endingPeriod)
			if err != nil {
				return err
			}

			delegatorRewards, _ := distributionInfo.Rewards.GetOrSetDefault(delegatorAddr, NewOrderedMap[string, sdk.DecCoins]())
			delegatorRewards.SetNew(validatorOperatorAddr, rewardsRaw)

			outstandingRewards = outstandingRewards.Sub(rewardsRaw.Intersect(outstandingRewards))

		}

		validatorAccountAddress, err := convertOperatorAddressToAccount(validatorOperatorAddr)
		if err != nil {
			return err
		}

		distributionInfo.ValidatorRewards.Set(validatorAccountAddress, outstandingRewards)

	}

	return nil
}

func parseGenesisDistribution(jsonData map[string]interface{}, genesisAccounts *OrderedMap[string, *AccountInfo], validators *OrderedMap[string, *ValidatorInfo]) (*DistributionInfo, error) {
	distribution := jsonData[distributiontypes.ModuleName].(map[string]interface{})
	distributionInfo := DistributionInfo{}
	var err error

	distributionInfo.DelegatorStartingInfos, err = parseDelegatorStartingInfos(distribution)
	if err != nil {
		return nil, err
	}

	distributionInfo.ValidatorHistoricalRewards, err = parseValidatorHistoricalRewards(distribution)
	if err != nil {
		return nil, err
	}

	distributionInfo.ValidatorCurrentRewards, err = parseValidatorCurrentRewards(distribution)
	if err != nil {
		return nil, err
	}

	distributionInfo.ValidatorSlashEvents, err = parseValidatorSlashEvents(distribution)
	if err != nil {
		return nil, err
	}

	distributionInfo.OutstandingRewards, err = parseOutstandingRewards(distribution)
	if err != nil {
		return nil, err
	}

	distributionInfo.FeePool, err = parseFeePool(distribution)
	if err != nil {
		return nil, err
	}

	distributionInfo.DelegatorWithdrawInfos, err = parseDelegatorWithdrawInfos(distribution)
	if err != nil {
		return nil, err
	}

	distributionInfo.ValidatorAccumulatedCommissions, err = parseValidatorAccumulatedCommissions(distribution)
	if err != nil {
		return nil, err
	}

	distributionInfo.DistributionModuleAccountAddress, err = GetAddressByName(genesisAccounts, DistributionAccName)
	if err != nil {
		return nil, err
	}

	err = aggregateRewards(&distributionInfo, validators)
	if err != nil {
		return nil, err
	}

	return &distributionInfo, nil
}

func checkTolerance(coins sdk.Coins, maxToleratedDiff sdk.Int) error {
	for _, coin := range coins {
		if coin.Amount.GT(maxToleratedDiff) {
			return fmt.Errorf("remaining balance %s is too high", coin.String())
		}
	}
	return nil
}

func verifyOutstandingBalances(genesisData *GenesisData, cudosCfg *CudosMergeConfig) error {

	maxToleratedRemainingDistributionBalance := unwrapOrDefault(
		cudosCfg.Config.MaxToleratedRemainingDistributionBalance,
		DefaultMaxToleratedRemainingDistributionBalance,
	)

	for i := range genesisData.DistributionInfo.OutstandingRewards.Iterate() {
		validatorAddr, validatorOutstandingReward := i.Key, i.Value

		validatorAccumulatedCommission := genesisData.DistributionInfo.ValidatorAccumulatedCommissions.MustGet(validatorAddr)

		diff := validatorOutstandingReward.Sub(validatorAccumulatedCommission)

		err := checkDecTolerance(diff, maxToleratedRemainingDistributionBalance)
		if err != nil {
			return fmt.Errorf("outstanding balance of validator %s is too high: %w", validatorAddr, err)
		}
	}

	return nil
}

func withdrawGenesisDistributionRewards(logger log.Logger, genesisData *GenesisData, cudosCfg *CudosMergeConfig, manifest *UpgradeManifest) error {
	// Withdraw all delegation rewards
	for _, validatorOperatorAddr := range genesisData.DistributionInfo.DelegatorStartingInfos.Keys() {
		delegatorStartInfo := genesisData.DistributionInfo.DelegatorStartingInfos.MustGet(validatorOperatorAddr)

		for _, delegatorAddr := range delegatorStartInfo.Keys() {

			_, err := withdrawDelegationRewards(logger, genesisData, validatorOperatorAddr, delegatorAddr, cudosCfg, manifest)
			if err != nil {
				return err
			}
		}

	}

	// Check that remaining balance is equal to AccumulatedCommissions
	err := verifyOutstandingBalances(genesisData, cudosCfg)
	if err != nil {
		return err
	}

	// Withdraw validator accumulated commission
	//	err = withdrawAccumulatedCommissions(genesisData, cudosCfg, manifest)
	err = withdrawValidatorOutstandingRewards(genesisData, cudosCfg, manifest)
	if err != nil {
		return err
	}

	// Withdraw Community pool balance to address if defined
	communityBalance, _ := genesisData.DistributionInfo.FeePool.CommunityPool.TruncateDecimal()
	distributionModuleAccount := genesisData.Accounts.MustGet(genesisData.DistributionInfo.DistributionModuleAccountAddress)

	remainingBalance := distributionModuleAccount.Balance.Sub(communityBalance)

	// Withdraw remaining balance
	if logger != nil {
		logger.Info("cudos merge: remaining dist balance", "amount", remainingBalance.String())
	}

	maxToleratedRemainingDistributionBalance := unwrapOrDefault(
		cudosCfg.Config.MaxToleratedRemainingDistributionBalance,
		DefaultMaxToleratedRemainingDistributionBalance,
	)

	err = checkTolerance(remainingBalance, maxToleratedRemainingDistributionBalance)
	if err != nil {
		return fmt.Errorf("remaining distribution balance %s is too high", remainingBalance.String())
	}

	err = moveGenesisBalance(genesisData, genesisData.DistributionInfo.DistributionModuleAccountAddress, cudosCfg.Config.RemainingDistributionBalanceAddr, distributionModuleAccount.Balance, "remaining_distribution_module_balance", manifest, cudosCfg)
	if err != nil {
		return err
	}

	return nil
}

func withdrawAccumulatedCommissions(genesisData *GenesisData, cudosCfg *CudosMergeConfig, manifest *UpgradeManifest) error {

	for _, validatorAddress := range genesisData.DistributionInfo.ValidatorAccumulatedCommissions.Keys() {
		accumulatedCommission := genesisData.DistributionInfo.ValidatorAccumulatedCommissions.MustGet(validatorAddress)

		accountAddress, err := ConvertAddressPrefix(validatorAddress, genesisData.Prefix)
		if err != nil {
			return err
		}

		finalRewards, _ := accumulatedCommission.TruncateDecimal()

		err = moveGenesisBalance(genesisData, genesisData.DistributionInfo.DistributionModuleAccountAddress, accountAddress, finalRewards, "accumulated_commission", manifest, cudosCfg)
		if err != nil {
			return err
		}
	}

	return nil
}

func withdrawValidatorOutstandingRewards(genesisData *GenesisData, cudosCfg *CudosMergeConfig, manifest *UpgradeManifest) error {

	for _, validatorAddress := range genesisData.DistributionInfo.OutstandingRewards.Keys() {
		outstandingRewards := genesisData.DistributionInfo.OutstandingRewards.MustGet(validatorAddress)

		accountAddress, err := ConvertAddressPrefix(validatorAddress, genesisData.Prefix)
		if err != nil {
			return err
		}

		finalRewards, _ := outstandingRewards.TruncateDecimal()

		err = moveGenesisBalance(genesisData, genesisData.DistributionInfo.DistributionModuleAccountAddress, accountAddress, finalRewards, "outstanding_rewards", manifest, cudosCfg)
		if err != nil {
			return err
		}
	}

	return nil
}

// calculate the rewards accrued by a delegation between two periods
func calculateDelegationRewardsBetween(distributionInfo *DistributionInfo, val *ValidatorInfo,
	startingPeriod, endingPeriod uint64, stake sdk.Dec,
) (sdk.DecCoins, error) {
	// sanity check
	if startingPeriod > endingPeriod {
		return sdk.DecCoins{}, fmt.Errorf("startingPeriod cannot be greater than endingPeriod")
	}

	// sanity check
	if stake.IsNegative() {
		return sdk.DecCoins{}, fmt.Errorf("stake should not be negative")
	}

	// return staking * (ending - starting)

	operatorRewards := distributionInfo.ValidatorHistoricalRewards.MustGet(val.OperatorAddress)
	starting := operatorRewards.MustGet(startingPeriod)
	ending := operatorRewards.MustGet(endingPeriod)

	difference := ending.cumulativeRewardRatio.Sub(starting.cumulativeRewardRatio)
	if difference.IsAnyNegative() {
		return sdk.DecCoins{}, fmt.Errorf("negative rewards should not be possible")
	}
	// note: necessary to truncate so we don't allow withdrawing more rewards than owed
	rewards := difference.MulDecTruncate(stake)
	return rewards, nil
}

// iterate over slash events between heights, inclusive
func IterateValidatorSlashEventsBetween(distributionInfo *DistributionInfo, val string, startingHeight uint64, endingHeight uint64,
	handler func(height uint64, event *ValidatorSlashEvent) (stop bool, err error),
) error {
	slashEvents, exists := distributionInfo.ValidatorSlashEvents.Get(val)
	// No slashing events
	if !exists {
		return nil
	}

	// Ensure correct order of iteration, won't have any effect if keys are already sorted
	sortUint64Keys(slashEvents)

	keys := slashEvents.Keys()

	// Perform binary search to find the starting point
	startIdx := sort.Search(len(keys), func(i int) bool {
		return keys[i] >= startingHeight
	})

	stop := false
	var err error = nil
	// Iterate from the startIdx up to the ending height
	for i := startIdx; !stop && err == nil && i < len(keys); i++ {
		height := keys[i]
		if height > endingHeight {
			break
		}

		event := slashEvents.MustGet(height)
		stop, err = handler(height, event)
	}
	return err
}

// calculate the total rewards accrued by a delegation
func CalculateDelegationRewards(blockHeight uint64, distributionInfo *DistributionInfo, val *ValidatorInfo, del *DelegationInfo, endingPeriod uint64) (rewards sdk.DecCoins, err error) {
	// fetch starting info for delegation
	delStartingInfo := distributionInfo.DelegatorStartingInfos.MustGet(val.OperatorAddress)
	startingInfo := delStartingInfo.MustGet(del.DelegatorAddress)

	if startingInfo.height == blockHeight {
		// started this height, no rewards yet
		return rewards, nil
	}

	startingPeriod := startingInfo.previousPeriod
	stake := startingInfo.stake

	// Iterate through slashes and withdraw with calculated staking for
	// distribution periods. These period offsets are dependent on *when* slashes
	// happen - namely, in BeginBlock, after rewards are allocated...
	// Slashes which happened in the first block would have been before this
	// delegation existed, UNLESS they were slashes of a redelegation to this
	// validator which was itself slashed (from a fault committed by the
	// redelegation source validator) earlier in the same BeginBlock.
	startingHeight := startingInfo.height
	// Slashes this block happened after reward allocation, but we have to account
	// for them for the stake sanity check below.
	endingHeight := blockHeight
	if endingHeight > startingHeight {
		err := IterateValidatorSlashEventsBetween(distributionInfo, val.OperatorAddress, startingHeight, endingHeight,
			func(height uint64, event *ValidatorSlashEvent) (stop bool, err error) {
				endingPeriod := event.validatorPeriod
				if endingPeriod > startingPeriod {
					if periodRewards, _err := calculateDelegationRewardsBetween(distributionInfo, val, startingPeriod, endingPeriod, stake); _err != nil {
						return true, _err
					} else {
						rewards = rewards.Add(periodRewards...)
					}
					// Note: It is necessary to truncate so we don't allow withdrawing
					// more rewards than owed.
					stake = stake.MulTruncate(sdk.OneDec().Sub(event.fraction))
					startingPeriod = endingPeriod
				}
				return false, nil
			},
		)
		if err != nil {
			return nil, err
		}
	}

	// A total stake sanity check; Recalculated final stake should be less than or
	// equal to current stake here. We cannot use Equals because stake is truncated
	// when multiplied by slash fractions (see above). We could only use equals if
	// we had arbitrary-precision rationals.
	currentStake := val.TokensFromShares(del.Shares)

	if stake.GT(currentStake) {
		// AccountI for rounding inconsistencies between:
		//
		//     currentStake: calculated as in staking with a single computation
		//     stake:        calculated as an accumulation of stake
		//                   calculations across validator's distribution periods
		//
		// These inconsistencies are due to differing order of operations which
		// will inevitably have different accumulated rounding and may lead to
		// the smallest decimal place being one greater in stake than
		// currentStake. When we calculated slashing by period, even if we
		// round down for each slash fraction, it's possible due to how much is
		// being rounded that we slash less when slashing by period insteafd of
		// for when we slash without periods. In other words, the single slash,
		// and the slashing by period could both be rounding down but the
		// slashing by period is simply rounding down less, thus making stake >
		// currentStake
		//
		// A small amount of this error is tolerated and corrected for,
		// however any greater amount should be considered a breach in expected
		// behaviour.
		marginOfErr := sdk.SmallestDec().MulInt64(3)
		if stake.LTE(currentStake.Add(marginOfErr)) {
			stake = currentStake
		} else {
			return sdk.DecCoins{}, fmt.Errorf("calculated final stake for delegator %s greater than current stake"+
				"\n\tfinal stake:\t%s"+
				"\n\tcurrent stake:\t%s",
				del.DelegatorAddress, stake, currentStake)
		}
	}

	// calculate rewards for final period
	if periodRewards, _err := calculateDelegationRewardsBetween(distributionInfo, val, startingPeriod, endingPeriod, stake); _err != nil {
		return periodRewards, _err
	} else {
		rewards = rewards.Add(periodRewards...)
	}

	return rewards, nil
}

// get the delegator withdraw address, defaulting to the delegator address
func (d DistributionInfo) GetDelegatorWithdrawAddr(delAddr string) string {
	b, exists := d.DelegatorWithdrawInfos.Get(delAddr)
	if !exists {
		return delAddr
	}
	return b
}

func withdrawDelegationRewards(logger log.Logger, genesisData *GenesisData, validatorOperatorAddress string, delegatorAddress string, cudosCfg *CudosMergeConfig, manifest *UpgradeManifest) (sdk.Coins, error) {

	// check existence of delegator starting info
	genesisData.DistributionInfo.DelegatorStartingInfos.Has(validatorOperatorAddress)
	StartingInfoMap, exists := genesisData.DistributionInfo.DelegatorStartingInfos.Get(validatorOperatorAddress)
	if !exists || !StartingInfoMap.Has(delegatorAddress) {
		return nil, fmt.Errorf("delegator starting info not found")
	}

	delegatorRewards := genesisData.DistributionInfo.Rewards.MustGet(delegatorAddress)
	rewardsRaw := delegatorRewards.MustGet(validatorOperatorAddress)
	outstanding := genesisData.DistributionInfo.OutstandingRewards.MustGet(validatorOperatorAddress)

	// defensive edge case may happen on the very final digits
	// of the decCoins due to operation order of the distribution mechanism.
	rewards := rewardsRaw.Intersect(outstanding)
	if !rewards.IsEqual(rewardsRaw) {
		if logger != nil {
			logger.Error(
				"rounding error withdrawing rewards from validator",
				"delegator", delegatorAddress,
				"validator", validatorOperatorAddress,
				"got", rewards.String(),
				"expected", rewardsRaw.String(),
			)
		}
	}

	// truncate reward dec coins, return remainder to community pool
	finalRewards, remainder := rewards.TruncateDecimal()

	// add coins to user account
	if !finalRewards.IsZero() {
		withdrawAddr := genesisData.DistributionInfo.GetDelegatorWithdrawAddr(delegatorAddress)

		// SendCoinsFromModuleToAccount
		err := moveGenesisBalance(genesisData, genesisData.DistributionInfo.DistributionModuleAccountAddress, withdrawAddr, finalRewards, "delegation_reward", manifest, cudosCfg)
		if err != nil {
			return nil, err
		}
	}

	// update the outstanding rewards and the community pool only if the
	// transaction was successful

	genesisData.DistributionInfo.OutstandingRewards.Set(validatorOperatorAddress, outstanding.Sub(rewards))
	genesisData.DistributionInfo.FeePool.CommunityPool = genesisData.DistributionInfo.FeePool.CommunityPool.Add(remainder...)

	// decrement reference count of starting period
	//startingInfo := k.GetDelegatorStartingInfo(ctx, del.GetValidatorAddr(), del.GetDelegatorAddr())
	//startingPeriod := startingInfo.PreviousPeriod
	//k.decrementReferenceCount(ctx, del.GetValidatorAddr(), startingPeriod)

	// remove delegator starting info
	//k.DeleteDelegatorStartingInfo(ctx, del.GetValidatorAddr(), del.GetDelegatorAddr())

	if finalRewards.IsZero() {
		/*
			baseDenom, _ := sdk.GetBaseDenom()
			if baseDenom == "" {
				baseDenom = cudosCfg.Config.OriginalDenom
			}
		*/
		baseDenom := genesisData.BondDenom

		// Note, we do not call the NewCoins constructor as we do not want the zero
		// coin removed.
		finalRewards = sdk.Coins{sdk.NewCoin(baseDenom, sdk.ZeroInt())}
	}

	/*
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventTypeWithdrawRewards,
				sdk.NewAttribute(sdk.AttributeKeyAmount, finalRewards.String()),
				sdk.NewAttribute(types.AttributeKeyValidator, val.GetOperator().String()),
			),
		)
	*/

	return finalRewards, nil
}

// Code based on IncrementValidatorPeriod
func UpdateValidatorData(distributionInfo *DistributionInfo, val *ValidatorInfo) uint64 {
	// fetch current rewards
	rewards := distributionInfo.ValidatorCurrentRewards.MustGet(val.OperatorAddress)

	// calculate current ratio
	var current sdk.DecCoins
	if val.Stake.IsZero() {

		// can't calculate ratio for zero-token validators
		// ergo we instead add to the community pool

		outstanding := distributionInfo.OutstandingRewards.MustGet(val.OperatorAddress)
		distributionInfo.FeePool.CommunityPool = distributionInfo.FeePool.CommunityPool.Add(rewards.reward...)
		outstanding = outstanding.Sub(rewards.reward)
		distributionInfo.OutstandingRewards.Set(val.OperatorAddress, outstanding)

		current = sdk.DecCoins{}
	} else {
		// note: necessary to truncate so we don't allow withdrawing more rewards than owed
		current = rewards.reward.QuoDecTruncate(val.Stake.ToDec())
	}

	// fetch historical rewards for last period
	//historical := k.GetValidatorHistoricalRewards(ctx, val.GetOperator(), rewards.Period-1).CumulativeRewardRatio
	historicalValInfo := distributionInfo.ValidatorHistoricalRewards.MustGet(val.OperatorAddress)
	historical := historicalValInfo.MustGet(rewards.period - 1)

	// decrement reference count
	//k.decrementReferenceCount(ctx, val.GetOperator(), rewards.Period-1)

	// set new historical rewards with reference count of 1
	// k.SetValidatorHistoricalRewards(ctx, val.GetOperator(), rewards.Period, types.NewValidatorHistoricalRewards(historical.Add(current...), 1))
	newCumulativeRatio := historical.cumulativeRewardRatio.Add(current...)
	historicalValInfo.Set(rewards.period, &ValidatorHistoricalReward{cumulativeRewardRatio: newCumulativeRatio})

	// set current rewards, incrementing period by 1
	//k.SetValidatorCurrentRewards(ctx, val.GetOperator(), types.NewValidatorCurrentRewards(sdk.DecCoins{}, rewards.Period+1))

	return rewards.period
}
