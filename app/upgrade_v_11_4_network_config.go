package app

var NetworkInfos = map[string]NetworkConfig{
	"fetchhub-4": {
		IbcTargetAddr: "fetch1zydegef0z6lz4gamamzlnu52ethe8xnm0xe5fkyrgwumsh9pplus5he63f",
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
					Contract: "contract-fetch-asi-reconciliation",
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
	},

	"dorado-1": {
		IbcTargetAddr: "fetch18rlg4hs2p03yuvvdu389pe65qa789asmyqsfftdxsh2qjfwmt94qmrf7g0",
		ReconciliationInfo: &ReconciliationInfo{
			TargetAddress:   "fetch1g5ur2wc5xnlc7sw9wd895lw7mmxz04r5syj3s6ew8md6pvwuweqqavkgt0",
			InputCSVRecords: readInputReconciliationData(reconciliationDataTestnet),
		},
		Contracts: &ContractSet{
			Reconciliation: &Reconciliation{
				Addr: "fetch1g5ur2wc5xnlc7sw9wd895lw7mmxz04r5syj3s6ew8md6pvwuweqqavkgt0",
				NewContractVersion: &ContractVersion{
					Contract: "contract-fetch-asi-reconciliation",
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
	},
}

type NetworkConfig struct {
	ReconciliationInfo *ReconciliationInfo
	Contracts          *ContractSet
	IbcTargetAddr      string
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
