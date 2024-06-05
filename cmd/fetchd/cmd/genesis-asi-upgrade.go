package cmd

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/binary"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/btcsuite/btcutil/bech32"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	ibccore "github.com/cosmos/ibc-go/v3/modules/core/24-host"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/types"
	"regexp"
	"strconv"
	"strings"
)

const (
	Bech32Chars        = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
	AddrDataLength     = 32
	WasmAddrDataLength = 52
	MaxAddrDataLength  = 100
	AddrChecksumLength = 6

	AccAddressPrefix  = ""
	ValAddressPrefix  = "valoper"
	ConsAddressPrefix = "valcons"

	NewAddrPrefix = "asi"
	OldAddrPrefix = "fetch"
)

var (
	// Reconciliation balances contract state key
	reconciliationBalancesKey              = prefixStringWithLength("balances")
	reconciliationTotalBalanceKey          = []byte("total_balance")
	reconciliationNOutstandingAddressesKey = []byte("n_outstanding_addresses")
	reconciliationStatesKey                = []byte("state")

	// Mobix staking contract keys
	stakesKey        = prefixStringWithLength("stakes")
	unbondEntriesKey = prefixStringWithLength("unbond_entries")
	configKey        = []byte("config")

	// Fcc issuance contract keys
	claimsKey               = prefixStringWithLength("claims")
	issuanceKey             = prefixStringWithLength("issuance")
	issuerAddressKey        = []byte("issuer_address")
	sourceOfFundsAddressKey = []byte("source_of_funds_address")
	cw20AddressKey          = []byte("cw20_address")
	issuanceFccAddressKey   = []byte("issuance_fcc_address")

	// Fcc Cw20 contract keys
	marketingInfoKey    = []byte("marketing_info")
	tokenInfoKey        = []byte("token_info")
	allowanceSpenderKey = prefixStringWithLength("allowance_spender")
	allowanceKey        = prefixStringWithLength("allowance")
	balanceKey          = prefixStringWithLength("balance")
)

//go:embed reconciliation_data.csv
var reconciliationData []byte

//go:embed reconciliation_data_testnet.csv
var reconciliationDataTestnet []byte

var networkInfos = map[string]NetworkConfig{
	"fetchhub-4": {
		NewChainID:     "asi-1",
		NewDescription: "ASI Network token",
		DenomInfo: DenomInfo{
			NewBaseDenom: "asi",
			NewDenom:     "aasi",
			OldDenom:     "afet",
		},
		SupplyInfo: SupplyInfo{
			SupplyToMint:              "0",
			UpdatedSupplyOverflowAddr: "fetch15p3rl5aavw9rtu86tna5lgxfkz67zzr6ed4yhw",
		},
		IbcTargetAddr: "fetch1zydegef0z6lz4gamamzlnu52ethe8xnm0xe5fkyrgwumsh9pplus5he63f",
		ReconciliationInfo: &ReconciliationInfo{
			TargetAddress:   "fetch1tynmzk68pq6kzawqffrqdhquq475gw9ccmlf9gk24mxjjy6ugl3q70aeyd",
			InputCSVRecords: readInputReconciliationData(reconciliationData),
		},
		Contracts: &Contracts{
			Almanac: &Almanac{
				ProdAddr: "fetch1mezzhfj7qgveewzwzdk6lz5sae4dunpmmsjr9u7z0tpmdsae8zmquq3y0y",
			},
			AName: &AName{
				ProdAddr: "fetch1479lwv5vy8skute5cycuz727e55spkhxut0valrcm38x9caa2x8q99ef0q",
			},
			MobixStaking: &MobixStaking{
				Addresses: []string{"fetch174kgn5rtw4kf6f938wm7kwh70h2v4vcfcnfkl0", "fetch1sh36qn08g4cqg685cfzmyxqv2952q6r8actxru"},
			},
			TokenBridge: &TokenBridge{
				Addr:     "fetch1qxxlalvsdjd07p07y3rc5fu6ll8k4tmetpha8n",
				NewAdmin: getStringPtr("fetch15p3rl5aavw9rtu86tna5lgxfkz67zzr6ed4yhw"),
			},
			Reconciliation: &Reconciliation{
				Addr:     "fetch1tynmzk68pq6kzawqffrqdhquq475gw9ccmlf9gk24mxjjy6ugl3q70aeyd",
				NewAdmin: getStringPtr("fetch15p3rl5aavw9rtu86tna5lgxfkz67zzr6ed4yhw"),
			},
			FccCw20: &FccCw20{
				Addr: "fetch1vsarnyag5d2c72k86yh2aq4l5jxhwz8fms6yralxqggxzmmwnq4q0avxv7",
			},
		},
	},

	"dorado-1": {
		NewChainID:     "eridanus-1",
		NewDescription: "ASI Network testnet token",
		DenomInfo: DenomInfo{
			NewBaseDenom: "testasi",
			NewDenom:     "atestasi",
			OldDenom:     "atestfet",
		},
		SupplyInfo: SupplyInfo{
			SupplyToMint:              "0",
			UpdatedSupplyOverflowAddr: "fetch1faucet4p2h432pxlh9ez8jfcl9jyr2ndlx2992",
		},
		IbcTargetAddr: "fetch18rlg4hs2p03yuvvdu389pe65qa789asmyqsfftdxsh2qjfwmt94qmrf7g0",
		ReconciliationInfo: &ReconciliationInfo{
			TargetAddress:   "fetch1g5ur2wc5xnlc7sw9wd895lw7mmxz04r5syj3s6ew8md6pvwuweqqavkgt0",
			InputCSVRecords: readInputReconciliationData(reconciliationDataTestnet),
		},
		Contracts: &Contracts{
			Almanac: &Almanac{
				ProdAddr: "fetch1tjagw8g8nn4cwuw00cf0m5tl4l6wfw9c0ue507fhx9e3yrsck8zs0l3q4w",
				DevAddr:  "fetch135h26ys2nwqealykzey532gamw4l4s07aewpwc0cyd8z6m92vyhsplf0vp",
			},
			AName: &AName{
				ProdAddr: "fetch1mxz8kn3l5ksaftx8a9pj9a6prpzk2uhxnqdkwuqvuh37tw80xu6qges77l",
				DevAddr:  "fetch1kewgfwxwtuxcnppr547wj6sd0e5fkckyp48dazsh89hll59epgpspmh0tn",
			},
			MobixStaking: &MobixStaking{
				Addresses: []string{"fetch1xr3rq8yvd7qplsw5yx90ftsr2zdhg4e9z60h5duusgxpv72hud3szdul6e"},
			},
			FccCw20: &FccCw20{
				Addr: "fetch1s0p7pwtm8qhvh2sfpg0ajgl20hwtehr0vcztyeku0vkzzvg044xqx4t7pt",
			},
			FccIssuance: &FccIssuance{
				Addr: "fetch17z773v8ree3e75s5sme38vvenlcyavcfs2ct3y6w77rwa5ag3srslelug5",
			},
			Reconciliation: &Reconciliation{
				Addr: "fetch1g5ur2wc5xnlc7sw9wd895lw7mmxz04r5syj3s6ew8md6pvwuweqqavkgt0",
			},
		},
	},
}

