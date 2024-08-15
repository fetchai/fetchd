package app

var NetworkInfos = map[string]NetworkConfig{
	"fetchhub-4": {
		IbcTargetAddr: "cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv", // Replace!!
		NewAddrPrefix: "fetch",
		OldAddrPrefix: "cudos",

		OriginalDenom:  "acudos",
		ConvertedDenom: "afet",

		MergeTime:     123456,                // Epoch time of merge
		VestingPeriod: 3 * 30 * 24 * 60 * 60, // 3 months period

	},

	"dorado-1": {
		IbcTargetAddr: "cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv", // Replace!!
	},
}

type NetworkConfig struct {
	IbcTargetAddr string

	NewAddrPrefix string
	OldAddrPrefix string

	OriginalDenom  string
	ConvertedDenom string

	MergeTime     uint64 // Epoch time of merge
	VestingPeriod uint64 // 3 months period

}
