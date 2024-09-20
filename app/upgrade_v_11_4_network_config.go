package app

import sdk "github.com/cosmos/cosmos-sdk/types"

var (
	acudosToafetExchangeRateReducedByCommission, _ = sdk.NewDecFromStr("118.344")
	maxToleratedRemainingDistributionBalance, _    = sdk.NewIntFromString("1000000000000000000")
	maxToleratedRemainingStakingBalance, _         = sdk.NewIntFromString("100000000")
	maxToleratedRemainingMintBalance, _            = sdk.NewIntFromString("100000000")
	totalCudosSupply, _                            = sdk.NewIntFromString("10000000000000000000000000000")
	totalFetchSupplyToMint, _                      = sdk.NewIntFromString("88946755672000000000000000")

	totalCudosTestnetSupply, _       = sdk.NewIntFromString("20845618401448224096752009000")
	totalFetchTestnetSupplyToMint, _ = sdk.NewIntFromString("185415012678536211688587264")
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
			Almanac: &ProdDevContract{
				ProdAddr: "fetch1mezzhfj7qgveewzwzdk6lz5sae4dunpmmsjr9u7z0tpmdsae8zmquq3y0y",
			},
			AName: &ProdDevContract{
				ProdAddr: "fetch1479lwv5vy8skute5cycuz727e55spkhxut0valrcm38x9caa2x8q99ef0q",
			},
			TokenBridge: &TokenBridge{
				Addr:     "fetch1qxxlalvsdjd07p07y3rc5fu6ll8k4tmetpha8n",
				NewAdmin: getStringPtr("fetch15p3rl5aavw9rtu86tna5lgxfkz67zzr6ed4yhw"),
			},
		},
		CudosMerge: &CudosMergeConfig{
			ibcTargetAddr:                    "cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv", // Replace!!
			remainingStakingBalanceAddr:      "cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv", // Replace!!
			remainingGravityBalanceAddr:      "cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv", // Replace!!
			remainingDistributionBalanceAddr: "cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv", // Replace!!
			commissionFetchAddr:              "fetch122j02czdt5ca8cf576wy2hassyxyx67wg5xmgc", // Replace!!
			extraSupplyFetchAddr:             "fetch122j02czdt5ca8cf576wy2hassyxyx67wg5xmgc", // Reokace!!
			vestingCollisionDestAddr:         "fetch122j02czdt5ca8cf576wy2hassyxyx67wg5xmgc", // Replace!!
			communityPoolBalanceDestAddr:     "cudos1nj49l56x7sss5hqyvfmctxr3mq64whg273g3x5",

			newAddrPrefix: "fetch",
			oldAddrPrefix: "cudos",

			originalDenom:  "acudos",
			convertedDenom: "afet",
			stakingDenom:   "afet",

			vestingPeriod: 3 * 30 * 24 * 60 * 60, // 3 months period

			balanceConversionConstants: map[string]sdk.Dec{
				"acudos": acudosToafetExchangeRateReducedByCommission},

			totalCudosSupply:       totalCudosSupply,
			totalFetchSupplyToMint: totalFetchSupplyToMint,

			notVestedAccounts: map[string]bool{
				"cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv": true,
			},

			notDelegatedAccounts: map[string]bool{
				"cudos1qx3yaanre054nlq84qdzufsjmrrxcqxwzdkh6c": true,
			},

			MovedAccounts: map[string]string{
				"cudos1h6r6g0pwq7kcys5jcvfm9r7gcj3n2753hvk2ym": "cudos1w63ph9e4l07vpx7xdnje43cr2tlnr4jsfm4mvq",
				"cudos1jxyc7lny4q7te6sj5xyt9j86kyz82vlfdprl4a": "cudos1tfmkdzx9hm8g28vpgc3xhhxjjn460wzkwtayxr",
			},

			backupValidators: []string{"fetchvaloper122j02czdt5ca8cf576wy2hassyxyx67wdsecml"},
			validatorsMap: map[string]string{
				"cudosvaloper1s5qa3dpghnre6dqfgfhudxqjhwsv0mx43xayku": "fetchvaloper122j02czdt5ca8cf576wy2hassyxyx67wdsecml",
				"cudosvaloper1ctcrpuyumt60733u0yd5htwzedgfae0n8gql5n": "fetchvaloper122j02czdt5ca8cf576wy2hassyxyx67wdsecml"},
		},
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
			Almanac: &ProdDevContract{
				ProdAddr: "fetch1tjagw8g8nn4cwuw00cf0m5tl4l6wfw9c0ue507fhx9e3yrsck8zs0l3q4w",
				DevAddr:  "fetch135h26ys2nwqealykzey532gamw4l4s07aewpwc0cyd8z6m92vyhsplf0vp",
			},
			AName: &ProdDevContract{
				ProdAddr: "fetch1mxz8kn3l5ksaftx8a9pj9a6prpzk2uhxnqdkwuqvuh37tw80xu6qges77l",
				DevAddr:  "fetch1kewgfwxwtuxcnppr547wj6sd0e5fkckyp48dazsh89hll59epgpspmh0tn",
			},
		},
		CudosMerge: &CudosMergeConfig{
			ibcTargetAddr:                    "cudos1qqqd8x95ectdhwujwkq2kq6y09qgeal7t67kyz", // Replace!!
			remainingStakingBalanceAddr:      "cudos1qqqd8x95ectdhwujwkq2kq6y09qgeal7t67kyz", // Replace!!
			remainingGravityBalanceAddr:      "cudos1qqqd8x95ectdhwujwkq2kq6y09qgeal7t67kyz", // Replace!!
			remainingDistributionBalanceAddr: "cudos1qqqd8x95ectdhwujwkq2kq6y09qgeal7t67kyz", // Replace!!
			commissionFetchAddr:              "fetch122j02czdt5ca8cf576wy2hassyxyx67wg5xmgc", // Replace!!
			extraSupplyFetchAddr:             "fetch122j02czdt5ca8cf576wy2hassyxyx67wg5xmgc", // Reokace!!
			vestingCollisionDestAddr:         "fetch122j02czdt5ca8cf576wy2hassyxyx67wg5xmgc", // Replace!!
			communityPoolBalanceDestAddr:     "cudos1nj49l56x7sss5hqyvfmctxr3mq64whg273g3x5", // Replace!!

			newAddrPrefix: "fetch",
			oldAddrPrefix: "cudos",

			originalDenom:  "acudos",
			convertedDenom: "atestfet",
			stakingDenom:   "atestfet",

			vestingPeriod: 3 * 30 * 24 * 60 * 60, // 3 months period

			balanceConversionConstants: map[string]sdk.Dec{
				"acudos": acudosToafetExchangeRateReducedByCommission},

			totalCudosSupply:       totalCudosTestnetSupply,
			totalFetchSupplyToMint: totalFetchTestnetSupplyToMint,

			notVestedAccounts: map[string]bool{
				"cudos1qqqd8x95ectdhwujwkq2kq6y09qgeal7t67kyz": true,
			},

			notDelegatedAccounts: map[string]bool{
				"cudos1qqqd8x95ectdhwujwkq2kq6y09qgeal7t67kyz": true,
			},

			backupValidators: []string{"fetchvaloper122j02czdt5ca8cf576wy2hassyxyx67wdsecml"},
			validatorsMap: map[string]string{
				"cudosvaloper1s5qa3dpghnre6dqfgfhudxqjhwsv0mx43xayku": "fetchvaloper122j02czdt5ca8cf576wy2hassyxyx67wdsecml",
				"cudosvaloper1ctcrpuyumt60733u0yd5htwzedgfae0n8gql5n": "fetchvaloper122j02czdt5ca8cf576wy2hassyxyx67wdsecml"},
		},
	},
}