// ASIGenesisUpgradeCmd returns replace-genesis-values cobra Command.
func ASIGenesisUpgradeCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "asi-genesis-upgrade",
		Short: "This command carries out a full upgrade of the genesis file to the new ASI chain parameters.",
		Long: `The following command will upgrade the current genesis file to the new ASI chain parameters. The following changes will be made:
              - Chain ID will be updated to "asi-1"
              - The native coin denom will be updated to "asi"
              - The denom metadata will be updated to the new ASI token
              - The address prefix will be updated to "asi"
              - The old fetch addresses will be updated to the new asi addresses, e.g. asivaloper1, asivalcons1, asi1, etc.
              - The bridge contract admin will be updated to the new address
              - The IBC channel funds will be transferred to the IBC withdrawal address
              - The reconciliation withdrawal funds (if applicable) will be transferred to the reconciliation withdrawal address
`,

		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			genFile := config.GenesisFile()

			_, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			// fetch the network config using chain-id
			var ok bool
			var networkConfig NetworkConfig
			if networkConfig, ok = networkInfos[genDoc.ChainID]; !ok {
				return fmt.Errorf("network not found, no match for Chain-ID in genesis file")
			}

			// unmarshal the app state
			var jsonData map[string]interface{}
			if err = json.Unmarshal(genDoc.AppState, &jsonData); err != nil {
				return fmt.Errorf("failed to unmarshal app state: %w", err)
			}

			// create a new manifest
			manifest := ASIUpgradeManifest{}

			// replace chain-id
			ASIGenesisUpgradeReplaceChainID(genDoc, networkConfig)

			// replace bridge contract admin
			ASIGenesisUpgradeReplaceBridgeAdmin(jsonData, networkConfig)

			// update mobix staking contract
			ASIGenesisUpgradeUpdateMobixStakingContract(jsonData, networkConfig)

			// update fcc cw20 contract
			ASIGenesisUpgradeUpdateFccCw20Contract(jsonData, networkConfig)

			// replace almanac contract state
			ASIGenesisUpgradeReplaceAlmanacState(jsonData, networkConfig)

			// replace aname contract state
			ASIGenesisUpgradeReplaceANameState(jsonData, networkConfig)

			// update fcc issuance contract
			ASIGenesisUpgradeUpdateFccIssuanceContract(jsonData, networkConfig)

			// withdraw balances from IBC channels
			ASIGenesisUpgradeWithdrawIBCChannelsBalances(jsonData, networkConfig, &manifest)

			// withdraw balances from reconciliation addresses
			ASIGenesisUpgradeWithdrawReconciliationBalances(jsonData, networkConfig, &manifest)

			// set denom metadata in bank module
			ASIGenesisUpgradeReplaceDenomMetadata(jsonData, networkConfig)

			// replace denom across the genesis file
			ASIGenesisUpgradeReplaceDenom(jsonData, networkConfig)

			// supplement the genesis supply
			ASIGenesisUpgradeASISupply(jsonData, networkConfig, &manifest)

			// replace addresses across the genesis file
			ASIGenesisUpgradeReplaceAddresses(jsonData, networkConfig)

			if err = SaveASIManifest(&manifest, config); err != nil {
				return err
			}

			var encodedAppState []byte
			if encodedAppState, err = json.Marshal(jsonData); err != nil {
				return err
			}

			genDoc.AppState = encodedAppState
			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring backend (os|file|kwallet|pass|test)")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

type Bytes []byte

func (a Bytes) StartsWith(with []byte) bool {
	return len(a) >= len(with) && bytes.Compare(a[0:len(with)], with) == 0
}

