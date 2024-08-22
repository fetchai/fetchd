package app

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/spf13/cast"
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
	communityPool sdk.DecCoins
}

type DistributionInfo struct {
	feePool            FeePool
	outstandingRewards OrderedMap[string, sdk.DecCoins]

	// params
	// previousProposer string

	validatorAccumulatedCommissions OrderedMap[string, sdk.DecCoins]                                  // validator_addr -> validator_accumulated_commissions
	validatorCurrentRewards         OrderedMap[string, ValidatorCurrentReward]                        // validator_addr -> validator_current_rewards
	validatorHistoricalRewards      OrderedMap[string, OrderedMap[uint64, ValidatorHistoricalReward]] // validator_addr -> period -> validator_historical_reward

	delegatorStartingInfos           OrderedMap[string, OrderedMap[string, DelegatorStartingInfo]] // validator_addr -> delegator_addr -> starting_info
	delegatorWithdrawInfos           OrderedMap[string, string]                                    // delegator_address -> withdraw_address
	validatorSlashEvents             OrderedMap[string, OrderedMap[uint64, ValidatorSlashEvent]]   // validatior_address -> height -> validator_slash_event
	distributionModuleAccountAddress string
}

func parseDelegatorStartingInfos(distribution map[string]interface{}) (*OrderedMap[string, OrderedMap[string, DelegatorStartingInfo]], error) {

	delegatorStartingInfos := NewOrderedMap[string, OrderedMap[string, DelegatorStartingInfo]]()
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

		if _, exists := delegatorStartingInfos.Get(validatorAddress); !exists {
			delegatorStartingInfos.Set(validatorAddress, *NewOrderedMap[string, DelegatorStartingInfo]())
		}
		valStartingInfo := delegatorStartingInfos.MustGet(validatorAddress)

		valStartingInfo.Set(delegatorAddress, delegatorStartingInfo)
		delegatorStartingInfos.Set(validatorAddress, *valStartingInfo)

	}
	return delegatorStartingInfos, nil
}

func parseValidatorHistoricalRewards(distribution map[string]interface{}) (*OrderedMap[string, OrderedMap[uint64, ValidatorHistoricalReward]], error) {

	validatorHistoricalRewards := NewOrderedMap[string, OrderedMap[uint64, ValidatorHistoricalReward]]()
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

		if _, exists := validatorHistoricalRewards.Get(validatorAddress); !exists {
			validatorHistoricalRewards.Set(validatorAddress, *NewOrderedMap[uint64, ValidatorHistoricalReward]())
		}
		valRewards := validatorHistoricalRewards.MustGet(validatorAddress)

		valRewards.SetNew(period, delegatorStartingInfo)
		//validatorHistoricalRewards.Set(validatorAddress, *valRewards)

	}
	return validatorHistoricalRewards, nil
}