type NetworkConfig struct {
	ReconciliationInfo *ReconciliationInfo
	Contracts          *ContractSet
	CudosMerge         *CudosMergeConfig
}

type CudosMergeConfig struct {
	ibcTargetAddr                    string // Cudos address
	remainingStakingBalanceAddr      string // Cudos account for remaining bonded and not-bonded pool balances
	remainingGravityBalanceAddr      string // Cudos address
	remainingDistributionBalanceAddr string // Cudos address
	commissionFetchAddr              string // Fetch address for comission
	extraSupplyFetchAddr             string // Fetch address for extra supply
	vestingCollisionDestAddr         string // This gets converted to raw address, so it can be fetch or cudos address
	communityPoolBalanceDestAddr     string // This gets converted to raw address, so it can be fetch or cudos address

	newAddrPrefix string
	oldAddrPrefix string

	originalDenom  string
	convertedDenom string
	stakingDenom   string

	vestingPeriod int64 // Vesting period

	balanceConversionConstants map[string]sdk.Dec
	totalCudosSupply           sdk.Int
	totalFetchSupplyToMint     sdk.Int

	notVestedAccounts    map[string]bool
	notDelegatedAccounts map[string]bool
	MovedAccounts        map[string]string

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
	Almanac        *ProdDevContract
	AName          *ProdDevContract
}