func replaceAddressInContractStateKey2(keyBytes []byte, prefix []byte) string {
	address1StartIdx := len(prefix) + 2
	address1Len := int(binary.BigEndian.Uint16(keyBytes[len(prefix):address1StartIdx]))
	address2StartIdx := address1StartIdx + address1Len

	address1, err := convertAddressToASI(string(keyBytes[address1StartIdx:address2StartIdx]), AccAddressPrefix)
	if err != nil {
		panic(err)
	}

	address2, err := convertAddressToASI(string(keyBytes[address2StartIdx:]), AccAddressPrefix)
	if err != nil {
		panic(err)
	}

	// set to new address length
	address1Len = len(address1)

	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	if _, err := writer.Write(prefix); err != nil {
		panic(err)
	}

	if err := binary.Write(writer, binary.BigEndian, uint16(address1Len)); err != nil {
		panic(err)
	}

	if _, err := writer.WriteString(address1); err != nil {
		panic(err)
	}
	if _, err := writer.WriteString(address2); err != nil {
		panic(err)
	}

	if err := writer.Flush(); err != nil {
		panic(err)
	}

	return hex.EncodeToString(buffer.Bytes())
}

func ASIGenesisUpgradeUpdateFccCw20Contract(jsonData map[string]interface{}, networkInfo NetworkConfig) {
	if networkInfo.Contracts == nil || networkInfo.Contracts.FccCw20 == nil {
		return
	}

	FccCw20Address := networkInfo.Contracts.FccCw20.Addr
	FccCw20Contract := getContractFromAddr(FccCw20Address, jsonData)

	re := regexp.MustCompile(fmt.Sprintf(`%s%s1([%s]{%d,%d})`, OldAddrPrefix, "", Bech32Chars, AddrDataLength+AddrChecksumLength, MaxAddrDataLength))

	states := FccCw20Contract["contract_state"].([]interface{})
	for i, state := range states {
		stateMap := state.(map[string]interface{})
		hexKey := stateMap["key"].(string)
		b64Value := stateMap["value"].(string)

		keyBytes, err := hex.DecodeString(hexKey)
		if err != nil {
			panic(err)
		}

		valueBytes, err := base64.StdEncoding.DecodeString(b64Value)
		if err != nil {
			panic(err)
		}

		updatedKey := hexKey
		updatedValue := b64Value
		_keyBytes := Bytes(keyBytes)
		switch {
		case _keyBytes.StartsWith(balanceKey):
			updatedKey = replaceAddressInContractStateKey(keyBytes, balanceKey)
		case _keyBytes.StartsWith(tokenInfoKey):
			updatedValue = replaceAddressInContractStateValue(re, string(valueBytes))
		case _keyBytes.StartsWith(marketingInfoKey):
			updatedValue = replaceAddressInContractStateValue(re, string(valueBytes))
		case _keyBytes.StartsWith(allowanceSpenderKey):
			updatedKey = replaceAddressInContractStateKey2(keyBytes, allowanceSpenderKey)
		case _keyBytes.StartsWith(allowanceKey):
			updatedKey = replaceAddressInContractStateKey2(keyBytes, allowanceKey)
		}

		states[i] = map[string]interface{}{
			"key":   updatedKey,
			"value": updatedValue,
		}
	}
}

func ASIGenesisUpgradeUpdateFccIssuanceContract(jsonData map[string]interface{}, networkInfo NetworkConfig) {
	if networkInfo.Contracts == nil || networkInfo.Contracts.FccIssuance == nil {
		return
	}

	FccIssuanceContractAddr := networkInfo.Contracts.FccIssuance.Addr
	FccIssuanceContract := getContractFromAddr(FccIssuanceContractAddr, jsonData)
	re := regexp.MustCompile(fmt.Sprintf(`%s%s1([%s]{%d,%d})`, OldAddrPrefix, "", Bech32Chars, AddrDataLength+AddrChecksumLength, MaxAddrDataLength))

	replaceContractValueString := func(value string) string {
		newVal := re.ReplaceAllStringFunc(value, func(match string) string {
			newAddr, err := convertAddressToASI(match, AccAddressPrefix)
			if err != nil {
				panic(err)
			}
			return newAddr
		})
		return base64.StdEncoding.EncodeToString([]byte(newVal))
	}

	states := FccIssuanceContract["contract_state"].([]interface{})
	for i, val := range states {
		state := val.(map[string]interface{})
		hexKey := state["key"].(string)
		b64Value := state["value"].(string)

		valueBytes, err := base64.StdEncoding.DecodeString(b64Value)
		if err != nil {
			panic(err)
		}

		keyBytes, err := hex.DecodeString(hexKey)
		if err != nil {
			panic(err)
		}

		updatedKey := hexKey
		updatedValue := b64Value

		_keyBytes := Bytes(keyBytes)
		switch {
		case _keyBytes.StartsWith(claimsKey):
			updatedKey = replaceAddressInContractStateKey(keyBytes, claimsKey)
		case _keyBytes.StartsWith(issuerAddressKey):
			updatedValue = replaceContractValueString(string(valueBytes))
		case _keyBytes.StartsWith(issuanceKey):
			updatedKey = replaceAddressInContractStateKey(keyBytes, issuanceKey)
		case _keyBytes.StartsWith(sourceOfFundsAddressKey):
			updatedValue = replaceContractValueString(string(valueBytes))
		case _keyBytes.StartsWith(cw20AddressKey):
			updatedValue = replaceContractValueString(string(valueBytes))
		case _keyBytes.StartsWith(issuanceFccAddressKey):
			updatedValue = replaceContractValueString(string(valueBytes))
		}

		states[i] = map[string]interface{}{
			"key":   updatedKey,
			"value": updatedValue,
		}
	}
}

