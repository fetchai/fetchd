package app

var NetworkInfos = map[string]NetworkConfig{
	"fetchhub-4": {
		ibcTargetAddr:        "cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv", // Replace!!
		remainingBalanceAddr: "cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv", // Replace!!
		newAddrPrefix:        "fetch",
		oldAddrPrefix:        "cudos",

		originalDenom:  "acudos",
		convertedDenom: "afet",

		mergeTime:     123456,                // Epoch time of merge
		vestingPeriod: 3 * 30 * 24 * 60 * 60, // 3 months period

		balanceConversionConstants: map[string]int{
			"acudos": 11},

		notVestedAccounts: map[string]bool{
			"cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv": true,
		},

		backupValidators: []string{"F9F8271C4C395A557DF85F460793278BB45E63D8"},
		validatorsMap: map[string]string{
			"2FB481C55D2B93F7AC832A4423E47A5569FF23DF": "F9F8271C4C395A557DF85F460793278BB45E63D8"},
	},

	"dorado-1": {
		ibcTargetAddr: "cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv", // Replace!!
	},
}

type NetworkConfig struct {
	ibcTargetAddr        string
	remainingBalanceAddr string // Account for remaining bonded and not-bonded pool balances and balances from all other module accounts

	newAddrPrefix string
	oldAddrPrefix string

	originalDenom  string
	convertedDenom string

	mergeTime     int64 // Epoch time of merge - beginning of vesting period
	vestingPeriod int64 // Vesting period

	balanceConversionConstants map[string]int

	notVestedAccounts map[string]bool

	validatorsMap    map[string]string
	backupValidators []string
}
