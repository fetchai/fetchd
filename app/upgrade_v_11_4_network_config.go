package app

import sdk "github.com/cosmos/cosmos-sdk/types"

var (
	maxToleratedRemainingDistributionBalance, _ = sdk.NewIntFromString("1000000000000000000")
	maxToleratedRemainingStakingBalance, _      = sdk.NewIntFromString("100000000")
	maxToleratedRemainingMintBalance, _         = sdk.NewIntFromString("100000000")

	// Mainnet constants
	totalCudosSupply, _                            = sdk.NewIntFromString("10000000000000000000000000000")
	totalFetchSupplyToMint, _                      = sdk.NewIntFromString("88946755672000000000000000")
	acudosToafetExchangeRateReducedByCommission, _ = sdk.NewDecFromStr("118.344")

	// Testnet constants
	acudosToatestfetExchangeRateReducedByCommission, _ = sdk.NewDecFromStr("266.629")
	totalCudosTestnetSupply, _                         = sdk.NewIntFromString("22530000000000000000000000000")
	totalFetchTestnetSupplyToMint, _                   = sdk.NewIntFromString("88946755672000000000000000")
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
					cw2version: &CW2ContractVersion{
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

			backupValidators: []string{"fetchvaloper14w6a4al72uc3fpfy4lqtg0a7xtkx3w7hda0vel"},
			validatorsMap: map[string]string{
				"cudosvaloper1s5qa3dpghnre6dqfgfhudxqjhwsv0mx43xayku": "fetchvaloper14w6a4al72uc3fpfy4lqtg0a7xtkx3w7hda0vel",
				"cudosvaloper1ctcrpuyumt60733u0yd5htwzedgfae0n8gql5n": "fetchvaloper14w6a4al72uc3fpfy4lqtg0a7xtkx3w7hda0vel"},
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
					cw2version: &CW2ContractVersion{
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
				"acudos": acudosToatestfetExchangeRateReducedByCommission},

			totalCudosSupply:       totalCudosTestnetSupply,
			totalFetchSupplyToMint: totalFetchTestnetSupplyToMint,

			notVestedAccounts: map[string]bool{
				"cudos1qqqd8x95ectdhwujwkq2kq6y09qgeal7t67kyz": true,
			},

			notDelegatedAccounts: map[string]bool{
				"cudos1qqqd8x95ectdhwujwkq2kq6y09qgeal7t67kyz": true,
			},

			backupValidators: []string{"fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k", "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg", "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w"},
			validatorsMap: map[string]string{
				"cudosvaloper1qr5rt72yf7s340azajpxay6hw3z5ldner7r4jv": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1vz78ezuzskf9fgnjkmeks75xum49hug6zeqfc4": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper1v08h0h7sv3wm0t9uaz6x2q26p0g63tyzw68ynj": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1n8lx8qac4d3gj63m4av29q755hy83kchzfkshd": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper16ndhv2a69jrwv32e0smpz6fy57kdsx36egshcm": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper1vlpn2rne3vrms9gtjpwl2vfx2mky5dmv44rqqr": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1suxdyst8l9zw64rs8nd9yfygvlynjs7sqcqvfw": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1n738v30yvnge2cc3lq773d8vyvqj3rejnspa3r": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper15juh20tsfmx3ezustkn4ph7as67rcd5q4hv259": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1zsc3uv725d59t654t5vcmcflt2k68ahfvxqd6k": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1fdppshyftmfxsqy9ln66qkxc8q6faktdn9mlnq": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper1tqee5y5lv00z0p9gchjeky5w0256lgknjqsrpt": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1wx95r29g06awaspmfjlwmqnvtdvcfd8gjgc69a": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper14e64k86uf8yzud3vph3dw7czc30c2pdcw8uu3a": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper1k604fk7nf2zw3vagxx6u3gf90n3kdasahp9nk5": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1hhj7k93cpxtc3su5dfcmnx723am58f6mykges8": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1yse8fsc5pzpxzaxh0jxwdawf0nyacy0glg72aq": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper1xx85kun8rmzfh6cc82ynpaqx77z4h0akzuuygd": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper14w5g85yhx0ssqra5lh0y795mp6058k5ru7sjjr": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1kqrysavaahlzp4mxyjce3c84fc9peherzwhz07": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper1hlc6zhkthzkp9x88vn3lf75pvrkh3tzkgjt2jz": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper16ag06pkl53xy3e4r4pfc44n4r2v59fmcz6agxe": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1myyl3a0vlsqwqp0phjarqup0w3xl9wtqde4m36": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper1a8vtcyjk2v8939y0x0mhg3n4wdmx7r9620l9qn": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper17g4pjqjtxrxz35jcsx3jhns9e36epgs9e40rxt": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1l786h40rnqlgrd3pyp08dknm65wx8xk22gjq49": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper1qeylm4vnerwdgd34ar2yz6dkgzvfzh4ru98aud": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1zzq44hr9lvkm0xls4kelta93d9c3hagw5pxqqu": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper15jpukx39rtkt8w3u3gzwwvyptdeyejcjq7hmky": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper1cmcg0uh5qq3lsp5zg97sz0sss5zlrqk8vk6teq": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1lvprqa9djcqmj5yxdyf2zhc79qdncsu42ne66q": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1qedyz8rsfegfqj477lz0praa9x8jw4x7qye28y": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper1y36l53f8uj0t2t2pch43zngpt5hqcwzudzw3vu": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper198qaeg4wkf9tn7y345dhk2wyjmm0krdm68jp09": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1j9csyu6dptzwhrmv9fyhaw2hzw44mn5jjg60uz": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper14ephhlg2pmjgw2rzf7wqgynrutrdlp2yh5c7m7": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1hap6rg0kk0pgmew5vua99tm2ue96f8aahyw4zj": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1avnl53xv6kj7dk5u35jlk52vxwg95n9zlh3jl8": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper17594fddghhfjfwl634qj82tkqtq6wqt9x7aqea": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1zje9zjjx3k8u3g6d57daxq4q9wpsqqyfvzpu3c": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1r9nsfhl2ul5alytrqlsvynexg7563s9nxn3fjn": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper1dslwarknhfsw3pfjzxxf5mn28q3ewfeckapf2q": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1kw3y4p2gc5u025wl2wwyek94yxwdcgj3nupvtj": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1pasz9ppwxggyty7fl5745c6lfqrs8g2shhs74e": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper1tmtmme96cuj0fn94xg463dq3zjfy8ad0uxsc4u": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1wkd0e3zzamaa2xwawe5tvh80n0qzrcwj7pzgdq": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper10hpr2qpmmp3da7trujcezutx3vaenysruemsd7": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper134a4es94hjqqej732cymf0w3988zh3c4pqfy0s": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1n746qrg5ja87mfu9y6u0acwe20jz045uky99le": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper14rqsrn8mr2lcyzra2dy77emgurw00tlm4dxj8u": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper1evkgaxd24y7n5cghgh6d4q9wcr8tuft8wlrv06": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper19hyhj7ace4r6uppv4ll7vh8mfzv4ed6x0dtl4l": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1tutujcdd40rcdjlmuuxtvqsyfspdr64ft6mnnv": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper16ksrhugywq8yewndp0thsceqwgcupjvras48xl": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1a08d0kwjgns5w09narc4zvfmp349dlch2j7j7x": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1zdhktkrjyye0vn877rg40unec0mele5e4uxrav": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper1yvd9h3nhukllwdhzy6r48q82aq900nwf8wfdvt": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1g03p85mj85rzkntt5u5qzjw7k5hedwwz0xre4h": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1spx72c3lskj7jh4svauy7cnez4e7wf4zavegrv": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper16c50t6wwfagzcswdnmnk2f5ntdyrjsxyzjjgkf": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1uckrunvmqugrhem0hlu9r8j08x0uhufryc9mc6": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper17lkl2ce4ee440teswxr757kafyz3m9x655pq7g": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper1y9ag39jrrdqq0wmadd7a49nqxzcnjr2qf48nze": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper12jk94gydmc20ygn68c6ly0cans3zytvx2svdwy": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1d0new9u2f80rcmmlw0zftxeh0385c2gwugrdx3": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper10ku077u0mfgj5pt8mla8hd4n0lw9yyqj7p68z0": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1esx247nhaal7epksdwh8zpa25vtjvy570lqa48": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1a0gdnlchyh5flrtqmvvpdcxf3tk5kgngtufwg3": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper1zd6jfttjrh8rr6hkkce7v550wsrsx9lsc0w8sq": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1wvqepntffmqfzngujy24x6zue49u724twt5zlf": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1e57dml5afa2gy5wrv9wd0dhc7jsca6cdjkd9a4": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper19x6hxpuasshs80vy39j2kh2rcjhz0982wwz9r0": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1xgqr7pn80g6w398wzdmye3lu6wsgk32sq5zz89": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1f245xp5v3gg5cuwz4eg3j42mxhjyqhy0629nec": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper1wyv883wzp84d7hzf3q2jeq3l43aga72td8spvs": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1sh38ackq0aquq9kd6s2dsfxm43vwpxgf98le9q": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1jxyc7lny4q7te6sj5xyt9j86kyz82vlfsjd75q": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper1j5ssarsg0yqah4a3s0ejhkenlldg65mt0vpudr": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1ns848uhesnue734232srhgjf5vfyuhd4spt2kk": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1n7t5l7ck9slnnpl5pt0smua0p34qfhusv8rfsf": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper16n9r2nkg06ewuun9hmkt8c6urjsyv3ruee3ypj": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1706mmp25jy7s6xymdm9ar8pvny8eynwa8kxycp": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1x5wgh6vwye60wv3dtshs9dmqggwfx2ld0zt8n8": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper1ayq6mrlk5neyuugv7u0wt2tn8zf7e6dxhp2cd2": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1a025fyry0gpk0c9sec00u938jvevvvjevp5a8e": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper18ygfl4ahy988k4skf6mat0ya8n3sj565qqufdf": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper1gz2c6rkwf9quztvwynwr2hsa9lmldnjlu2ypxq": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1anrcmsms8l3rm0td3qspm375vf3pu329gjv3xx": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper17h6h89mm5jxhxjsqjlxths8g8f5xhw40ss4nfa": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper1qx9agmqrruhhyhmdz4ncxdkl769fv7kc0xfk8c": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper1per8wmhqh9868j42v8a4rfmkdef5eymap2yde6": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1y4ye66cymhnl0pexpcja84ujcjjg0dr0gkjlw0": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper18mqn8vssedje6ux24apmpl5y8fsn7ajv2evdpk": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper125xd54s26aylmnys3ruh2yrtyw5p2wjfhqrgzt": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1sda6ngnryjv8lpqefpdqycl2ttg7gcspkf3tvc": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper1nutnsjtxx37spn2juhgn046e6p324nx537eclp": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
				"cudosvaloper16q3f9ech805pa995x9u5g3hqk0cvtah8mnaepp": "fetchvaloper1rsane988vksrgp2mlqzclmt8wucxv0ej4hrn2k",
				"cudosvaloper1mnc7gm9sazrmcfdkshhmx3f0s4n2wp94gavnng": "fetchvaloper1je7r8yuqgaf5f2tx4z2f9008wp4jx0ct6msnzg",
				"cudosvaloper1aqndlcfcyw3adhwe24gngnw7e8hy69v79jdh3l": "fetchvaloper1edqmkwy4rh87020rvf9xn7kktyu7x894led46w",
			},
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
	cw2version *CW2ContractVersion
}

type CW2ContractVersion struct {
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