func ASIGenesisUpgradeUpdateMobixStakingContract(jsonData map[string]interface{}, networkInfo NetworkConfig) {
	if networkInfo.Contracts == nil || networkInfo.Contracts.MobixStaking == nil || len(networkInfo.Contracts.MobixStaking.Addresses) < 1 {
		return
	}

	for _, mobixStakingContractAddress := range networkInfo.Contracts.MobixStaking.Addresses {
		mobixStakingContract := getContractFromAddr(mobixStakingContractAddress, jsonData)

		re := regexp.MustCompile(fmt.Sprintf(`%s%s1([%s]{%d,%d})$`, OldAddrPrefix, "", Bech32Chars, AddrDataLength+AddrChecksumLength, MaxAddrDataLength))
		states := mobixStakingContract["contract_state"].([]interface{})
		for i, val := range states {
			state := val.(map[string]interface{})
			hexKey := state["key"].(string)
			b64Value := state["value"].(string)

			valueBytes, err := base64.StdEncoding.DecodeString(b64Value)
			if err != nil {
				panic(err)
			}

			updatedKey := hexKey
			updatedValue := b64Value

			keyBytes, err := hex.DecodeString(hexKey)
			if err != nil {
				panic(err)
			}

			_keyBytes := Bytes(keyBytes)
			switch {
			case _keyBytes.StartsWith(stakesKey):
				updatedKey = replaceAddressInContractStateKey(keyBytes, stakesKey)
			case _keyBytes.StartsWith(unbondEntriesKey):
				updatedKey = replaceAddressInContractStateKey(keyBytes, unbondEntriesKey)
			case _keyBytes.StartsWith(configKey):
				updatedValue = replaceAddressInContractStateValue(re, string(valueBytes))
			}

			states[i] = map[string]interface{}{
				"key":   updatedKey,
				"value": updatedValue,
			}
		}
	}
}

func replaceAddressInContractStateValue(re *regexp.Regexp, value string) string {
	var newValue []byte

	// replace value
	valJson := make(map[string]interface{})
	if err := json.Unmarshal([]byte(value), &valJson); err != nil {
		panic(err)
	}

	var err error
	replaceAddresses(AccAddressPrefix, valJson, AddrDataLength+AddrChecksumLength)
	newValue, err = json.Marshal(valJson)
	if err != nil {
		panic(err)
	}

	// return new value
	return base64.StdEncoding.EncodeToString(newValue)
}

func prefixStringWithLength(val string) []byte {
	length := len(val)

	if length > 0xFFFF {
		panic("length of input string does not fit into uint16")
	}

	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	if err := binary.Write(writer, binary.BigEndian, uint16(length)); err != nil {
		panic(err)
	}

	if _, err := writer.WriteString(val); err != nil {
		panic(err)
	}

	if err := writer.Flush(); err != nil {
		panic(err)
	}

	return buffer.Bytes()
}

func replaceAddressInContractStateKey(keyBytes []byte, prefix []byte) string {
	newAddr, err := convertAddressToASI(string(keyBytes[len(prefix):]), AccAddressPrefix)
	if err != nil {
		panic(err)
	}
	key := append(prefix, []byte(newAddr)...)

	return hex.EncodeToString(key)
}

func reconciliationContractStateBalancesRecord(ethAddr string, coins sdk.Coins, networkConfig *NetworkConfig) (*map[string]string, sdk.Int) {
	amount := coins.AmountOfNoDenomValidation(networkConfig.DenomInfo.OldDenom)
	if amount.IsZero() {
		return nil, amount
	}
	if amount.IsNegative() {
		panic(fmt.Errorf("netgative amount value for ethereum '%s' address", ethAddr))
	}

	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	if _, err := writer.Write(reconciliationBalancesKey); err != nil {
		panic(err)
	}
	if _, err := writer.WriteString(ethAddr); err != nil {
		panic(err)
	}
	if err := writer.Flush(); err != nil {
		panic(err)
	}

	balanceRecord := map[string]string{
		"key":   hex.EncodeToString(buffer.Bytes()),
		"value": base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("\"%s\"", amount.String()))),
	}

	return &balanceRecord, amount
}

func addReconciliationContractStateBalancesRecord(contractStateRecords *[]interface{}, ethAddr string, coins sdk.Coins, networkConfig *NetworkConfig, manifest *ASIUpgradeManifest) {
	newContractStateBalancesRecord, amount := reconciliationContractStateBalancesRecord(ethAddr, coins, networkConfig)
	if newContractStateBalancesRecord != nil {
		*contractStateRecords = append(*contractStateRecords, *newContractStateBalancesRecord)
		manifest.Reconciliation.ContractState.Balances = append(manifest.Reconciliation.ContractState.Balances, ASIUpgradeReconciliationContractStateBalanceRecord{EthAddr: ethAddr, Amount: amount})
		manifest.Reconciliation.ContractState.AggregatedBalancesAmount = manifest.Reconciliation.ContractState.AggregatedBalancesAmount.Add(amount)
		manifest.Reconciliation.ContractState.NumberOfBalanceRecords += 1
	}
}