func parseValidatorCurrentRewards(distribution map[string]interface{}) (*OrderedMap[string, ValidatorCurrentReward], error) {

	validatorCurrentRewards := NewOrderedMap[string, ValidatorCurrentReward]()
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

		validatorCurrentRewards.SetNew(validatorAddress, validatorCurrentReward)

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

func parseValidatorSlashEvents(distribution map[string]interface{}) (*OrderedMap[string, OrderedMap[uint64, ValidatorSlashEvent]], error) {

	validatorSlashEvents := NewOrderedMap[string, OrderedMap[uint64, ValidatorSlashEvent]]()
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

		if _, exists := validatorSlashEvents.Get(validatorAddress); !exists {
			validatorSlashEvents.Set(validatorAddress, *NewOrderedMap[uint64, ValidatorSlashEvent]())
		}
		valEvents := validatorSlashEvents.MustGet(validatorAddress)

		if delegatorStartingInfo.validatorPeriod != delegatorStartingInfo.period {
			panic("erorr")
		}

		valEvents.SetNew(height, delegatorStartingInfo)
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

	return &FeePool{communityPool: communityPool}, nil
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

func parseGenesisDistribution(jsonData map[string]interface{}, genesisAccounts *OrderedMap[string, AccountInfo]) (*DistributionInfo, error) {
	distribution := jsonData[distributiontypes.ModuleName].(map[string]interface{})
	distributionInfo := DistributionInfo{}

	delegatorStartingInfos, err := parseDelegatorStartingInfos(distribution)
	if err != nil {
		return nil, err
	}
	distributionInfo.delegatorStartingInfos = *delegatorStartingInfos

	validatorHistoricalRewards, err := parseValidatorHistoricalRewards(distribution)
	if err != nil {
		return nil, err
	}
	distributionInfo.validatorHistoricalRewards = *validatorHistoricalRewards

	validatorCurrentRewards, err := parseValidatorCurrentRewards(distribution)
	if err != nil {
		return nil, err
	}
	distributionInfo.validatorCurrentRewards = *validatorCurrentRewards

	validatorSlashEvents, err := parseValidatorSlashEvents(distribution)
	if err != nil {
		return nil, err
	}
	distributionInfo.validatorSlashEvents = *validatorSlashEvents

	outstandingRewards, err := parseOutstandingRewards(distribution)
	if err != nil {
		return nil, err
	}
	distributionInfo.outstandingRewards = *outstandingRewards

	feePool, err := parseFeePool(distribution)
	if err != nil {
		return nil, err
	}
	distributionInfo.feePool = *feePool

	delegatorWithdrawInfos, err := parseDelegatorWithdrawInfos(distribution)
	if err != nil {
		return nil, err
	}
	distributionInfo.delegatorWithdrawInfos = *delegatorWithdrawInfos

	validatorAccumulatedCommissions, err := parseValidatorAccumulatedCommissions(distribution)
	if err != nil {
		return nil, err
	}
	distributionInfo.validatorAccumulatedCommissions = *validatorAccumulatedCommissions

	distributionModuleAddress, err := GetAddressByName(genesisAccounts, DistributionAccName)
	if err != nil {
		return nil, err
	}
	distributionInfo.distributionModuleAccountAddress = distributionModuleAddress

	return &distributionInfo, nil
}

func withdrawGenesisDistributionRewards(genesisData *GenesisData, networkInfo NetworkConfig, manifest *UpgradeManifest) error {
	// We need to get this somehow
	blockHeight := uint64(11376855)

	// Withdraw all delegation rewards
	for _, validatorOpertorAddr := range *genesisData.distributionInfo.delegatorStartingInfos.Keys() {
		validator := genesisData.validators.MustGet(validatorOpertorAddr)
		delegatorStartInfo := genesisData.distributionInfo.delegatorStartingInfos.MustGet(validatorOpertorAddr)

		endingPeriod := updateValidatorData(&genesisData.distributionInfo, *validator)

		for _, delegatorAddr := range *delegatorStartInfo.Keys() {
			delegation := validator.delegations.MustGet(delegatorAddr)

			_, err := withdrawDelegationRewards(genesisData, *validator, *delegation, endingPeriod, blockHeight, networkInfo, manifest)
			if err != nil {
				return err
			}
		}

	}

	// Withdraw validator accumulated comission
	err := withdrawAccumulatedCommissions(genesisData, networkInfo, manifest)
	if err != nil {
		return err
	}

	// Withdraw remaining balance
	distributionModuleAccount := genesisData.accounts.MustGet(genesisData.distributionInfo.distributionModuleAccountAddress)
	println("Remaining dist balance: ", distributionModuleAccount.balance.String())

	err = MoveGenesisBalance(genesisData, genesisData.distributionInfo.distributionModuleAccountAddress, networkInfo.remainingDistributionBalanceAddr, distributionModuleAccount.balance, manifest)
	if err != nil {
		return err
	}

	/*
		totalOutstandingRewards := sdk.DecCoins{}
		for _, validatorAddr := range *genesisData.distributionInfo.outstandingRewards.Keys() {
			validatorReward := genesisData.distributionInfo.outstandingRewards.MustGet(validatorAddr)

			//validatorAddress := outstandingRewardMap["validator_address"].(string)

			totalOutstandingRewards = totalOutstandingRewards.Add(*validatorReward...)
		}
		println("Total outstanding rewards ", totalOutstandingRewards.String())
	*/
	/*

		/*
			distribution := jsonData[distributiontypes.ModuleName].(map[string]interface{})
			communityPoolBalance, err := getDecCoinsFromInterfaceSlice(distribution["fee_pool"].(map[string]interface{})["community_pool"].([]interface{}))
			if err != nil {
				return err
			}
			println("Community ", communityPoolBalance.String())

			DistributionModuleAddress, err := GetAddressByName(genesisAccounts, DistributionAccName)
			if err != nil {
				return err
			}
			DistributionAcc, _ := genesisAccounts.Get(DistributionModuleAddress)
			DistributionDecBalance := sdk.NewDecCoinsFromCoins(DistributionAcc.balance...)
			println("Distribution ", DistributionDecBalance.String())

			totalOutstandingRewards := sdk.NewDecCoins()
			outstandingRewards := distribution["outstanding_rewards"].([]interface{})
			for _, outstandingReward := range outstandingRewards {
				outstandingRewardMap := outstandingReward.(map[string]interface{})

				//validatorAddress := outstandingRewardMap["validator_address"].(string)

				outstandingRewardBalance, err := getDecCoinsFromInterfaceSlice(outstandingRewardMap["outstanding_rewards"].([]interface{}))
				if err != nil {
					return err
				}

				totalOutstandingRewards = totalOutstandingRewards.Add(outstandingRewardBalance...)
			}
			println("Total outstanding rewards ", totalOutstandingRewards.String())

			totalcurrentRewards := sdk.NewDecCoins()
			currentRewards := distribution["validator_current_rewards"].([]interface{})
			for _, currentReward := range currentRewards {
				currentRewardMap := currentReward.(map[string]interface{})

				//validatorAddress := currentRewardMap["validator_address"].(string)

				rewardBalance, err := getDecCoinsFromInterfaceSlice(currentRewardMap["rewards"].(map[string]interface{})["rewards"].([]interface{}))
				if err != nil {
					return err
				}

				totalcurrentRewards = totalcurrentRewards.Add(rewardBalance...)

			}
			println("Total current rewards ", totalcurrentRewards.String())

			totalAccumulatedComissions := sdk.NewDecCoins()
			ValidatorAccumulatedCommissions := distribution["validator_accumulated_commissions"].([]interface{})
			for _, validatorAccumulatedCommission := range ValidatorAccumulatedCommissions {
				validatorAccumulatedCommissionMap := validatorAccumulatedCommission.(map[string]interface{})

				//validatorAddress := validatorAccumulatedCommissionMap["validator_address"].(string)

				AccumulatedComissionsBalance, err := getDecCoinsFromInterfaceSlice(validatorAccumulatedCommissionMap["accumulated"].(map[string]interface{})["commission"].([]interface{}))
				if err != nil {
					return err
				}

				totalAccumulatedComissions = totalAccumulatedComissions.Add(AccumulatedComissionsBalance...)

				//println(validatorAddress, AccumulatedComissionsBalance)
			}

			println("Total accumulated comission ", totalAccumulatedComissions.String())

			totalStartingStake := sdk.NewDec(0)
			delegatorStartingInfos := distribution["delegator_starting_infos"].([]interface{})
			for _, delegatorStartingInfo := range delegatorStartingInfos {
				delegatorStartingInfoMap := delegatorStartingInfo.(map[string]interface{})

				//validatorAddress := validatorAccumulatedCommissionMap["validator_address"].(string)

				delegatorStartingInfoStake, err := sdk.NewDecFromStr(delegatorStartingInfoMap["starting_info"].(map[string]interface{})["stake"].(string))
				if err != nil {
					return err
				}

				totalStartingStake = totalStartingStake.Add(delegatorStartingInfoStake)

				//println(validatorAddress, AccumulatedComissionsBalance)
			}
			println("Total starting info stake ", totalStartingStake.String())

			println(communityPoolBalance.Add(communityPoolBalance...).String())
			println(DistributionDecBalance.Sub(communityPoolBalance.Add(totalOutstandingRewards...)).String())
	*/

	return nil
}

func withdrawAccumulatedCommissions(genesisData *GenesisData, networkInfo NetworkConfig, manifest *UpgradeManifest) error {

	for _, validatorAddress := range *genesisData.distributionInfo.validatorAccumulatedCommissions.Keys() {
		accumulatedCommission := genesisData.distributionInfo.validatorAccumulatedCommissions.MustGet(validatorAddress)

		accountAddress, err := convertAddressPrefix(validatorAddress, networkInfo.oldAddrPrefix)
		if err != nil {
			return err
		}

		finalRewards, _ := accumulatedCommission.TruncateDecimal()

		err = MoveGenesisBalance(genesisData, genesisData.distributionInfo.distributionModuleAccountAddress, accountAddress, finalRewards, manifest)
		if err != nil {
			return err
		}
	}

	return nil
}

// calculate the rewards accrued by a delegation between two periods
func calculateDelegationRewardsBetween(distributionInfo DistributionInfo, val ValidatorInfo,
	startingPeriod, endingPeriod uint64, stake sdk.Dec,
) (rewards sdk.DecCoins) {
	// sanity check
	if startingPeriod > endingPeriod {
		panic("startingPeriod cannot be greater than endingPeriod")
	}

	// sanity check
	if stake.IsNegative() {
		panic("stake should not be negative")
	}

	// return staking * (ending - starting)

	operatorRewards := distributionInfo.validatorHistoricalRewards.MustGet(val.operatorAddress)
	starting := operatorRewards.MustGet(startingPeriod)
	ending := operatorRewards.MustGet(endingPeriod)

	difference := ending.cumulativeRewardRatio.Sub(starting.cumulativeRewardRatio)
	if difference.IsAnyNegative() {
		panic("negative rewards should not be possible")
	}
	// note: necessary to truncate so we don't allow withdrawing more rewards than owed
	rewards = difference.MulDecTruncate(stake)
	return
}

// iterate over slash events between heights, inclusive
func IterateValidatorSlashEventsBetween(distributionInfo DistributionInfo, val string, startingHeight uint64, endingHeight uint64,
	handler func(height uint64, event ValidatorSlashEvent) (stop bool),
) {
	slashEvents, exists := distributionInfo.validatorSlashEvents.Get(val)
	// No slashing events
	if !exists {
		return
	}

	sortUint64Keys(slashEvents)

	keys := slashEvents.Keys()

	// Perform binary search to find the starting point
	startIdx := sort.Search(len(*keys), func(i int) bool {
		return (*keys)[i] >= startingHeight
	})

	// Iterate from the startIdx up to the ending height
	for i := startIdx; i < len(*keys); i++ {
		height := (*keys)[i]
		if height > endingHeight {
			break
		}

		event := slashEvents.MustGet(height)
		if handler(height, *event) {
			break
		}
	}
}

// calculate the total rewards accrued by a delegation
func CalculateDelegationRewards(blockHeight uint64, distributionInfo DistributionInfo, val ValidatorInfo, del DelegationInfo, endingPeriod uint64) (rewards sdk.DecCoins) {
	// fetch starting info for delegation
	delStartingInfo := distributionInfo.delegatorStartingInfos.MustGet(val.operatorAddress)
	startingInfo := delStartingInfo.MustGet(del.delegatorAddress)

	if startingInfo.height == blockHeight {
		// started this height, no rewards yet
		return
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
		IterateValidatorSlashEventsBetween(distributionInfo, val.operatorAddress, startingHeight, endingHeight,
			func(height uint64, event ValidatorSlashEvent) (stop bool) {
				endingPeriod := event.validatorPeriod
				if endingPeriod > startingPeriod {
					rewards = rewards.Add(calculateDelegationRewardsBetween(distributionInfo, val, startingPeriod, endingPeriod, stake)...)

					// Note: It is necessary to truncate so we don't allow withdrawing
					// more rewards than owed.
					stake = stake.MulTruncate(sdk.OneDec().Sub(event.fraction))
					startingPeriod = endingPeriod
				}
				return false
			},
		)
	}

	// A total stake sanity check; Recalculated final stake should be less than or
	// equal to current stake here. We cannot use Equals because stake is truncated
	// when multiplied by slash fractions (see above). We could only use equals if
	// we had arbitrary-precision rationals.
	currentStake := val.TokensFromShares(del.shares)

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
			panic(fmt.Sprintf("calculated final stake for delegator %s greater than current stake"+
				"\n\tfinal stake:\t%s"+
				"\n\tcurrent stake:\t%s",
				del.delegatorAddress, stake, currentStake))
		}
	}

	// calculate rewards for final period
	rewards = rewards.Add(calculateDelegationRewardsBetween(distributionInfo, val, startingPeriod, endingPeriod, stake)...)
	return rewards
}

