package app

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	wasmTypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	"github.com/cosmos/cosmos-sdk/types"
	"os"
	"path"
	"strings"
)

//go:embed reconciliation_data.csv
var reconciliationData []byte

//go:embed reconciliation_data_testnet.csv
var reconciliationDataTestnet []byte

var reconciliationBalancesKey = prefixStringWithLength("balances")

const manifestFilenameBase = "upgrade_manifest.json"

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
		ReconciliationInfo: &ReconciliationInfo{
			TargetAddress:   "fetch1g5ur2wc5xnlc7sw9wd895lw7mmxz04r5syj3s6ew8md6pvwuweqqavkgt0",
			InputCSVRecords: readInputReconciliationData(reconciliationDataTestnet),
		},
		Contracts: &ContractSet{
			Reconciliation: &Reconciliation{
				Addr: "fetch1g5ur2wc5xnlc7sw9wd895lw7mmxz04r5syj3s6ew8md6pvwuweqqavkgt0",
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

func getStringPtr(val string) *string {
	return &val
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

func readInputReconciliationData(csvData []byte) [][]string {
	r := csv.NewReader(bytes.NewReader(csvData))
	records, err := r.ReadAll()
	if err != nil {
		panic(fmt.Sprintf("error reading reconciliation data: %v", err))
	}
	return records
}

func (app *App) ChangeContractLabel(ctx types.Context, contractAddr *string, newLabel *string, manifest *UpgradeManifest) error {
	addr, store, _, err := app.getContractData(ctx, *contractAddr)
	if err != nil {
		return err
	}

	// Get contract info
	var contractInfo = app.WasmKeeper.GetContractInfo(ctx, *addr)
	oldLabel := contractInfo.Label
	contractInfo.Label = *newLabel

	// Store contract info
	contractBz, err := app.AppCodec().Marshal(contractInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal updated contract info: %v", err)
	}

	contractAddrKey := append(wasmTypes.ContractKeyPrefix, *addr...)
	(*store).Set(contractAddrKey, contractBz)

	manifest.Contracts.LabelUpdated = append(manifest.Contracts.LabelUpdated, ContractValueUpdate{*contractAddr, oldLabel, *newLabel})
	return nil
}

func (app *App) ChangeContractLabels(ctx types.Context, networkInfo *NetworkConfig, manifest *UpgradeManifest) error {
	contracts := []struct{ addr, newLabel *string }{
		{addr: &networkInfo.Contracts.Reconciliation.Addr, newLabel: networkInfo.Contracts.Reconciliation.NewLabel},
	}
	for _, contract := range contracts {
		err := app.ChangeContractLabel(ctx, contract.addr, contract.newLabel, manifest)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *App) ProcessReconciliation(ctx types.Context, networkInfo *NetworkConfig, manifest *UpgradeManifest) error {
	records := networkInfo.ReconciliationInfo.InputCSVRecords

	err := app.WithdrawReconciliationBalances(ctx, networkInfo, records, manifest)
	if err != nil {
		return fmt.Errorf("error withdrawing reconciliation balances: %v", err)
	}

	err = app.ReplaceReconciliationContractState(ctx, networkInfo, manifest)
	if err != nil {
		return fmt.Errorf("error replacing reconciliation contract state: %v", err)
	}

	return nil
}

func (app *App) WithdrawReconciliationBalances(ctx types.Context, networkInfo *NetworkConfig, records [][]string, manifest *UpgradeManifest) error {
	landingAddr, err := types.AccAddressFromBech32(networkInfo.ReconciliationInfo.TargetAddress)
	if err != nil {
		return err
	}

	if !app.AccountKeeper.HasAccount(ctx, landingAddr) {
		return fmt.Errorf("landing address does not exist")
	}

	for _, record := range records {
		recordAddr, err := types.AccAddressFromBech32(record[2])
		if err != nil {
			return err
		}

		if !app.AccountKeeper.HasAccount(ctx, recordAddr) {
			continue
		}

		recordAccount := app.AccountKeeper.GetAccount(ctx, recordAddr)
		recordBalanceCoins := app.BankKeeper.GetAllBalances(ctx, recordAddr)
		if !recordBalanceCoins.IsAllPositive() || recordAccount.GetSequence() != 0 {
			continue
		}

		err = app.BankKeeper.SendCoins(ctx, recordAddr, landingAddr, recordBalanceCoins)
		if err != nil {
			return err
		}

		transfer := UpgradeReconciliationTransfer{
			EthAddr: record[0],
			From:    record[2],
			Amount:  recordBalanceCoins,
		}
		manifest.Reconciliation.Transfers.Transfers = append(manifest.Reconciliation.Transfers.Transfers, transfer)
	}

	return nil
}

func (app *App) ReplaceReconciliationContractState(ctx types.Context, networkInfo *NetworkConfig, manifest *UpgradeManifest) error {
	reconciliationTransfers := manifest.Reconciliation.Transfers.Transfers
	_, _, prefixStore, err := app.getContractData(ctx, networkInfo.Contracts.Reconciliation.Addr)
	if err != nil {
		return err
	}

	for _, transfer := range reconciliationTransfers {
		key, value := reconciliationContractStateBalancesRecord(transfer.EthAddr, transfer.Amount)
		if key == nil {
			continue
		}
		manifest.Reconciliation.ContractState.Balances = append(manifest.Reconciliation.ContractState.Balances, UpgradeReconciliationContractStateBalanceRecord{EthAddr: transfer.EthAddr, Balances: transfer.Amount})
		manifest.Reconciliation.ContractState.AggregatedBalancesAmount = manifest.Reconciliation.ContractState.AggregatedBalancesAmount.Add(transfer.Amount...)
		manifest.Reconciliation.ContractState.NumberOfBalanceRecords += 1

		prefixStore.Set(key, value)
	}

	return nil
}

func reconciliationContractStateBalancesRecord(ethAddrHex string, coins types.Coins) ([]byte, []byte) {
	resCoins := types.Coins{}
	for _, coin := range coins {
		if coin.IsPositive() {
			resCoins = resCoins.Add(coin)
		}
	}

	if resCoins.Empty() {
		return nil, nil
	}

	resCoins.Sort()

	ethAddrHexNoPrefix := DropHexPrefix(ethAddrHex)

	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)

	if _, err := writer.Write(reconciliationBalancesKey); err != nil {
		panic(err)
	}

	if ethAddrRaw, err := hex.DecodeString(ethAddrHexNoPrefix); err != nil {
		panic(err)
	} else {
		if _, err := writer.Write(ethAddrRaw); err != nil {
			panic(err)
		}
	}

	if err := writer.Flush(); err != nil {
		panic(err)
	}

	if value, err := resCoins.MarshalJSON(); err != nil {
		panic(err)
	} else {
		return buffer.Bytes(), value
	}
}

func DropHexPrefix(hexEncodedData string) string {
	strLen := len(hexEncodedData)
	if strLen < 1 {
		return hexEncodedData
	}

	prefixEstimateLen := 1
	if strLen > 1 {
		prefixEstimateLen = 2
	}

	prefixEstimate := strings.ToLower(hexEncodedData[:prefixEstimateLen])

	if strings.HasPrefix(prefixEstimate, "0x") {
		return hexEncodedData[2:]
	} else if strings.HasPrefix(prefixEstimate, "x") {
		return hexEncodedData[1:]
	}

	return hexEncodedData
}

func (app *App) UpgradeContractAdmin(ctx types.Context, newAdmin *string, contractAddr *string, manifest *UpgradeManifest) error {
	if newAdmin == nil || contractAddr == nil {
		return nil
	}

	addr, store, _, err := app.getContractData(ctx, *contractAddr)
	if err != nil {
		return err
	}

	// Get contract info
	var contract = app.WasmKeeper.GetContractInfo(ctx, *addr)
	oldAdmin := contract.Admin
	contract.Admin = *newAdmin

	// Store contract info
	contractBz, err := app.AppCodec().Marshal(contract)
	if err != nil {
		return fmt.Errorf("failed to marshal updated contract info: %v", err)
	}

	contractAddrKey := append(wasmTypes.ContractKeyPrefix, *addr...)
	(*store).Set(contractAddrKey, contractBz)

	manifest.Contracts.AdminUpdated = append(manifest.Contracts.AdminUpdated, ContractValueUpdate{*contractAddr, oldAdmin, *newAdmin})
	return nil
}

func (app *App) UpgradeContractAdmins(ctx types.Context, networkInfo *NetworkConfig, manifest *UpgradeManifest) error {
	contracts := []struct{ Addr, NewAdmin *string }{
		{Addr: &networkInfo.Contracts.Reconciliation.Addr, NewAdmin: networkInfo.Contracts.Reconciliation.NewAdmin},
		{Addr: &networkInfo.Contracts.TokenBridge.Addr, NewAdmin: networkInfo.Contracts.TokenBridge.NewAdmin},
	}

	for _, contract := range contracts {
		err := app.UpgradeContractAdmin(ctx, contract.NewAdmin, contract.Addr, manifest)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *App) DeleteContractState(ctx types.Context, contractAddr string, manifest *UpgradeManifest) error {
	if contractAddr == "" {
		return nil
	}

	_, _, prefixStore, err := app.getContractData(ctx, contractAddr)
	if err != nil {
		return err
	}

	iter := prefixStore.Iterator(nil, nil)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		prefixStore.Delete(iter.Key())
	}
	manifest.Contracts.StateCleaned = append(manifest.Contracts.StateCleaned, contractAddr)
	return nil
}

func (app *App) DeleteContractStates(ctx types.Context, networkInfo *NetworkConfig, manifest *UpgradeManifest) error {
	contractsToWipe := []string{
		networkInfo.Contracts.Reconciliation.Addr,
		networkInfo.Contracts.Almanac.ProdAddr,
		networkInfo.Contracts.Almanac.DevAddr,
		networkInfo.Contracts.AName.DevAddr,
		networkInfo.Contracts.AName.ProdAddr,
	}

	for _, contract := range contractsToWipe {
		err := app.DeleteContractState(ctx, contract, manifest)
		if err != nil {
			return err
		}
	}

	return nil
}

// getContractData returns the contract address, info, and states for a given contract address
func (app *App) getContractData(ctx types.Context, contractAddr string) (*types.AccAddress, *types.KVStore, *prefix.Store, error) {
	addr, err := types.AccAddressFromBech32(contractAddr)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("invalid contract address: %v", err)
	}

	store := ctx.KVStore(app.keys[wasmTypes.StoreKey])
	contractAddrKey := wasmTypes.GetContractStorePrefix(addr)
	prefixStore := prefix.NewStore(store, contractAddrKey)

	return &addr, &store, &prefixStore, nil
}

type UpgradeReconciliationTransfer struct {
	From    string      `json:"from"`
	EthAddr string      `json:"eth_addr"`
	Amount  types.Coins `json:"amount"`
}

type NetworkConfig struct {
	ReconciliationInfo *ReconciliationInfo
	Contracts          *ContractSet
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

type Reconciliation struct {
	Addr     string
	NewAdmin *string
	NewLabel *string
}

type Almanac struct {
	DevAddr  string
	ProdAddr string
}

type AName struct {
	DevAddr  string
	ProdAddr string
}

type UpgradeManifest struct {
	Reconciliation *UpgradeReconciliation `json:"reconciliation,omitempty"`
	Contracts      *Contracts             `json:"contracts,omitempty"`
}

func initManifest() *UpgradeManifest {
	return &UpgradeManifest{
		Reconciliation: &UpgradeReconciliation{
			Transfers: UpgradeReconciliationTransfers{
				Transfers: make([]UpgradeReconciliationTransfer, 0),
			},
			ContractState: &UpgradeReconciliationContractState{
				Balances: make([]UpgradeReconciliationContractStateBalanceRecord, 0),
			},
		},
		Contracts: &Contracts{
			StateCleaned: make([]string, 0),
			AdminUpdated: make([]ContractValueUpdate, 0),
			LabelUpdated: make([]ContractValueUpdate, 0),
		},
	}
}

type Contracts struct {
	StateCleaned []string              `json:"contracts_state_cleaned,omitempty"`
	AdminUpdated []ContractValueUpdate `json:"contracts_admin_updated,omitempty"`
	LabelUpdated []ContractValueUpdate `json:"contracts_label_updated,omitempty"`
}

type ContractValueUpdate struct {
	Address string `json:"address"`
	From    string `json:"from"`
	To      string `json:"to"`
}

type ValueUpdate struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type UpgradeReconciliation struct {
	Transfers     UpgradeReconciliationTransfers      `json:"transfers"`
	ContractState *UpgradeReconciliationContractState `json:"contract_state"`
}

type UpgradeReconciliationTransfers struct {
	Transfers                   []UpgradeReconciliationTransfer `json:"transfers"`
	To                          string                          `json:"to"`
	AggregatedTransferredAmount types.Coins                     `json:"aggregated_transferred_amount"`
	NumberOfTransfers           int                             `json:"number_of_transfers"`
}

type UpgradeReconciliationContractStateBalanceRecord struct {
	EthAddr  string      `json:"eth_addr"`
	Balances types.Coins `json:"balances"`
}

type UpgradeReconciliationContractState struct {
	Balances                 []UpgradeReconciliationContractStateBalanceRecord `json:"balances"`
	AggregatedBalancesAmount types.Coins                                       `json:"aggregated_balances_amount"`
	NumberOfBalanceRecords   int                                               `json:"number_of_balance_records"`
}

func (app *App) GetManifestFilePath(prefix string) (string, error) {
	var upgradeFilePath string
	var err error

	if upgradeFilePath, err = app.UpgradeKeeper.GetUpgradeInfoPath(); err != nil {
		return "", err
	}

	upgradeDir := path.Dir(upgradeFilePath)

	manifestFileName := manifestFilenameBase
	if prefix != "" {
		manifestFileName = fmt.Sprintf("%s_%s", prefix, manifestFilenameBase)
	}

	manifestFilePath := path.Join(upgradeDir, manifestFileName)

	return manifestFilePath, nil
}

func (app *App) SaveManifest(manifest *UpgradeManifest, upgradeLabel string) error {
	var serialisedManifest []byte
	var err error

	var manifestFilePath string
	if manifestFilePath, err = app.GetManifestFilePath(upgradeLabel); err != nil {
		return err
	}

	if serialisedManifest, err = json.MarshalIndent(manifest, "", "\t"); err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	var f *os.File
	if f, err = os.Create(manifestFilePath); err != nil {
		return fmt.Errorf("failed to create file \"%s\": %w", manifestFilePath, err)
	}
	defer f.Close()

	if _, err = f.Write(serialisedManifest); err != nil {
		return fmt.Errorf("failed to write manifest to the \"%s\" file : %w", manifestFilePath, err)
	}

	return nil
}