func addReconciliationContractState(contractStateRecords *[]interface{}, networkConfig *NetworkConfig, manifest *ASIUpgradeManifest) {
	totalBalanceRecordEnc := map[string]string{
		"key":   hex.EncodeToString(reconciliationTotalBalanceKey),
		"value": base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("\"%s\"", manifest.Reconciliation.ContractState.AggregatedBalancesAmount.String()))),
	}

	nOutstandingAddressesRecordEnc := map[string]string{
		"key":   hex.EncodeToString(reconciliationNOutstandingAddressesKey),
		"value": base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%v", manifest.Reconciliation.ContractState.NumberOfBalanceRecords))),
	}

	stateRecord := ReconciliationContractStateRecord{
		Denom:  networkConfig.DenomInfo.NewDenom,
		Paused: true,
	}

	var err error
	var stateRecordJSONStr []byte
	stateRecordJSONStr, err = json.Marshal(stateRecord)
	if err == nil {
		panic(err)
	}
	stateRecordEnc := map[string]string{
		"key":   hex.EncodeToString(reconciliationStatesKey),
		"value": base64.StdEncoding.EncodeToString(stateRecordJSONStr),
	}

	*contractStateRecords = append(*contractStateRecords, totalBalanceRecordEnc)
	*contractStateRecords = append(*contractStateRecords, nOutstandingAddressesRecordEnc)
	*contractStateRecords = append(*contractStateRecords, stateRecordEnc)
}

func ASIGenesisUpgradeReplaceDenomMetadata(jsonData map[string]interface{}, networkInfo NetworkConfig) {
	type jsonMap map[string]interface{}

	newBaseDenom := networkInfo.DenomInfo.NewBaseDenom
	oldDenom := networkInfo.DenomInfo.OldDenom
	newDenom := networkInfo.DenomInfo.NewDenom
	newDescription := networkInfo.NewDescription

	NewBaseDenomUpper := strings.ToUpper(newBaseDenom)

	newMetadata := jsonMap{
		"base": newDenom,
		"denom_units": []jsonMap{
			{
				"denom":    NewBaseDenomUpper,
				"exponent": 18,
			},
			{
				"denom":    fmt.Sprintf("m%s", newBaseDenom),
				"exponent": 15,
			},
			{
				"denom":    fmt.Sprintf("u%s", newBaseDenom),
				"exponent": 12,
			},
			{
				"denom":    fmt.Sprintf("n%s", newBaseDenom),
				"exponent": 9,
			},
			{
				"denom":    fmt.Sprintf("p%s", newBaseDenom),
				"exponent": 6,
			},
			{
				"denom":    fmt.Sprintf("f%s", newBaseDenom),
				"exponent": 3,
			},
			{
				"denom":    fmt.Sprintf("a%s", newBaseDenom),
				"exponent": 0,
			},
		},
		"description": newDescription,
		"display":     NewBaseDenomUpper,
		"name":        NewBaseDenomUpper,
		"symbol":      NewBaseDenomUpper,
	}

	bank := jsonData["bank"].(map[string]interface{})
	denomMetadata := bank["denom_metadata"].([]interface{})

	for i, metadata := range denomMetadata {
		denomUnit := metadata.(map[string]interface{})
		if denomUnit["base"] == oldDenom {
			denomMetadata[i] = newMetadata
			break
		}
	}
}

func ASIGenesisUpgradeReplaceChainID(genesisData *types.GenesisDoc, networkInfo NetworkConfig) {
	genesisData.ChainID = networkInfo.NewChainID
}

func ASIGenesisUpgradeReplaceBridgeAdmin(jsonData map[string]interface{}, networkInfo NetworkConfig) {
	if networkInfo.Contracts == nil || networkInfo.Contracts.TokenBridge == nil {
		return
	}

	tokenBridgeContractAddress := networkInfo.Contracts.TokenBridge.Addr
	tokenBridgeContract := getContractFromAddr(tokenBridgeContractAddress, jsonData)

	replaceContractAdmin(tokenBridgeContract, networkInfo.Contracts.TokenBridge.NewAdmin)
}

func ASIGenesisUpgradeReplaceDenom(jsonData map[string]interface{}, networkInfo NetworkConfig) {
	targets := map[string]struct{}{"denom": {}, "bond_denom": {}, "mint_denom": {}, "base_denom": {}, "base": {}}
	oldDenom := networkInfo.DenomInfo.OldDenom
	newDenom := networkInfo.DenomInfo.NewDenom

	crawlJson("", jsonData, -1, func(key string, value interface{}, idx int) interface{} {
		if str, ok := value.(string); ok {
			_, isInTargets := targets[key]
			if str == oldDenom && isInTargets {
				return newDenom
			}
		}
		return value
	})
}

func ASIGenesisUpgradeReplaceAlmanacState(jsonData map[string]interface{}, networkInfo NetworkConfig) {
	if networkInfo.Contracts == nil || networkInfo.Contracts.Almanac == nil {
		return
	}

	for _, addr := range []string{networkInfo.Contracts.Almanac.ProdAddr, networkInfo.Contracts.Almanac.DevAddr} {
		if addr == "" {
			continue
		}

		almanacContract := getContractFromAddr(addr, jsonData)

		// empty the almanac contract state
		almanacContract["contract_state"] = []interface{}{}
	}
}