// get the delegator withdraw address, defaulting to the delegator address
func (d DistributionInfo) GetDelegatorWithdrawAddr(delAddr string) string {
	b, exists := d.delegatorWithdrawInfos.Get(delAddr)
	if !exists {
		return delAddr
	}
	return *b
}

func withdrawDelegationRewards(genesisData *GenesisData, val ValidatorInfo, del DelegationInfo, endingPeriod uint64, blockHeight uint64, networkInfo NetworkConfig, manifest *UpgradeManifest) (sdk.Coins, error) {
	// check existence of delegator starting info
	genesisData.distributionInfo.delegatorStartingInfos.Has(val.operatorAddress)
	StartingInfoMap, exists := genesisData.distributionInfo.delegatorStartingInfos.Get(val.operatorAddress)
	if !exists || !StartingInfoMap.Has(del.delegatorAddress) {
		return nil, fmt.Errorf("delegator starting info not found")
	}

	// end current period and calculate rewards
	//endingPeriod := k.IncrementValidatorPeriod(ctx, val)
	rewardsRaw := CalculateDelegationRewards(blockHeight, genesisData.distributionInfo, val, del, endingPeriod)
	outstanding := genesisData.distributionInfo.outstandingRewards.MustGet(val.operatorAddress)

	// defensive edge case may happen on the very final digits
	// of the decCoins due to operation order of the distribution mechanism.
	rewards := rewardsRaw.Intersect(*outstanding)
	if !rewards.IsEqual(rewardsRaw) {
		println(
			"rounding error withdrawing rewards from validator",
			"delegator", del.delegatorAddress,
			"validator", val.operatorAddress,
			"got", rewards.String(),
			"expected", rewardsRaw.String(),
		)
	}

	// truncate reward dec coins, return remainder to community pool
	finalRewards, remainder := rewards.TruncateDecimal()

	// add coins to user account
	if !finalRewards.IsZero() {
		withdrawAddr := genesisData.distributionInfo.GetDelegatorWithdrawAddr(del.delegatorAddress)

		// SendCoinsFromModuleToAccount
		err := MoveGenesisBalance(genesisData, genesisData.distributionInfo.distributionModuleAccountAddress, withdrawAddr, finalRewards, manifest)
		if err != nil {
			return nil, err
		}
	}

	// update the outstanding rewards and the community pool only if the
	// transaction was successful

	genesisData.distributionInfo.outstandingRewards.Set(val.operatorAddress, outstanding.Sub(rewards))
	genesisData.distributionInfo.feePool.communityPool = genesisData.distributionInfo.feePool.communityPool.Add(remainder...)

	// decrement reference count of starting period
	//startingInfo := k.GetDelegatorStartingInfo(ctx, del.GetValidatorAddr(), del.GetDelegatorAddr())
	//startingPeriod := startingInfo.PreviousPeriod
	//k.decrementReferenceCount(ctx, del.GetValidatorAddr(), startingPeriod)

	// remove delegator starting info
	//k.DeleteDelegatorStartingInfo(ctx, del.GetValidatorAddr(), del.GetDelegatorAddr())

	if finalRewards.IsZero() {
		baseDenom, _ := sdk.GetBaseDenom()
		if baseDenom == "" {
			baseDenom = networkInfo.originalDenom
		}

		// Note, we do not call the NewCoins constructor as we do not want the zero
		// coin removed.
		finalRewards = sdk.Coins{sdk.NewCoin(baseDenom, sdk.ZeroInt())}
	}

	// TODO: manifest event
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
func updateValidatorData(distributionInfo *DistributionInfo, val ValidatorInfo) uint64 {

	// fetch current rewards
	rewards := distributionInfo.validatorCurrentRewards.MustGet(val.operatorAddress)

	// calculate current ratio
	var current sdk.DecCoins
	if val.stake.IsZero() {

		// can't calculate ratio for zero-token validators
		// ergo we instead add to the community pool

		outstanding := *distributionInfo.outstandingRewards.MustGet(val.operatorAddress)
		distributionInfo.feePool.communityPool = distributionInfo.feePool.communityPool.Add(rewards.reward...)
		outstanding = outstanding.Sub(rewards.reward)
		distributionInfo.outstandingRewards.Set(val.operatorAddress, outstanding)

		current = sdk.DecCoins{}
	} else {
		// note: necessary to truncate so we don't allow withdrawing more rewards than owed
		current = rewards.reward.QuoDecTruncate(val.stake.ToDec())
	}

	// fetch historical rewards for last period
	//historical := k.GetValidatorHistoricalRewards(ctx, val.GetOperator(), rewards.Period-1).CumulativeRewardRatio
	historicalValInfo := distributionInfo.validatorHistoricalRewards.MustGet(val.operatorAddress)
	historical := historicalValInfo.MustGet(rewards.period - 1)

	// decrement reference count
	//k.decrementReferenceCount(ctx, val.GetOperator(), rewards.Period-1)

	// set new historical rewards with reference count of 1
	// k.SetValidatorHistoricalRewards(ctx, val.GetOperator(), rewards.Period, types.NewValidatorHistoricalRewards(historical.Add(current...), 1))
	historical.cumulativeRewardRatio = historical.cumulativeRewardRatio.Add(current...)
	historicalValInfo.Set(rewards.period, *historical)

	// set current rewards, incrementing period by 1
	//k.SetValidatorCurrentRewards(ctx, val.GetOperator(), types.NewValidatorCurrentRewards(sdk.DecCoins{}, rewards.Period+1))

	return rewards.period
}
