package app

var NetworkInfos = map[string]NetworkConfig{
	"fetchhub-4": {
		IbcTargetAddr:        "cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv", // Replace!!
		RemainingBalanceAddr: "cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv", // Replace!!
		NewAddrPrefix:        "fetch",
		OldAddrPrefix:        "cudos",

		OriginalDenom:  "acudos",
		ConvertedDenom: "afet",

		MergeTime:     123456,                // Epoch time of merge
		VestingPeriod: 3 * 30 * 24 * 60 * 60, // 3 months period

		BalanceConversionConstants: map[string]int{
			"acudos": 11},

		NotVestedAccounts: map[string]bool{
			"cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv": true,
		},
	},

	"dorado-1": {
		IbcTargetAddr: "cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv", // Replace!!
	},
}

type NetworkConfig struct {
	IbcTargetAddr        string
	RemainingBalanceAddr string // Account for remaining bonded and not-bonded pool balances and balances from all other module accounts

	NewAddrPrefix string
	OldAddrPrefix string

	OriginalDenom  string
	ConvertedDenom string

	MergeTime     int64 // Epoch time of merge - beginning of vesting period
	VestingPeriod int64 // Vesting period

	BalanceConversionConstants map[string]int

	NotVestedAccounts map[string]bool
}