func ASIGenesisUpgradeReplaceReconciliationState(jsonData map[string]interface{}, networkConfig NetworkConfig, manifest *ASIUpgradeManifest) {
	if networkConfig.Contracts == nil || networkConfig.Contracts.Reconciliation == nil || networkConfig.Contracts.Reconciliation.Addr != "" || len(manifest.Reconciliation.Transfers.Transfers) < 1 {
		return
	}

	reconciliationContract := getContractFromAddr(networkConfig.Contracts.Reconciliation.Addr, jsonData)
	var reconciliationContractState []interface{}

	contractStateManifest := ASIUpgradeReconciliationContractState{}
	manifest.Reconciliation.ContractState = &contractStateManifest

	replaceContractAdmin(reconciliationContract, networkConfig.Contracts.Reconciliation.NewAdmin)

	for _, transfer := range manifest.Reconciliation.Transfers.Transfers {
		addReconciliationContractStateBalancesRecord(&reconciliationContractState, transfer.EthAddr, transfer.Amount, &networkConfig, manifest)
	}

	addReconciliationContractState(&reconciliationContractState, &networkConfig, manifest)
	reconciliationContract["contract_state"] = reconciliationContractState
}

func ASIGenesisUpgradeReplaceANameState(jsonData map[string]interface{}, networkInfo NetworkConfig) {
	if networkInfo.Contracts == nil || networkInfo.Contracts.AName == nil {
		return
	}

	for _, addr := range []string{networkInfo.Contracts.AName.ProdAddr, networkInfo.Contracts.AName.DevAddr} {
		if addr == "" {
			continue
		}

		anameContract := getContractFromAddr(addr, jsonData)

		// empty the AName contract state
		anameContract["contract_state"] = []interface{}{}
	}
}

func getContractFromAddr(addr string, jsonData map[string]interface{}) map[string]interface{} {
	contracts := jsonData["wasm"].(map[string]interface{})["contracts"].([]interface{})
	for _, contract := range contracts {
		if contract.(map[string]interface{})["contract_address"] == addr {
			return contract.(map[string]interface{})
		}
	}
	panic("failed to find contract using provided address")
}

func ASIGenesisUpgradeReplaceAddresses(jsonData map[string]interface{}, networkInfo NetworkConfig) {
	// account addresses
	replaceAddresses(AccAddressPrefix, jsonData, AddrDataLength+AddrChecksumLength)
	// validator addresses
	replaceAddresses(ValAddressPrefix, jsonData, AddrDataLength+AddrChecksumLength)
	// consensus addresses
	replaceAddresses(ConsAddressPrefix, jsonData, AddrDataLength+AddrChecksumLength)
	// contract addresses
	replaceAddresses(AccAddressPrefix, jsonData, WasmAddrDataLength+AddrChecksumLength)
}

func replaceAddresses(addressTypePrefix string, jsonData map[string]interface{}, dataLength int) {
	re := regexp.MustCompile(fmt.Sprintf(`^%s%s1([%s]{%d})$`, OldAddrPrefix, addressTypePrefix, Bech32Chars, dataLength))

	crawlJson("", jsonData, -1, func(key string, value interface{}, idx int) interface{} {
		if str, ok := value.(string); ok {
			if !re.MatchString(str) {
				return value
			}
			newAddress, err := convertAddressToASI(str, addressTypePrefix)
			if err != nil {
				panic(err)
			}

			return newAddress
		}
		return value
	})
}

func ASIGenesisUpgradeWithdrawIBCChannelsBalances(jsonData map[string]interface{}, networkInfo NetworkConfig, manifest *ASIUpgradeManifest) {
	if networkInfo.IbcTargetAddr == "" {
		return
	}

	bank := jsonData[banktypes.ModuleName].(map[string]interface{})
	balances := bank["balances"].([]interface{})
	balanceMap := getGenesisBalancesMap(balances)
	ibcWithdrawalAddress := networkInfo.IbcTargetAddr

	manifest.IBC = &ASIUpgradeTransfers{
		Transfer: []ASIUpgradeTransfer{},
		To:       ibcWithdrawalAddress,
	}
	withdrawalBalanceIdx, ok := (*balanceMap)[ibcWithdrawalAddress]
	if !ok {
		panic("failed to find ibc withdrawal address in genesis balances")
	}

	ibc := jsonData[ibccore.ModuleName].(map[string]interface{})
	channelGenesis := ibc["channel_genesis"].(map[string]interface{})
	ibcChannels := channelGenesis["channels"].([]interface{})

	for _, channel := range ibcChannels {
		channelMap := channel.(map[string]interface{})
		channelId := channelMap["channel_id"].(string)
		portId := channelMap["port_id"].(string)

		// close channel
		channelMap["state"] = "STATE_CLOSED"

		rawAddr := ibctransfertypes.GetEscrowAddress(portId, channelId)
		channelAddr, err := sdk.Bech32ifyAddressBytes(OldAddrPrefix+AccAddressPrefix, rawAddr)
		if err != nil {
			panic(err)
		}

		balanceIdx, ok := (*balanceMap)[channelAddr]
		if !ok {
			// channel address not found in genesis balances
			continue
		}

		channelBalanceCoins := getCoinsFromInterfaceSlice(balances[balanceIdx])
		withdrawalBalanceCoins := getCoinsFromInterfaceSlice(balances[withdrawalBalanceIdx])

		manifest.IBC.Transfer = append(manifest.IBC.Transfer, ASIUpgradeTransfer{From: channelAddr, Amount: channelBalanceCoins})

		// add channel balance to withdrawal balance
		newWithdrawalBalanceCoins := withdrawalBalanceCoins.Add(channelBalanceCoins...)
		balances[withdrawalBalanceIdx].(map[string]interface{})["coins"] = getInterfaceSliceFromCoins(newWithdrawalBalanceCoins)

		// zero out the channel balance
		balances[balanceIdx].(map[string]interface{})["coins"] = []interface{}{}
	}
}

