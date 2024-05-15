package cmd

import (
	"bytes"
	_ "embed"
	"encoding/csv"
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
	BridgeContractAddress  = "fetch1qxxlalvsdjd07p07y3rc5fu6ll8k4tmetpha8n"
	NewBridgeContractAdmin = "fetch15p3rl5aavw9rtu86tna5lgxfkz67zzr6ed4yhw"

	IbcWithdrawAddress            = "fetch1rhrlzsx9z865dqen8t4v47r99dw6y4va4uph0x" /* "asi1rhrlzsx9z865dqen8t4v47r99dw6y4vaw76rd9" */
	ReconciliationWithdrawAddress = "fetch1rhrlzsx9z865dqen8t4v47r99dw6y4va4uph0x"

	Bech32Chars        = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"
	AddrDataLength     = 32
	WasmAddrDataLength = 52
	AddrChecksumLength = 6

	AccAddressPrefix  = ""
	ValAddressPrefix  = "valoper"
	ConsAddressPrefix = "valcons"

	NewBaseDenom   = "asi"
	NewDenom       = "aasi"
	NewAddrPrefix  = "asi"
	NewChainId     = "asi-1"
	NewDescription = "ASI Token" // TODO(JS): change this, potentially

	OldDenom      = "afet"
	OldAddrPrefix = "fetch"
)

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

			var jsonData map[string]interface{}
			if err = json.Unmarshal(genDoc.AppState, &jsonData); err != nil {
				return fmt.Errorf("failed to unmarshal app state: %w", err)
			}

			// replace chain-id
			ASIGenesisUpgradeReplaceChainID(genDoc)

			// replace bridge contract admin
			ASIGenesisUpgradeReplaceBridgeAdmin(jsonData)

			manifest := ASIUpgradeManifest{}

			// withdraw balances from IBC channels
			if err = ASIGenesisUpgradeWithdrawIBCChannelsBalances(jsonData, &manifest); err != nil {
				return err
			}

			// withdraw balances from reconciliation addresses
			if err = ASIGenesisUpgradeWithdrawReconciliationBalances(jsonData, &manifest); err != nil {
				return err
			}

			// set denom metadata in bank module
			ASIGenesisUpgradeReplaceDenomMetadata(jsonData)

			// replace denom across the genesis file
			ASIGenesisUpgradeReplaceDenom(jsonData)

			// replace addresses across the genesis file
			ASIGenesisUpgradeReplaceAddresses(jsonData)

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
	cmd.Flags().String(flags.FlagKeyringBackend, flags.DefaultKeyringBackend, "Select keyring's backend (os|file|kwallet|pass|test)")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func ASIGenesisUpgradeReplaceDenomMetadata(jsonData map[string]interface{}) {
	type jsonMap map[string]interface{}

	NewBaseDenomUpper := strings.ToUpper(NewBaseDenom)

	newMetadata := jsonMap{
		"base": NewDenom,
		"denom_units": []jsonMap{
			{
				"denom":    NewBaseDenomUpper,
				"exponent": 18,
			},
			{
				"denom":    fmt.Sprintf("m%s", NewBaseDenom),
				"exponent": 15,
			},
			{
				"denom":    fmt.Sprintf("u%s", NewBaseDenom),
				"exponent": 12,
			},
			{
				"denom":    fmt.Sprintf("n%s", NewBaseDenom),
				"exponent": 9,
			},
			{
				"denom":    fmt.Sprintf("p%s", NewBaseDenom),
				"exponent": 6,
			},
			{
				"denom":    fmt.Sprintf("f%s", NewBaseDenom),
				"exponent": 3,
			},
			{
				"denom":    fmt.Sprintf("a%s", NewBaseDenom),
				"exponent": 0,
			},
		},
		"description": NewDescription,
		"display":     NewBaseDenomUpper,
		"name":        NewBaseDenomUpper,
		"symbol":      NewBaseDenomUpper,
	}

	bank := jsonData["bank"].(map[string]interface{})
	denomMetadata := bank["denom_metadata"].([]interface{})

	for i, metadata := range denomMetadata {
		denomUnit := metadata.(map[string]interface{})
		if denomUnit["base"] == OldDenom {
			denomMetadata[i] = newMetadata
			break
		}
	}
}

func ASIGenesisUpgradeReplaceChainID(genesisData *types.GenesisDoc) {
	genesisData.ChainID = NewChainId
}

func ASIGenesisUpgradeReplaceBridgeAdmin(jsonData map[string]interface{}) {
	contracts := jsonData["wasm"].(map[string]interface{})["contracts"].([]interface{})

	for i, contract := range contracts {
		c := contract.(map[string]interface{})
		if c["contract_address"] == BridgeContractAddress {
			contractInfo := c["contract_info"].(map[string]interface{})
			contractInfo["admin"] = NewBridgeContractAdmin
			contracts[i] = c
			break
		}
	}
}

func ASIGenesisUpgradeReplaceDenom(jsonData map[string]interface{}) {
	targets := map[string]struct{}{"denom": {}, "bond_denom": {}, "mint_denom": {}, "base_denom": {}, "base": {}}

	crawlJson("", jsonData, -1, func(key string, value interface{}, idx int) interface{} {
		if str, ok := value.(string); ok {
			_, isInTargets := targets[key]
			if str == OldDenom && isInTargets {
				return NewDenom
			}
		}
		return value
	})
}

func ASIGenesisUpgradeReplaceAddresses(jsonData map[string]interface{}) {
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

func ASIGenesisUpgradeWithdrawIBCChannelsBalances(jsonData map[string]interface{}, manifest *ASIUpgradeManifest) error {
	bank := jsonData[banktypes.ModuleName].(map[string]interface{})
	balances := bank["balances"].([]interface{})
	balanceMap := getGenesisBalancesMap(balances)

	manifest.IBC = &ASIUpgradeTransfers{
		Transfer: []ASIUpgradeTransfer{},
		To:       IbcWithdrawAddress,
	}

	withdrawalBalanceIdx, ok := (*balanceMap)[IbcWithdrawAddress]
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

		manifest.IBC.Transfer = append(manifest.IBC.Transfer, ASIUpgradeTransfer{From: channelAddr, Amount: channelBalanceCoins})

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

func ASIGenesisUpgradeWithdrawReconciliationBalances(jsonData map[string]interface{}, manifest *ASIUpgradeManifest) error {
	bank := jsonData[banktypes.ModuleName].(map[string]interface{})
	balances := bank["balances"].([]interface{})

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

	reconciliationBalanceIdx, ok := (*balanceMap)[ReconciliationWithdrawAddress]
	if !ok {
		return fmt.Errorf("no genesis match for reconciliation address: %s", ReconciliationWithdrawAddress)
	}

	manifest.Reconciliation = &ASIUpgradeTransfers{
		Transfer: []ASIUpgradeTransfer{},
		To:       ReconciliationWithdrawAddress,
	}

	for _, row := range items {
		addr := row[2]

		//_ = row[3] balance from CSV

		accSequence, ok := (*accountSequenceMap)[addr]
		if !ok {
			return fmt.Errorf("no genesis match for reconciliation address: %s", addr)
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

		manifest.Reconciliation.Transfer = append(manifest.Reconciliation.Transfer, ASIUpgradeTransfer{From: addr, Amount: accBalanceCoins})

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
