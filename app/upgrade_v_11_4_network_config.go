package app

import sdk "github.com/cosmos/cosmos-sdk/types"

var (
	acudosToafet, _                             = sdk.NewDecFromStr("0.0909090909")
	commissionRate, _                           = sdk.NewDecFromStr("0.05")
	maxToleratedRemainingDistributionBalance, _ = sdk.NewIntFromString("1000000000000000000")
	maxToleratedRemainingStakingBalance, _      = sdk.NewIntFromString("100000000")
	maxToleratedRemainingMintBalance, _         = sdk.NewIntFromString("100000000")
)

var NetworkInfos = map[string]NetworkConfig{
	"fetchhub-4": {
		ReconciliationInfo: &ReconciliationInfo{
			TargetAddress:   "fetch1tynmzk68pq6kzawqffrqdhquq475gw9ccmlf9gk24mxjjy6ugl3q70aeyd",
			InputCSVRecords: readInputReconciliationData(reconciliationData),
		},
		Contracts: &ContractSet{
			Reconciliation: &Reconciliation{
				Addr:     "fetch1tynmzk68pq6kzawqffrqdhquq475gw9ccmlf9gk24mxjjy6ugl3q70aeyd",
				NewAdmin: getStringPtr("fetch15p3rl5aavw9rtu86tna5lgxfkz67zzr6ed4yhw"),
				NewLabel: getStringPtr("reconciliation-contract"),
				NewContractVersion: &ContractVersion{
					Contract: "contract-fetch-reconciliation",
					Version:  "1.0.0",
				},
			},
			Almanac: &Almanac{
				ProdAddr: "fetch1mezzhfj7qgveewzwzdk6lz5sae4dunpmmsjr9u7z0tpmdsae8zmquq3y0y",
			},
			AName: &AName{
				ProdAddr: "fetch1479lwv5vy8skute5cycuz727e55spkhxut0valrcm38x9caa2x8q99ef0q",
			},
			TokenBridge: &TokenBridge{
				Addr:     "fetch1qxxlalvsdjd07p07y3rc5fu6ll8k4tmetpha8n",
				NewAdmin: getStringPtr("fetch15p3rl5aavw9rtu86tna5lgxfkz67zzr6ed4yhw"),
			},
		},
		ibcTargetAddr:                    "cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv", // Replace!!
		remainingStakingBalanceAddr:      "cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv", // Replace!!
		remainingGravityBalanceAddr:      "cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv", // Replace!!
		remainingDistributionBalanceAddr: "cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv", // Replace!!
		commissionFetchAddr:              "fetch122j02czdt5ca8cf576wy2hassyxyx67wg5xmgc", // Replace!!

		newAddrPrefix: "fetch",
		oldAddrPrefix: "cudos",

		originalDenom:  "acudos",
		convertedDenom: "afet",
		stakingDenom:   "afet",

		mergeTime:     123456,                // Epoch time of merge
		vestingPeriod: 3 * 30 * 24 * 60 * 60, // 3 months period

		balanceConversionConstants: map[string]sdk.Dec{
			"acudos": acudosToafet},

		commissionRate: commissionRate,

		notVestedAccounts: map[string]bool{
			"cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv": true,
		},

		backupValidators: []string{"fetchvaloper122j02czdt5ca8cf576wy2hassyxyx67wdsecml"},
		validatorsMap: map[string]string{
			"cudosvaloper1s5qa3dpghnre6dqfgfhudxqjhwsv0mx43xayku": "fetchvaloper122j02czdt5ca8cf576wy2hassyxyx67wdsecml",
			"cudosvaloper1ctcrpuyumt60733u0yd5htwzedgfae0n8gql5n": "fetchvaloper122j02czdt5ca8cf576wy2hassyxyx67wdsecml"},
	},

	"dorado-1": {
		ReconciliationInfo: &ReconciliationInfo{
			TargetAddress:   "fetch1g5ur2wc5xnlc7sw9wd895lw7mmxz04r5syj3s6ew8md6pvwuweqqavkgt0",
			InputCSVRecords: readInputReconciliationData(reconciliationDataTestnet),
		},
		Contracts: &ContractSet{
			Reconciliation: &Reconciliation{
				Addr: "fetch1g5ur2wc5xnlc7sw9wd895lw7mmxz04r5syj3s6ew8md6pvwuweqqavkgt0",
				NewContractVersion: &ContractVersion{
					Contract: "contract-fetch-reconciliation",
					Version:  "1.0.0",
				},
			},
			Almanac: &Almanac{
				ProdAddr: "fetch1tjagw8g8nn4cwuw00cf0m5tl4l6wfw9c0ue507fhx9e3yrsck8zs0l3q4w",
				DevAddr:  "fetch135h26ys2nwqealykzey532gamw4l4s07aewpwc0cyd8z6m92vyhsplf0vp",
			},
			AName: &AName{
				ProdAddr: "fetch1mxz8kn3l5ksaftx8a9pj9a6prpzk2uhxnqdkwuqvuh37tw80xu6qges77l",
				DevAddr:  "fetch1kewgfwxwtuxcnppr547wj6sd0e5fkckyp48dazsh89hll59epgpspmh0tn",
			},
		},
		ibcTargetAddr: "fetchvaloper14w6a4al72uc3fpfy4lqtg0a7xtkx3w7hda0vel", // Replace!!
	},
}

type NetworkConfig struct {
	ReconciliationInfo               *ReconciliationInfo
	Contracts                        *ContractSet
	ibcTargetAddr                    string
	remainingStakingBalanceAddr      string // Account for remaining bonded and not-bonded pool balances and balances from all other module accounts
	remainingGravityBalanceAddr      string // Account for remaining bonded and not-bonded pool balances and balances from all other module accounts
	remainingDistributionBalanceAddr string // Account for remaining bonded and not-bonded pool balances and balances from all other module accounts
	commissionFetchAddr              string

	newAddrPrefix string
	oldAddrPrefix string

	originalDenom  string
	convertedDenom string
	stakingDenom   string

	mergeTime     int64 // Epoch time of merge - beginning of vesting period
	vestingPeriod int64 // Vesting period

	balanceConversionConstants map[string]sdk.Dec
	commissionRate             sdk.Dec

	notVestedAccounts map[string]bool

	validatorsMap    map[string]string
	backupValidators []string
}

type ReconciliationInfo struct {
	TargetAddress   string
	InputCSVRecords [][]string
}

type ContractSet struct {
	Reconciliation *Reconciliation
	TokenBridge    *TokenBridge
	Almanac        *Almanac
	AName          *AName
}

type TokenBridge struct {
	Addr     string
	NewAdmin *string
}

type ContractVersion struct {
	Contract string `json:"contract"`
	Version  string `json:"version"`
}

type Reconciliation struct {
	Addr               string
	NewAdmin           *string
	NewLabel           *string
	NewContractVersion *ContractVersion
}

type Almanac struct {
	DevAddr  string
	ProdAddr string
}

type AName struct {
	DevAddr  string
	ProdAddr string
}