func getGenesisAccountSequenceMap(accounts []interface{}) *map[string]int {
	const ModuleAccount = "/cosmos.auth.v1beta1.ModuleAccount"
	accountMap := make(map[string]int)

	for _, acc := range accounts {
		accMap := acc.(map[string]interface{})
		accType := accMap["@type"]

		accData := acc
		if accType == ModuleAccount {
			accData = accMap["base_account"]
		}

		accDataMap := accData.(map[string]interface{})
		addr := accDataMap["address"].(string)
		sequence := accDataMap["sequence"].(string)

		sequenceInt, ok := strconv.Atoi(sequence)
		if ok != nil {
			panic("getGenesisAccountSequenceMap: failed to convert sequence to int")
		}
		accountMap[addr] = sequenceInt
	}

	return &accountMap
}

func ASIGenesisUpgradeWithdrawReconciliationBalances(jsonData map[string]interface{}, networkConfig NetworkConfig, manifest *ASIUpgradeManifest) {
	if networkConfig.ReconciliationInfo == nil {
		return
	}

	bank := jsonData[banktypes.ModuleName].(map[string]interface{})
	balances := bank["balances"].([]interface{})
	reconciliationWithdrawAddress := networkConfig.ReconciliationInfo.TargetAddress

	balanceMap := getGenesisBalancesMap(balances)

	auth := jsonData[authtypes.ModuleName].(map[string]interface{})
	accounts := auth["accounts"].([]interface{})
	accountSequenceMap := getGenesisAccountSequenceMap(accounts)

	reconciliationBalanceIdx, ok := (*balanceMap)[reconciliationWithdrawAddress]
	if !ok {
		panic("no match in genesis for reconciliation withdraw address")
	}

	manifest.Reconciliation = &ASIUpgradeReconciliation{
		Transfers: ASIUpgradeReconciliationTransfers{
			To: reconciliationWithdrawAddress,
		},
	}

	for _, row := range networkConfig.ReconciliationInfo.InputCSVRecords {
		ethAddr := row[0]
		addr := row[2]

		accSequence, ok := (*accountSequenceMap)[addr]
		if !ok {
			panic("no match in genesis for reconciliation address")
		}

		balanceIdx, ok := (*balanceMap)[addr]
		if !ok {
			continue
		}

		accBalance := balances[balanceIdx]
		// Function below sanitises returned coins = removes zero balances & sorts coins based on denom
		accBalanceCoins := getCoinsFromInterfaceSlice(accBalance)

		reconciliationBalance := balances[reconciliationBalanceIdx]
		reconciliationBalanceCoins := getCoinsFromInterfaceSlice(reconciliationBalance)

		// check if the reconciliation address is still dormant and contains funds
		if accSequence != 0 || !accBalanceCoins.IsAllPositive() {
			continue
		}

		// add reconciliation account balance to withdrawal balance
		newReconciliationBalanceCoins := reconciliationBalanceCoins.Add(accBalanceCoins...)
		reconciliationBalance.(map[string]interface{})["coins"] = getInterfaceSliceFromCoins(newReconciliationBalanceCoins)

		// zero out the reconciliation account balance
		balances[balanceIdx].(map[string]interface{})["coins"] = []interface{}{}

		manifest.Reconciliation.Transfers.Transfers = append(manifest.Reconciliation.Transfers.Transfers, ASIUpgradeReconciliationTransfer{From: addr, EthAddr: ethAddr, Amount: accBalanceCoins})
		manifest.Reconciliation.Transfers.NumberOfTransfers += 1
		manifest.Reconciliation.Transfers.AggregatedBalancesAmount = manifest.Reconciliation.Transfers.AggregatedBalancesAmount.Add(accBalanceCoins...)
	}

	ASIGenesisUpgradeReplaceReconciliationState(jsonData, networkConfig, manifest)
}

func ASIGenesisUpgradeASISupply(jsonData map[string]interface{}, networkInfo NetworkConfig, manifest *ASIUpgradeManifest) {
	denomInfo := networkInfo.DenomInfo
	supplyInfo := networkInfo.SupplyInfo
	additionalSupply, ok := sdk.NewIntFromString(supplyInfo.SupplyToMint)
	if !ok {
		panic("asi upgrade update supply: failed to convert new supply value to int")
	}

	if additionalSupply.IsZero() {
		return
	} else if additionalSupply.LT(sdk.ZeroInt()) {
		panic("asi upgrade update supply: additional supply value is negative")
	}

	bank := jsonData[banktypes.ModuleName].(map[string]interface{})
	supply := bank["supply"].([]interface{})
	balances := bank["balances"].([]interface{})
	balancesMap := getGenesisBalancesMap(bank["balances"].([]interface{}))

	var curSupply sdk.Int
	var curSupplyIdx int
	for idx, coin := range supply {
		coinData := coin.(map[string]interface{})
		if coinData["denom"] == denomInfo.NewDenom {
			curSupplyIdx = idx
			curSupply, ok = sdk.NewIntFromString(coinData["amount"].(string))
			if !ok {
				panic("asi upgrade update supply: failed to convert coin amount to int")
			}
			break
		}
	}

	overflowAddressBalance := balances[(*balancesMap)[supplyInfo.UpdatedSupplyOverflowAddr]]
	overflowAddressBalanceCoins := getCoinsFromInterfaceSlice(overflowAddressBalance)

	additionalSupplyCoin := sdk.NewCoin(denomInfo.NewDenom, additionalSupply)
	curSupplyCoin := sdk.NewCoin(denomInfo.NewDenom, curSupply)

	// add new coins to the current supply
	newSupplyCoins := curSupplyCoin.Add(additionalSupplyCoin)

	// add the additional coins to the overflow address balance
	overflowAddressBalanceCoins = overflowAddressBalanceCoins.Add(additionalSupplyCoin)

	// add the new supply mint record to the manifest
	supplyRecord := ASIUpgradeSupply{
		LandingAddress:       supplyInfo.UpdatedSupplyOverflowAddr,
		MintedAmount:         sdk.NewCoins(additionalSupplyCoin),
		ResultingSupplyTotal: sdk.NewCoins(newSupplyCoins),
	}
	manifest.Supply = &supplyRecord

	// update the supply in the bank module
	supply[curSupplyIdx].(map[string]interface{})["amount"] = newSupplyCoins.Amount.String()
	balances[(*balancesMap)[supplyInfo.UpdatedSupplyOverflowAddr]].(map[string]interface{})["coins"] = getInterfaceSliceFromCoins(overflowAddressBalanceCoins)
}

