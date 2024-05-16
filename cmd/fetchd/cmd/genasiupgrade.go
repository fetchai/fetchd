package cmd

import (
	"bytes"
	_ "embed"
	"encoding/base64"
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
	"log"
	"regexp"
	"strconv"
	"strings"
)

const (
	Bech32Chars        = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
	AddrDataLength     = 32
	WasmAddrDataLength = 52
	AddrChecksumLength = 6

	AccAddressPrefix  = ""
	ValAddressPrefix  = "valoper"
	ConsAddressPrefix = "valcons"

	NewAddrPrefix = "asi"
	OldAddrPrefix = "fetch"
)

var ReconciliationTargetAddr = "fetch1rhrlzsx9z865dqen8t4v47r99dw6y4va4uph0x"

var networkInfos = map[string]NetworkConfig{
	"fetchhub-4": {
		NewChainID:     "asi-1",
		NewDescription: "ASI token", // TODO(JS): confirm this
		DenomInfo: DenomInfo{
			NewBaseDenom: "asi",
			NewDenom:     "aasi",
			OldDenom:     "afet",
		},
		IbcTargetAddr:            "fetch1rhrlzsx9z865dqen8t4v47r99dw6y4va4uph0x", // TODO(JS): amend this
		ReconciliationTargetAddr: &ReconciliationTargetAddr,                      // TODO(JS): amend this
		Contracts: &Contracts{
			MobixStaking: &MobixStaking{
				Addr: "fetch1xr3rq8yvd7qplsw5yx90ftsr2zdhg4e9z60h5duusgxpv72hud3szdul6e", // TODO(JS): amend this
			},
			TokenBridge: &TokenBridge{
				Addr:     "fetch1qxxlalvsdjd07p07y3rc5fu6ll8k4tmetpha8n",
				NewAdmin: "fetch15p3rl5aavw9rtu86tna5lgxfkz67zzr6ed4yhw",
			},
		},
	},

	"dorado-1": {
		NewChainID:     "asi-1",          // TODO(JS): likely amend this
		NewDescription: "Test ASI token", // TODO(JS): confirm this
		DenomInfo: DenomInfo{
			NewBaseDenom: "testasi",
			NewDenom:     "atestasi",
			OldDenom:     "atestfet",
		},
		IbcTargetAddr: "fetch1rhrlzsx9z865dqen8t4v47r99dw6y4va4uph0x", // TODO(JS): amend this
		Contracts: &Contracts{
			MobixStaking: &MobixStaking{
				Addr: "fetch1xr3rq8yvd7qplsw5yx90ftsr2zdhg4e9z60h5duusgxpv72hud3szdul6e",
			},
		},
	},
}

//go:embed reconciliation_data.csv
var reconciliationData []byte

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
              - The old fetch addresses will be updated to the new asi addresses
              - The bridge contract admin will be updated to the new address
              - The IBC withdrawal address will be updated to the new address
              - The reconciliation withdrawal address will be updated to the new address
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

			var ok bool
			var networkConfig NetworkConfig // TODO(JS): potentially just read Chain-ID, instead of taking a new arg
			if networkConfig, ok = networkInfos[genDoc.ChainID]; !ok {
				return fmt.Errorf("network not found, not match for Chain-ID")
			}

			var jsonData map[string]interface{}
			if err = json.Unmarshal(genDoc.AppState, &jsonData); err != nil {
				return fmt.Errorf("failed to unmarshal app state: %w", err)
			}

			// replace chain-id
			ASIGenesisUpgradeReplaceChainID(genDoc, networkConfig)

			// replace bridge contract admin, if address and new admin present
			if networkConfig.Contracts != nil && networkConfig.Contracts.TokenBridge != nil {
				ASIGenesisUpgradeReplaceBridgeAdmin(jsonData, networkConfig)
			}

			// update mobix staking contract, if address present
			if networkConfig.Contracts != nil && networkConfig.Contracts.MobixStaking != nil {
				ASIGenesisUpgradeUpdateMobixStakingContract(jsonData, networkConfig)
			}

			// withdraw balances from IBC channels
			if err = ASIGenesisUpgradeWithdrawIBCChannelsBalances(jsonData, networkConfig); err != nil {
				return err
			}

			// withdraw balances from reconciliation addresses
			if networkConfig.ReconciliationTargetAddr != nil {
				if err = ASIGenesisUpgradeWithdrawReconciliationBalances(jsonData, networkConfig); err != nil {
					return err
				}
			}

			// set denom metadata in bank module
			ASIGenesisUpgradeReplaceDenomMetadata(jsonData, networkConfig)

			// replace denom across the genesis file
			ASIGenesisUpgradeReplaceDenom(jsonData, networkConfig)

			// replace addresses across the genesis file
			ASIGenesisUpgradeReplaceAddresses(jsonData, networkConfig)

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

