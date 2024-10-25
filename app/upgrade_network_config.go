package app

import (
	"encoding/json"
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"io/ioutil"
	"os"
)

var (
	DefaultMaxToleratedRemainingDistributionBalance, _ = sdk.NewIntFromString("1000000000000000000")
	DefaultMaxToleratedRemainingStakingBalance, _      = sdk.NewIntFromString("100000000")
	DefaultMaxToleratedRemainingMintBalance, _         = sdk.NewIntFromString("100000000")
)

func unwrapOrDefault[T any](ptr *T, defaultValue T) T {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

func newInt(val string) sdk.Int {
	res, ok := sdk.NewIntFromString(val)
	if !ok {
		panic(fmt.Errorf("Failed to parse INT %s", val))
	}
	return res
}

func newIntRef(val string) *sdk.Int {
	res, ok := sdk.NewIntFromString(val)
	if !ok {
		panic(fmt.Errorf("Failed to parse INT %s", val))
	}
	return &res
}

func newDec(val string) sdk.Dec {
	res, err := sdk.NewDecFromStr(val)
	if err != nil {
		panic(err)
	}
	return res
}

type BalanceMovement struct {
	SourceAddress      string   `json:"from"`
	DestinationAddress string   `json:"to"`
	Amount             *sdk.Int `json:"amount,omitempty"`
	Memo               string   `json:"memo,omitempty"`
}

var ReconciliationRecords = map[string]*[][]string{
	"fetchhub-4":            readInputReconciliationData(reconciliationData),
	"fetchhub-cudos-test-4": readInputReconciliationData(reconciliationData),
	"dorado-1":              readInputReconciliationData(reconciliationDataTestnet),
}

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
					CW2version: &CW2ContractVersion{
						Contract: "contract-fetch-reconciliation",
						Version:  "1.0.0",
					},
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
		CudosMerge: &CudosMergeConfigJSON{
			IbcTargetAddr:                    "cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv", // Replace!!
			RemainingStakingBalanceAddr:      "cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv", // Replace!!
			RemainingGravityBalanceAddr:      "cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv", // Replace!!
			RemainingDistributionBalanceAddr: "cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv", // Replace!!
			ContractDestinationFallbackAddr:  "cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv", // Replace!!
			GenericModuleRemainingBalance:    "cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv", // Replace!!

			CommissionFetchAddr:          "fetch122j02czdt5ca8cf576wy2hassyxyx67wg5xmgc", // Replace!!
			ExtraSupplyFetchAddr:         "fetch122j02czdt5ca8cf576wy2hassyxyx67wg5xmgc", // Replace!!
			VestingCollisionDestAddr:     "fetch122j02czdt5ca8cf576wy2hassyxyx67wg5xmgc", // Replace!!
			CommunityPoolBalanceDestAddr: "cudos1nj49l56x7sss5hqyvfmctxr3mq64whg273g3x5",

			VestingPeriod:    3 * 30 * 24 * 60 * 60, // 3 months period
			NewMaxValidators: 91,

			BalanceConversionConstants: []Pair[string, sdk.Dec]{
				{"acudos", newDec("118.344")},
			},

			TotalCudosSupply: newInt("10000000000000000000000000000"),

			TotalFetchSupplyToMint: newInt("88946755672000000000000000"),

			NotVestedAccounts: []string{
				"cudos1qqz5ezf9ylgft0eq97d66v5aakynux540ds9mv",
			},

			NotDelegatedAccounts: []string{
				"cudos1qx3yaanre054nlq84qdzufsjmrrxcqxwzdkh6c",
			},

			MovedAccounts: []BalanceMovement{
				BalanceMovement{"cudos1h6r6g0pwq7kcys5jcvfm9r7gcj3n2753hvk2ym", "cudos1w63ph9e4l07vpx7xdnje43cr2tlnr4jsfm4mvq", nil, ""},
				BalanceMovement{"cudos1jxyc7lny4q7te6sj5xyt9j86kyz82vlfdprl4a", "cudos1tfmkdzx9hm8g28vpgc3xhhxjjn460wzkwtayxr", nil, ""},
			},

			BackupValidators: []string{"fetchvaloper14w6a4al72uc3fpfy4lqtg0a7xtkx3w7hda0vel"},
			ValidatorsMap: []Pair[string, string]{
				{"cudosvaloper1s5qa3dpghnre6dqfgfhudxqjhwsv0mx43xayku", "fetchvaloper14w6a4al72uc3fpfy4lqtg0a7xtkx3w7hda0vel"},
				{"cudosvaloper1ctcrpuyumt60733u0yd5htwzedgfae0n8gql5n", "fetchvaloper14w6a4al72uc3fpfy4lqtg0a7xtkx3w7hda0vel"},
			},
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
					CW2version: &CW2ContractVersion{
						Contract: "contract-fetch-reconciliation",
						Version:  "1.0.0",
					},
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
		CudosMerge: &CudosMergeConfigJSON{
			IbcTargetAddr:                    "cudos1c3qgr4df6u3awsz6rqwkxcpsef7aau7p23pew5",
			RemainingStakingBalanceAddr:      "cudos1nj49l56x7sss5hqyvfmctxr3mq64whg273g3x5",
			RemainingGravityBalanceAddr:      "cudos1nj49l56x7sss5hqyvfmctxr3mq64whg273g3x5",
			RemainingDistributionBalanceAddr: "cudos1nj49l56x7sss5hqyvfmctxr3mq64whg273g3x5",
			ContractDestinationFallbackAddr:  "cudos1nj49l56x7sss5hqyvfmctxr3mq64whg273g3x5",
			GenericModuleRemainingBalance:    "cudos1nj49l56x7sss5hqyvfmctxr3mq64whg273g3x5",

			CommissionFetchAddr:          "fetch15p3rl5aavw9rtu86tna5lgxfkz67zzr6ed4yhw", // Fetch ecosystem wallet
			ExtraSupplyFetchAddr:         "fetch1wp8fl6fl4je40cfh2reeyj6cvucve9s6antdav",
			VestingCollisionDestAddr:     "cudos1nj49l56x7sss5hqyvfmctxr3mq64whg273g3x5",
			CommunityPoolBalanceDestAddr: "cudos1dslwarknhfsw3pfjzxxf5mn28q3ewfectw0gta",

			VestingPeriod: 3 * 30 * 24 * 60 * 60, // 3 months period

			BalanceConversionConstants: []Pair[string, sdk.Dec]{
				{"acudos", newDec("266.629")}},

			TotalCudosSupply:       newInt("22530000000000000000000000000"),
			TotalFetchSupplyToMint: newInt("88946755672000000000000000"),

			NotVestedAccounts: []string{
				"cudos1nj49l56x7sss5hqyvfmctxr3mq64whg273g3x5",
				"cudos15jpukx39rtkt8w3u3gzwwvyptdeyejcjade6he",
				"fetch15p3rl5aavw9rtu86tna5lgxfkz67zzr6ed4yhw",
			},

			NotDelegatedAccounts: []string{
				"cudos1dslwarknhfsw3pfjzxxf5mn28q3ewfectw0gta",
				"cudos1nj49l56x7sss5hqyvfmctxr3mq64whg273g3x5",
				"cudos15jpukx39rtkt8w3u3gzwwvyptdeyejcjade6he",
			},

			// Active wallets given by Cudos
			//cudos1nj49l56x7sss5hqyvfmctxr3mq64whg273g3x5
			//cudos1c3qgr4df6u3awsz6rqwkxcpsef7aau7p23pew5
			//cudos1dslwarknhfsw3pfjzxxf5mn28q3ewfectw0gta
			//cudos15jpukx39rtkt8w3u3gzwwvyptdeyejcjade6he

			MovedAccounts: []BalanceMovement{
				{"cudos196nrmandtwz67d8h4h0ux7amlcluecglx00wlw", "cudos1nj49l56x7sss5hqyvfmctxr3mq64whg273g3x5", newIntRef("10000"), ""}, // Replace this
				{"cudos1xcwjdw09cc9dyshr4gt5520sgsh582mjj03jge", "cudos1dslwarknhfsw3pfjzxxf5mn28q3ewfectw0gta", nil, ""},                // Replace this
				{"cudos1ejmf96efvjp6pmsaj8djv3gpmnsvmpnctger4v", "cudos15p3rl5aavw9rtu86tna5lgxfkz67zzr6tp4ltv", nil, ""},                // Replace this
			},

			BackupValidators: []string{"fetchvaloper1m9cjw6xgt04f9ddw25fff3cfe2exgwk07eu46u", "fetchvaloper122j02czdt5ca8cf576wy2hassyxyx67wdsecml"},
			ValidatorsMap: []Pair[string, string]{
				{"cudosvaloper1qr5rt72yf7s340azajpxay6hw3z5ldner7r4jv", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper1vz78ezuzskf9fgnjkmeks75xum49hug6zeqfc4", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				//{"cudosvaloper1v08h0h7sv3wm0t9uaz6x2q26p0g63tyzw68ynj", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"}, // Active
				{"cudosvaloper1n8lx8qac4d3gj63m4av29q755hy83kchzfkshd", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper16ndhv2a69jrwv32e0smpz6fy57kdsx36egshcm", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				{"cudosvaloper1vlpn2rne3vrms9gtjpwl2vfx2mky5dmv44rqqr", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
				{"cudosvaloper1suxdyst8l9zw64rs8nd9yfygvlynjs7sqcqvfw", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper1n738v30yvnge2cc3lq773d8vyvqj3rejnspa3r", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				{"cudosvaloper15juh20tsfmx3ezustkn4ph7as67rcd5q4hv259", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
				{"cudosvaloper1zsc3uv725d59t654t5vcmcflt2k68ahfvxqd6k", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper1j9csyu6dptzwhrmv9fyhaw2hzw44mn5jjg60uz", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				//{"cudosvaloper14ephhlg2pmjgw2rzf7wqgynrutrdlp2yh5c7m7", "fetchvaloper1m9cjw6xgt04f9ddw25fff3cfe2exgwk07eu46u", // Active (map to Cudos validator)
				{"cudosvaloper1hap6rg0kk0pgmew5vua99tm2ue96f8aahyw4zj", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper1avnl53xv6kj7dk5u35jlk52vxwg95n9zlh3jl8", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				{"cudosvaloper17594fddghhfjfwl634qj82tkqtq6wqt9x7aqea", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
				{"cudosvaloper1zje9zjjx3k8u3g6d57daxq4q9wpsqqyfvzpu3c", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper1r9nsfhl2ul5alytrqlsvynexg7563s9nxn3fjn", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				{"cudosvaloper1dslwarknhfsw3pfjzxxf5mn28q3ewfeckapf2q", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
				{"cudosvaloper1kw3y4p2gc5u025wl2wwyek94yxwdcgj3nupvtj", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper1pasz9ppwxggyty7fl5745c6lfqrs8g2shhs74e", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				{"cudosvaloper1tmtmme96cuj0fn94xg463dq3zjfy8ad0uxsc4u", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
				{"cudosvaloper1wkd0e3zzamaa2xwawe5tvh80n0qzrcwj7pzgdq", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper10hpr2qpmmp3da7trujcezutx3vaenysruemsd7", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				{"cudosvaloper134a4es94hjqqej732cymf0w3988zh3c4pqfy0s", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
				{"cudosvaloper1n746qrg5ja87mfu9y6u0acwe20jz045uky99le", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper14rqsrn8mr2lcyzra2dy77emgurw00tlm4dxj8u", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				{"cudosvaloper1evkgaxd24y7n5cghgh6d4q9wcr8tuft8wlrv06", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
				{"cudosvaloper19hyhj7ace4r6uppv4ll7vh8mfzv4ed6x0dtl4l", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper1tutujcdd40rcdjlmuuxtvqsyfspdr64ft6mnnv", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				{"cudosvaloper16ksrhugywq8yewndp0thsceqwgcupjvras48xl", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
				{"cudosvaloper1a08d0kwjgns5w09narc4zvfmp349dlch2j7j7x", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper1zdhktkrjyye0vn877rg40unec0mele5e4uxrav", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				{"cudosvaloper1yvd9h3nhukllwdhzy6r48q82aq900nwf8wfdvt", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
				{"cudosvaloper1g03p85mj85rzkntt5u5qzjw7k5hedwwz0xre4h", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper1spx72c3lskj7jh4svauy7cnez4e7wf4zavegrv", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				{"cudosvaloper16c50t6wwfagzcswdnmnk2f5ntdyrjsxyzjjgkf", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
				{"cudosvaloper1uckrunvmqugrhem0hlu9r8j08x0uhufryc9mc6", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper17lkl2ce4ee440teswxr757kafyz3m9x655pq7g", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				{"cudosvaloper1y9ag39jrrdqq0wmadd7a49nqxzcnjr2qf48nze", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
				{"cudosvaloper12jk94gydmc20ygn68c6ly0cans3zytvx2svdwy", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper1d0new9u2f80rcmmlw0zftxeh0385c2gwugrdx3", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				{"cudosvaloper10ku077u0mfgj5pt8mla8hd4n0lw9yyqj7p68z0", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
				{"cudosvaloper1esx247nhaal7epksdwh8zpa25vtjvy570lqa48", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper1a0gdnlchyh5flrtqmvvpdcxf3tk5kgngtufwg3", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				{"cudosvaloper1zd6jfttjrh8rr6hkkce7v550wsrsx9lsc0w8sq", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
				{"cudosvaloper1wvqepntffmqfzngujy24x6zue49u724twt5zlf", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper1e57dml5afa2gy5wrv9wd0dhc7jsca6cdjkd9a4", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				{"cudosvaloper19x6hxpuasshs80vy39j2kh2rcjhz0982wwz9r0", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
				{"cudosvaloper1xgqr7pn80g6w398wzdmye3lu6wsgk32sq5zz89", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper1f245xp5v3gg5cuwz4eg3j42mxhjyqhy0629nec", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				{"cudosvaloper1wyv883wzp84d7hzf3q2jeq3l43aga72td8spvs", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
				{"cudosvaloper1sh38ackq0aquq9kd6s2dsfxm43vwpxgf98le9q", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper1jxyc7lny4q7te6sj5xyt9j86kyz82vlfsjd75q", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				{"cudosvaloper1j5ssarsg0yqah4a3s0ejhkenlldg65mt0vpudr", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
				{"cudosvaloper1ns848uhesnue734232srhgjf5vfyuhd4spt2kk", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper1n7t5l7ck9slnnpl5pt0smua0p34qfhusv8rfsf", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				{"cudosvaloper16n9r2nkg06ewuun9hmkt8c6urjsyv3ruee3ypj", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
				{"cudosvaloper1706mmp25jy7s6xymdm9ar8pvny8eynwa8kxycp", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper1x5wgh6vwye60wv3dtshs9dmqggwfx2ld0zt8n8", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				{"cudosvaloper1ayq6mrlk5neyuugv7u0wt2tn8zf7e6dxhp2cd2", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
				{"cudosvaloper1a025fyry0gpk0c9sec00u938jvevvvjevp5a8e", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper18ygfl4ahy988k4skf6mat0ya8n3sj565qqufdf", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				{"cudosvaloper1gz2c6rkwf9quztvwynwr2hsa9lmldnjlu2ypxq", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
				{"cudosvaloper1anrcmsms8l3rm0td3qspm375vf3pu329gjv3xx", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper17h6h89mm5jxhxjsqjlxths8g8f5xhw40ss4nfa", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				{"cudosvaloper1qx9agmqrruhhyhmdz4ncxdkl769fv7kc0xfk8c", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
				{"cudosvaloper1per8wmhqh9868j42v8a4rfmkdef5eymap2yde6", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper1y4ye66cymhnl0pexpcja84ujcjjg0dr0gkjlw0", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				{"cudosvaloper18mqn8vssedje6ux24apmpl5y8fsn7ajv2evdpk", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
				{"cudosvaloper125xd54s26aylmnys3ruh2yrtyw5p2wjfhqrgzt", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper1sda6ngnryjv8lpqefpdqycl2ttg7gcspkf3tvc", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				{"cudosvaloper1nutnsjtxx37spn2juhgn046e6p324nx537eclp", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
				{"cudosvaloper16q3f9ech805pa995x9u5g3hqk0cvtah8mnaepp", "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k"},
				{"cudosvaloper1mnc7gm9sazrmcfdkshhmx3f0s4n2wp94gavnng", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg"},
				{"cudosvaloper1aqndlcfcyw3adhwe24gngnw7e8hy69v79jdh3l", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
			},
		},
	},
}

type NetworkConfig struct {
	MergeSourceChainID string `json:"merge_source_chain_id"`
	DestinationChainID string `json:"destination_chain_id"`

	ReconciliationInfo *ReconciliationInfo   `json:"reconciliation_info,omitempty"`
	Contracts          *ContractSet          `json:"contracts,omitempty"`
	CudosMerge         *CudosMergeConfigJSON `json:"cudos_merge,omitempty"`
}

func LoadNetworkConfigFromFile(configFilePath string) (*NetworkConfig, *[]byte, error) {
	// Open the JSON file
	file, err := os.Open(configFilePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open config file: %v", err)
	}
	defer file.Close()

	// Read the file contents
	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read config file: %v", err)
	}

	// Initialize an empty struct to hold the JSON data
	var config NetworkConfig

	// Unmarshal the JSON data into the struct
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if config.ReconciliationInfo.InputCSVRecords == nil {
		if val, exists := ReconciliationRecords[config.DestinationChainID]; exists {
			config.ReconciliationInfo.InputCSVRecords = val
		}
	}

	return &config, &byteValue, nil
}

func LoadAndVerifyNetworkConfigFromFile(configFilePath string, expectedSha256Hex *string) (*NetworkConfig, error) {
	config, byteValue, err := LoadNetworkConfigFromFile(configFilePath)

	if err != nil {
		return nil, err
	}

	if isVerified, actualHashHex, err := VerifySha256(*byteValue, expectedSha256Hex); err != nil {
		return nil, err
	} else if !isVerified {
		return nil, fmt.Errorf("failed to verify sha256: NetworkConfig file \"%s\" hash \"%s\" does not match expected hash \"%s\"", configFilePath, actualHashHex, *expectedSha256Hex)
	}

	return config, nil
}

type CudosMergeConfigJSON struct {
	IbcTargetAddr                    string `json:"ibc_target_addr"`                            // Cudos address
	RemainingStakingBalanceAddr      string `json:"remaining_staking_balance_addr"`             // Cudos account for remaining bonded and not-bonded pool balances
	RemainingGravityBalanceAddr      string `json:"remaining_gravity_balance_addr"`             // Cudos address
	RemainingDistributionBalanceAddr string `json:"remaining_distribution_balance_addr"`        // Cudos address
	ContractDestinationFallbackAddr  string `json:"contract_destination_fallback_addr"`         // Cudos address
	CommunityPoolBalanceDestAddr     string `json:"community_pool_balance_dest_addr,omitempty"` // Cudos address, funds are moved to destination chain community pool if not set
	GenericModuleRemainingBalance    string `json:"generic_module_remaining_balance"`           // Cudos address for all leftover funds remaining on module accounts after the processing

	CommissionFetchAddr      string `json:"commission_fetch_addr"`       // Fetch address for commission
	ExtraSupplyFetchAddr     string `json:"extra_supply_fetch_addr"`     // Fetch address for extra supply
	VestingCollisionDestAddr string `json:"vesting_collision_dest_addr"` // This gets converted to raw address, so it can be fetch or cudos address

	VestingPeriod    int64  `json:"vesting_period"`               // Vesting period
	NewMaxValidators uint32 `json:"new_max_validators,omitempty"` // Set new value for staking params max validators

	BalanceConversionConstants []Pair[string, sdk.Dec] `json:"balance_conversion_constants,omitempty"`

	TotalCudosSupply       sdk.Int `json:"total_cudos_supply"`
	TotalFetchSupplyToMint sdk.Int `json:"total_fetch_supply_to_mint"`

	NotVestedAccounts    []string          `json:"not_vested_accounts,omitempty"`
	NotDelegatedAccounts []string          `json:"not_delegated_accounts,omitempty"`
	MovedAccounts        []BalanceMovement `json:"moved_accounts,omitempty"`

	ValidatorsMap []Pair[string, string] `json:"validators_map,omitempty"`

	BackupValidators []string `json:"backup_validators,omitempty"`

	MaxToleratedRemainingDistributionBalance *sdk.Int `json:"max_remaining_distribution_module_balance,omitempty"`
	MaxToleratedRemainingStakingBalance      *sdk.Int `json:"max_remaining_staking_module_balance,omitempty"`
	MaxToleratedRemainingMintBalance         *sdk.Int `json:"max_remaining_mint_module_balance,omitempty"`
}

type CudosMergeConfig struct {
	Config *CudosMergeConfigJSON

	BalanceConversionConstants *OrderedMap[string, sdk.Dec]

	NotVestedAccounts    *OrderedMap[string, bool]
	NotDelegatedAccounts *OrderedMap[string, bool]

	ValidatorsMap *OrderedMap[string, string]
}

func NewCudosMergeConfig(config *CudosMergeConfigJSON) *CudosMergeConfig {
	retval := new(CudosMergeConfig)
	retval.Config = config

	retval.BalanceConversionConstants = NewOrderedMapFromPairs(config.BalanceConversionConstants)
	retval.NotVestedAccounts = NewOrderedSet(config.NotVestedAccounts)
	retval.NotDelegatedAccounts = NewOrderedSet(config.NotDelegatedAccounts)

	retval.ValidatorsMap = NewOrderedMapFromPairs(config.ValidatorsMap)

	return retval
}

type ReconciliationInfo struct {
	TargetAddress   string      `json:"target_address"`
	InputCSVRecords *[][]string `json:"input_csv_records,omitempty"`
}

type ContractSet struct {
	Reconciliation *Reconciliation  `json:"reconciliation,omitempty"`
	TokenBridge    *TokenBridge     `json:"token_bridge,omitempty"`
	Almanac        *ProdDevContract `json:"almanac,omitempty"`
	AName          *ProdDevContract `json:"a_name,omitempty"`
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
	Addr     string  `json:"addr"`
	NewAdmin *string `json:"new_admin,omitempty"`
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
	CW2version *CW2ContractVersion `json:"cw_2_version,omitempty"`
}

type CW2ContractVersion struct {
	Contract string `json:"contract"`
	Version  string `json:"version"`
}

type Reconciliation struct {
	Addr               string           `json:"addr"`
	NewAdmin           *string          `json:"new_admin,omitempty"`
	NewLabel           *string          `json:"new_label,omitempty"`
	NewContractVersion *ContractVersion `json:"new_contract_version,omitempty"`
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
	DevAddr  string `json:"dev_addr"`
	ProdAddr string `json:"prod_addr"`
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

func verifyAddress(address string, expectedPrefix *string) error {
	prefix, decodedAddrData, err := bech32.DecodeAndConvert(address)
	if err != nil {
		return fmt.Errorf("decoding of the '%s' address failed: %w", address, err)
	}
	if expectedPrefix != nil && prefix != *expectedPrefix {
		return fmt.Errorf("expected address prefix '%s', got '%s'", *expectedPrefix, prefix)
	}

	reconstructedAddr, err := bech32.ConvertAndEncode(prefix, decodedAddrData)
	if err != nil {
		return fmt.Errorf("encoding raw addr repr. of the '%s' orig. address failed: %w", address, err)
	}
	if address != reconstructedAddr {
		return fmt.Errorf("invalid address '%s'", address)
	}

	return nil
}

func VerifyConfig(cudosCfg *CudosMergeConfig, sourceAddrPrefix string, DestAddrPrefix string) error {
	expectedSourceValoperPrefix := sourceAddrPrefix + ValAddressPrefix
	expectedDestValoperPrefix := DestAddrPrefix + ValAddressPrefix

	for i := range cudosCfg.ValidatorsMap.Iterate() {
		srcValidator, DestValidator := i.Key, i.Value
		err := verifyAddress(srcValidator, &expectedSourceValoperPrefix)
		if err != nil {
			return err
		}
		err = verifyAddress(DestValidator, &expectedDestValoperPrefix)
		if err != nil {
			return err
		}
	}

	for _, notDelegatedAccount := range cudosCfg.NotDelegatedAccounts.Keys() {
		err := verifyAddress(notDelegatedAccount, &sourceAddrPrefix)
		if err != nil {
			return err
		}
	}

	for _, notVestedAccount := range cudosCfg.NotVestedAccounts.Keys() {
		err := verifyAddress(notVestedAccount, &sourceAddrPrefix)
		if err != nil {
			return err
		}
	}

	for _, movement := range cudosCfg.Config.MovedAccounts {
		err := verifyAddress(movement.SourceAddress, &sourceAddrPrefix)
		if err != nil {
			return err
		}
		err = verifyAddress(movement.DestinationAddress, &sourceAddrPrefix)
		if err != nil {
			return err
		}
		if movement.SourceAddress == movement.DestinationAddress {
			return fmt.Errorf("movement source and destination address is the same for %s", movement.SourceAddress)
		}
		if movement.Amount != nil && movement.Amount.IsNegative() {
			return fmt.Errorf("negative amount %s for movement from account %s to %s", movement.Amount, movement.SourceAddress, movement.DestinationAddress)
		}
	}

	err := verifyAddress(cudosCfg.Config.IbcTargetAddr, &sourceAddrPrefix)
	if err != nil {
		return fmt.Errorf("ibc targer address error: %v", err)
	}

	err = verifyAddress(cudosCfg.Config.RemainingStakingBalanceAddr, &sourceAddrPrefix)
	if err != nil {
		return fmt.Errorf("remaining staking balance address error: %v", err)
	}

	err = verifyAddress(cudosCfg.Config.RemainingGravityBalanceAddr, &sourceAddrPrefix)
	if err != nil {
		return fmt.Errorf("remaining gravity balance address error: %v", err)
	}

	err = verifyAddress(cudosCfg.Config.RemainingDistributionBalanceAddr, &sourceAddrPrefix)
	if err != nil {
		return fmt.Errorf("remaining distribution balance address error: %v", err)
	}

	err = verifyAddress(cudosCfg.Config.ContractDestinationFallbackAddr, &sourceAddrPrefix)
	if err != nil {
		return fmt.Errorf("contract destination fallback address error: %v", err)
	}

	err = verifyAddress(cudosCfg.Config.GenericModuleRemainingBalance, &sourceAddrPrefix)
	if err != nil {
		return fmt.Errorf("remaining general module balance address error: %v", err)
	}

	// Community pool address is optional
	if cudosCfg.Config.CommunityPoolBalanceDestAddr != "" {
		err = verifyAddress(cudosCfg.Config.CommunityPoolBalanceDestAddr, &sourceAddrPrefix)
		if err != nil {
			return fmt.Errorf("community pool balance destination address error: %v", err)
		}
	}

	err = verifyAddress(cudosCfg.Config.CommissionFetchAddr, &DestAddrPrefix)
	if err != nil {
		return fmt.Errorf("comission address error: %v", err)
	}

	err = verifyAddress(cudosCfg.Config.ExtraSupplyFetchAddr, &DestAddrPrefix)
	if err != nil {
		return fmt.Errorf("extra supply address error: %v", err)
	}

	err = verifyAddress(cudosCfg.Config.VestingCollisionDestAddr, &sourceAddrPrefix)
	if err != nil {
		return fmt.Errorf("vesting collision destination address error: %v", err)
	}

	if len(cudosCfg.Config.BalanceConversionConstants) == 0 {
		return fmt.Errorf("list of conversion constants is empty")
	}

	if len(cudosCfg.Config.BackupValidators) == 0 {
		return fmt.Errorf("list of backup validators is empty")
	}

	return nil
}