type IContractBase interface {
	GetPrimaryContractAddr() *string
	GetContracts(contracts []string) []string
}

type IContractAdmin interface {
	IContractBase
	GetNewAdminAddr() *string
}

type IContractLabel interface {
	IContractBase
	GetNewLabel() *string
}

type IContractVersion interface {
	IContractBase
	GetNewVersion() *ContractVersion
}

type TokenBridge struct {
	Addr     string
	NewAdmin *string
}

func (c *TokenBridge) GetPrimaryContractAddr() *string {
	// NOTE(pb): This is a bit unorthodox approach allowing to call method for null pointer struct:
	if c == nil {
		return nil
	}
	return &c.Addr
}

func (c *TokenBridge) GetContracts(contracts []string) []string {
	// NOTE(pb): This is a bit unorthodox approach allowing to call method for null pointer struct:
	if c == nil {
		return contracts
	}

	if c.Addr != "" {
		contracts = append(contracts, c.Addr)
	}

	return contracts
}

func (c *TokenBridge) GetNewAdminAddr() *string {
	// NOTE(pb): This is a bit unorthodox approach allowing to call method for null pointer struct:
	if c == nil {
		return nil
	}

	return c.NewAdmin
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

func (c *Reconciliation) GetPrimaryContractAddr() *string {
	// NOTE(pb): This is a bit unorthodox approach allowing to call method for null pointer struct:
	if c == nil {
		return nil
	}
	return &c.Addr
}

func (c *Reconciliation) GetContracts(contracts []string) []string {
	// NOTE(pb): This is a bit unorthodox approach allowing to call method for null pointer struct:
	if c == nil {
		return contracts
	}

	if c.Addr != "" {
		contracts = append(contracts, c.Addr)
	}

	return contracts
}

func (c *Reconciliation) GetNewAdminAddr() *string {
	// NOTE(pb): This is a bit unorthodox approach allowing to call method for null pointer struct:
	if c == nil {
		return nil
	}

	return c.NewAdmin
}

func (c *Reconciliation) GetNewLabel() *string {
	// NOTE(pb): This is a bit unorthodox approach allowing to call method for null pointer struct:
	if c == nil {
		return nil
	}

	return c.NewLabel
}

func (c *Reconciliation) GetNewVersion() *ContractVersion {
	// NOTE(pb): This is a bit unorthodox approach allowing to call method for null pointer struct:
	if c == nil {
		return nil
	}

	return c.NewContractVersion
}

type ProdDevContract struct {
	DevAddr  string
	ProdAddr string
}

func (c *ProdDevContract) GetContracts(contracts []string) []string {
	// NOTE(pb): This is a bit unorthodox approach allowing to call method for null pointer struct:
	if c == nil {
		return contracts
	}

	if c.DevAddr != "" {
		contracts = append(contracts, c.DevAddr)
	}
	if c.ProdAddr != "" {
		contracts = append(contracts, c.ProdAddr)
	}

	return contracts
}