func ASIGenesisUpgradeUpdateMobixStakingContract(jsonData map[string]interface{}, networkInfo NetworkConfig) {
	contracts := jsonData["wasm"].(map[string]interface{})["contracts"].([]interface{})
	MobixStakingContractAddress := networkInfo.Contracts.MobixStaking.Addr

	re := regexp.MustCompile(fmt.Sprintf(`%s%s1([%s]{%d})$`, OldAddrPrefix, "", Bech32Chars, AddrDataLength+AddrChecksumLength))

	for _, contract := range contracts {
		if contract.(map[string]interface{})["contract_address"] == MobixStakingContractAddress {
			mobixContractStates := contract.(map[string]interface{})["contract_state"].([]interface{})
			for _, val := range mobixContractStates {
				state := val.(map[string]interface{})
				hexKey := state["key"].(string)
				b64Value := state["value"].(string)

				keyBytes, err := hex.DecodeString(hexKey)
				if err != nil {
					panic(err)
				}

				valueBytes, err := base64.StdEncoding.DecodeString(b64Value)
				if err != nil {
					panic(err)
				}

				val = replaceContractState(re, string(keyBytes), string(valueBytes))
			}

			return
		}
	}

	panic("mobix staking contract not found")
}

func replaceContractState(re *regexp.Regexp, key string, value string) map[string]interface{} {
	var newKey []byte
	var newValue []byte

	// replace key
	newKeyStr := re.ReplaceAllStringFunc(key, func(match string) string {
		newAddr, err := convertAddressToASI(match, AccAddressPrefix)
		if err != nil {
			panic(err)
		}
		return newAddr
	})
	newKey = []byte(newKeyStr)

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

	// return reconstructed contract state
	return map[string]interface{}{
		"key":   hex.EncodeToString(newKey),
		"value": base64.StdEncoding.EncodeToString(newValue),
	}
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
	contracts := jsonData["wasm"].(map[string]interface{})["contracts"].([]interface{})

	for i, contract := range contracts {
		c := contract.(map[string]interface{})
		if c["contract_address"] == networkInfo.Contracts.TokenBridge.Addr {
			contractInfo := c["contract_info"].(map[string]interface{})
			contractInfo["admin"] = networkInfo.Contracts.TokenBridge.NewAdmin
			contracts[i] = c
			break
		}
	}
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

func ASIGenesisUpgradeWithdrawIBCChannelsBalances(jsonData map[string]interface{}, networkInfo NetworkConfig) error {
	bank := jsonData[banktypes.ModuleName].(map[string]interface{})
	balances := bank["balances"].([]interface{})
	balanceMap := getGenesisBalancesMap(balances)
	ibcWithdrawalAddress := networkInfo.IbcTargetAddr

	withdrawalBalanceIdx, ok := (*balanceMap)[ibcWithdrawalAddress]
	if !ok {
		fmt.Println("failed to find Ibc withdrawal address in genesis balances - have addresses already been converted?")
		return nil
	}

	ibc := jsonData[ibccore.ModuleName].(map[string]interface{})
	channelGenesis := ibc["channel_genesis"].(map[string]interface{})
	ibcChannels := channelGenesis["channels"].([]interface{})

	for _, channel := range ibcChannels {
		channelId := channel.(map[string]interface{})["channel_id"].(string)
		portId := channel.(map[string]interface{})["port_id"].(string)

		rawAddr := ibctransfertypes.GetEscrowAddress(portId, channelId)
		channelAddr, err := sdk.Bech32ifyAddressBytes(OldAddrPrefix+AccAddressPrefix, rawAddr)
		if err != nil {
			return fmt.Errorf("failed to bech32ify address: %w", err)
		}

		balanceIdx, ok := (*balanceMap)[channelAddr]
		if !ok {
			// channel address not found in genesis balances
			continue
		}

		channelBalanceCoins := getCoinsFromInterfaceSlice(balances[balanceIdx])
		withdrawalBalanceCoins := getCoinsFromInterfaceSlice(balances[withdrawalBalanceIdx])

		// add channel balance to withdrawal balance
		newWithdrawalBalanceCoins := withdrawalBalanceCoins.Add(channelBalanceCoins...)
		balances[withdrawalBalanceIdx].(map[string]interface{})["coins"] = getInterfaceSliceFromCoins(newWithdrawalBalanceCoins)

		// zero out the channel balance
		balances[balanceIdx].(map[string]interface{})["coins"] = []interface{}{}
	}

	return nil
}

func getGenesisAccountSequenceMap(accounts []interface{}) *map[string]int {
	const ModuleAccount = "/cosmos.auth.v1beta1.ModuleAccount"
	accountMap := make(map[string]int)

	for _, acc := range accounts {
		accType := acc.(map[string]interface{})["@type"]

		accData := acc
		if accType == ModuleAccount {
			accData = acc.(map[string]interface{})["base_account"]
		}

		addr := accData.(map[string]interface{})["address"].(string)
		sequence := accData.(map[string]interface{})["sequence"].(string)

		sequenceInt, ok := strconv.Atoi(sequence)
		if ok != nil {
			panic("getGenesisAccountSequenceMap: failed to convert sequence to int")
		}
		accountMap[addr] = sequenceInt
	}

	return &accountMap
}

func ASIGenesisUpgradeWithdrawReconciliationBalances(jsonData map[string]interface{}, networkInfo NetworkConfig) error {
	bank := jsonData[banktypes.ModuleName].(map[string]interface{})
	balances := bank["balances"].([]interface{})
	reconciliationWithdrawAddress := networkInfo.ReconciliationTargetAddr

	balanceMap := getGenesisBalancesMap(balances)

	auth := jsonData[authtypes.ModuleName].(map[string]interface{})
	accounts := auth["accounts"].([]interface{})
	accountSequenceMap := getGenesisAccountSequenceMap(accounts)

	fileData := reconciliationData
	r := csv.NewReader(bytes.NewReader(fileData))
	items, err := r.ReadAll()
	if err != nil {
		log.Fatalf("Error reading reconciliation data: %s", err)
	}

	reconciliationBalanceIdx, ok := (*balanceMap)[*reconciliationWithdrawAddress]
	if !ok {
		return fmt.Errorf("no match in genesis for reconciliation address: %s", *reconciliationWithdrawAddress)
	}

	for _, row := range items {
		addr := row[2]

		//_ = row[3] balance from CSV

		accSequence, ok := (*accountSequenceMap)[addr]
		if !ok {
			return fmt.Errorf("no match in genesis for reconciliation address: %s", addr)
		}

		balanceIdx, ok := (*balanceMap)[addr]
		if !ok {
			continue
		}

		accBalance := balances[balanceIdx]
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
	}
	return nil
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
			panic("IBC withdraw: failed to convert coin amount to int")
		}
		balanceCoins = append(balanceCoins, sdk.NewCoin(coinDenom, coinAmount))
	}
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

type NetworkConfig struct {
	NewChainID               string
	NewDescription           string
	IbcTargetAddr            string
	ReconciliationTargetAddr *string
	Contracts                *Contracts
	DenomInfo                DenomInfo
}

type DenomInfo struct {
	NewBaseDenom string
	NewDenom     string
	OldDenom     string
}

type Contracts struct {
	TokenBridge  *TokenBridge
	MobixStaking *MobixStaking
}

type TokenBridge struct {
	Addr     string
	NewAdmin string
}

type MobixStaking struct {
	Addr string
}
