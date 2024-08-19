package app

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
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

type DistributionInfo struct {
	// fee_pool
	// outstanding_rewards
	// params
	previousProposer string
	// validator_accumulated_comissions
	// validator_current_rewards
	// validator_historical_rewards
	delegatorStartingInfos OrderedMap[string, OrderedMap[string, DelegatorStartingInfo]] // validator_addr -> delegator_addr -> starting_info
	delegatorWithdrawInfos OrderedMap[string, string]                                    // delegator_address -> withdraw_address
	ValidatorSlashEvents   OrderedMap[string, OrderedMap[uint64, ValidatorSlashEvent]]   // validatior_address -> height -> validator_slash_event
}

func withdrawGenesisRewards(jsonData map[string]interface{}, genesisValidators *OrderedMap[string, ValidatorInfo], genesisAccounts *OrderedMap[string, AccountInfo], contractAccountMap *OrderedMap[string, ContractInfo], networkInfo NetworkConfig, manifest *UpgradeManifest) error {

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

	println(communityPoolBalance.Add(communityPoolBalance...).String())
	println(DistributionDecBalance.Sub(communityPoolBalance.Add(totalOutstandingRewards...)).String())

	return nil
}