func convertAddressToASI(addr string, addressPrefix string) (string, error) {
	_, decodedAddrData, err := bech32.Decode(addr)
	if err != nil {
		return "", err
	}

	newAddress, err := bech32.Encode(NewAddrPrefix+addressPrefix, decodedAddrData)
	if err != nil {
		return "", err
	}

	err = sdk.VerifyAddressFormat(decodedAddrData)
	if err != nil {
		return "", err
	}

	return newAddress, nil
}

func replaceContractAdmin(genesisContractStruct map[string]interface{}, newAdmin *string) {
	if newAdmin == nil {
		return
	}

	contractInfo := genesisContractStruct["contract_info"].(map[string]interface{})
	contractInfo["admin"] = newAdmin
}

func crawlJson(key string, value interface{}, idx int, strHandler func(string, interface{}, int) interface{}) interface{} {
	switch val := value.(type) {
	case string:
		if strHandler != nil {
			return strHandler(key, val, idx)
		}
	case []interface{}:
		for i := range val {
			val[i] = crawlJson("", val[i], i, strHandler)
		}
	case map[string]interface{}:
		for k, v := range val {
			val[k] = crawlJson(k, v, -1, strHandler)
		}
	}
	return value
}

func getGenesisBalancesMap(balances []interface{}) *map[string]int {
	balanceMap := make(map[string]int)

	for i, balance := range balances {
		addr := balance.(map[string]interface{})["address"]
		if addr == nil {
			fmt.Println(balance)
		}
		addrStr := addr.(string)
		balanceMap[addrStr] = i
	}

	return &balanceMap
}

func getCoinsFromInterfaceSlice(data interface{}) sdk.Coins {
	balance := data.(map[string]interface{})["coins"]
	var balanceCoins sdk.Coins

	for _, coin := range balance.([]interface{}) {
		coinData := coin.(map[string]interface{})
		coinDenom := coinData["denom"].(string)
		coinAmount, ok := sdk.NewIntFromString(coinData["amount"].(string))
		if !ok {
			panic("ibc withdraw: failed to convert coin amount to int")
		}
		balanceCoins = append(balanceCoins, sdk.NewCoin(coinDenom, coinAmount))
	}

	balanceCoins = sdk.NewCoins(balanceCoins...)
	return balanceCoins
}

func getInterfaceSliceFromCoins(coins sdk.Coins) []interface{} {
	var balance []interface{}
	for _, coin := range coins {
		balance = append(balance, map[string]interface{}{
			"denom":  coin.Denom,
			"amount": coin.Amount.String(),
		})
	}
	return balance
}

func readInputReconciliationData(csvData []byte) [][]string {
	r := csv.NewReader(bytes.NewReader(csvData))
	records, err := r.ReadAll()
	if err != nil {
		panic(fmt.Sprintf("error reading reconciliation data: %v", err))
	}
	return records
}

func getStringPtr(val string) *string {
	return &val
}

type NetworkConfig struct {
	NewChainID         string
	NewDescription     string
	IbcTargetAddr      string
	ReconciliationInfo *ReconciliationInfo
	SupplyInfo         SupplyInfo
	DenomInfo          DenomInfo
	Contracts          *Contracts
}

type ReconciliationInfo struct {
	TargetAddress   string
	InputCSVRecords [][]string
}

type SupplyInfo struct {
	UpdatedSupplyOverflowAddr string
	SupplyToMint              string
}

type DenomInfo struct {
	NewBaseDenom string
	NewDenom     string
	OldDenom     string
}

type Contracts struct {
	TokenBridge    *TokenBridge
	Almanac        *Almanac
	AName          *AName
	MobixStaking   *MobixStaking
	FccIssuance    *FccIssuance
	FccCw20        *FccCw20
	Reconciliation *Reconciliation
}

type TokenBridge struct {
	Addr     string
	NewAdmin *string
}

type Almanac struct {
	DevAddr  string
	ProdAddr string
}

type AName struct {
	DevAddr  string
	ProdAddr string
}

type MobixStaking struct {
	Addresses []string
}

type FccCw20 struct {
	Addr string
}

type FccIssuance struct {
	Addr string
}

type Reconciliation struct {
	Addr     string
	NewAdmin *string
}

type ReconciliationContractStateRecord struct {
	Denom  string `json:"denom"`
	Paused bool   `json:"paused"`
}
