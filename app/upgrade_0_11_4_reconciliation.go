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
	"strings"
)

//go:embed reconciliation_data.csv
var reconciliationData []byte

//go:embed reconciliation_data_testnet.csv
var reconciliationDataTestnet []byte

var (
	cw2contractInfoKey                     = []byte("contract_info")
	reconciliationTotalBalanceKey          = []byte("total_balance")
	reconciliationNOutstandingAddressesKey = []byte("n_outstanding_addresses")
	reconciliationStateKey                 = []byte("state")
	reconciliationBalancesKey              = prefixStringWithLength("balances")
)

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

func readInputReconciliationData(csvData []byte) *[][]string {
	r := csv.NewReader(bytes.NewReader(csvData))
	records, err := r.ReadAll()
	if err != nil {
		panic(fmt.Sprintf("error reading reconciliation data: %v", err))
	}
	return &records
}

func (app *App) ChangeContractLabel(ctx types.Context, contractAddr *string, newLabel *string, manifest *UpgradeManifest) error {
	// No label to update
	if newLabel == nil || contractAddr == nil {
		return nil
	}

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

	if manifest.Contracts == nil {
		manifest.Contracts = new(Contracts)
	}

	manifest.Contracts.LabelUpdated = append(manifest.Contracts.LabelUpdated, ContractValueUpdate{*contractAddr, oldLabel, *newLabel})

	return nil
}

func (app *App) ChangeContractVersion(ctx types.Context, contractAddr *string, newVersion *ContractVersion, manifest *UpgradeManifest) error {
	if contractAddr == nil || newVersion == nil {
		return nil
	}

	_, _, prefixStore, err := app.getContractData(ctx, *contractAddr)
	if err != nil {
		return err
	}

	wasPresent := prefixStore.Has(cw2contractInfoKey)

	var origVersion *CW2ContractVersion
	if wasPresent {
		storeVal := prefixStore.Get(cw2contractInfoKey)
		var val CW2ContractVersion
		if err := json.Unmarshal(storeVal, &val); err != nil {
			return err
		}
		origVersion = &val
	}

	if newVersion.CW2version != nil {
		newVersionStoreValue, err := json.Marshal(*newVersion.CW2version)
		if err != nil {
			return err
		}
		prefixStore.Set(cw2contractInfoKey, newVersionStoreValue)
	} else if wasPresent {
		prefixStore.Delete(cw2contractInfoKey)
	}

	manifestVersionUpdate := ContractVersionUpdate{
		Address: *contractAddr,
		From:    origVersion,
		To:      newVersion.CW2version,
	}

	if manifest.Contracts == nil {
		manifest.Contracts = new(Contracts)
	}

	manifest.Contracts.VersionUpdated = append(manifest.Contracts.VersionUpdated, manifestVersionUpdate)

	return nil
}

func (app *App) ChangeContractLabels(ctx types.Context, networkInfo *NetworkConfig, manifest *UpgradeManifest) error {
	contracts := []IContractLabel{networkInfo.Contracts.Reconciliation}
	for _, contract := range contracts {
		if contract == nil {
			continue
		}
		err := app.ChangeContractLabel(ctx, contract.GetPrimaryContractAddr(), contract.GetNewLabel(), manifest)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *App) ChangeContractVersions(ctx types.Context, networkInfo *NetworkConfig, manifest *UpgradeManifest) error {
	contracts := []IContractVersion{networkInfo.Contracts.Reconciliation}
	for _, contract := range contracts {
		if contract == nil {
			continue
		}
		err := app.ChangeContractVersion(ctx, contract.GetPrimaryContractAddr(), contract.GetNewVersion(), manifest)
		if err != nil {
			return err
		}
	}

	return nil
}

func (app *App) ProcessReconciliation(ctx types.Context, networkInfo *NetworkConfig, manifest *UpgradeManifest) error {
	records := networkInfo.ReconciliationInfo.InputCSVRecords

	if records == nil {
		return nil
	}

	err := app.WithdrawReconciliationBalances(ctx, networkInfo, *records, manifest)
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

	transfers := UpgradeReconciliationTransfers{}

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
		transfers.Transfers = append(transfers.Transfers, transfer)
		transfers.AggregatedTransferredAmount = transfers.AggregatedTransferredAmount.Add(recordBalanceCoins...)
		transfers.NumberOfTransfers = len(transfers.Transfers)
	}

	if transfers.NumberOfTransfers > 0 {
		transfers.To = networkInfo.ReconciliationInfo.TargetAddress

		if manifest.Reconciliation == nil {
			manifest.Reconciliation = &UpgradeReconciliation{}
		}
		manifest.Reconciliation.Transfers = &transfers
	} else {
		if !transfers.AggregatedTransferredAmount.IsZero() {
			return fmt.Errorf("manifest: Transfers: `NumberOfTransfers` is zero but `AggregatedTransferredAmount` is not zero")
		}

		if manifest.Reconciliation != nil {
			manifest.Reconciliation.Transfers = nil
		}
	}

	return nil
}

func (app *App) ReplaceReconciliationContractState(ctx types.Context, networkInfo *NetworkConfig, manifest *UpgradeManifest) error {
	_, _, prefixStore, err := app.getContractData(ctx, networkInfo.Contracts.Reconciliation.Addr)
	if err != nil {
		return err
	}

	manifest.Reconciliation.ContractState = nil
	contractState := UpgradeReconciliationContractState{}

	for _, transfer := range manifest.Reconciliation.Transfers.Transfers {
		key, value := reconciliationContractStateBalancesRecord(transfer.EthAddr, transfer.Amount)
		if key == nil {
			continue
		}
		contractState.Balances = append(contractState.Balances, UpgradeReconciliationContractStateBalanceRecord{EthAddr: transfer.EthAddr, Balances: transfer.Amount})
		contractState.AggregatedBalancesAmount = contractState.AggregatedBalancesAmount.Add(transfer.Amount...)
		contractState.NumberOfBalanceRecords = len(contractState.Balances)

		prefixStore.Set(key, value)
	}

	totalBalanceRecord, err := contractState.AggregatedBalancesAmount.MarshalJSON()
	if err != nil {
		return err
	}

	stateRecord := ReconciliationContractStateRecord{
		Paused: true,
	}

	var stateRecordJsonBz []byte
	stateRecordJsonBz, err = json.Marshal(stateRecord)
	if err != nil {
		return err
	}

	prefixStore.Set(reconciliationTotalBalanceKey, totalBalanceRecord)
	prefixStore.Set(reconciliationNOutstandingAddressesKey, []byte(fmt.Sprintf("%d", contractState.NumberOfBalanceRecords)))
	prefixStore.Set(reconciliationStateKey, stateRecordJsonBz)

	if contractState.NumberOfBalanceRecords != len(contractState.Balances) {
		return fmt.Errorf("manifest: ContractState: number of elements in the `Balances` array does not match the `NumberOfBalanceRecords`")
	}

	if contractState.NumberOfBalanceRecords > 0 {
		if manifest.Reconciliation == nil {
			manifest.Reconciliation = &UpgradeReconciliation{}
		}

		manifest.Reconciliation.ContractState = &contractState
	} else {
		if !contractState.AggregatedBalancesAmount.IsZero() {
			return fmt.Errorf("manifest: ContractState: `NumberOfBalanceRecords` is zero but `AggregatedBalancesAmount` is not zero")
		}
		if manifest.Reconciliation != nil {
			manifest.Reconciliation.ContractState = nil
		}
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

func (app *App) UpgradeContractAdmin(ctx types.Context, contractAddr *string, newAdmin *string, manifest *UpgradeManifest) error {
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

	if manifest.Contracts == nil {
		manifest.Contracts = new(Contracts)
	}

	manifest.Contracts.AdminUpdated = append(manifest.Contracts.AdminUpdated, ContractValueUpdate{*contractAddr, oldAdmin, *newAdmin})

	return nil
}

func (app *App) UpgradeContractAdmins(ctx types.Context, networkInfo *NetworkConfig, manifest *UpgradeManifest) error {
	contracts := []IContractAdmin{networkInfo.Contracts.Reconciliation, networkInfo.Contracts.TokenBridge}
	for _, contract := range contracts {
		if contract == nil {
			continue
		}
		err := app.UpgradeContractAdmin(ctx, contract.GetPrimaryContractAddr(), contract.GetNewAdminAddr(), manifest)
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

	if manifest.Contracts == nil {
		manifest.Contracts = new(Contracts)
	}

	manifest.Contracts.StateCleaned = append(manifest.Contracts.StateCleaned, contractAddr)

	return nil
}

func (app *App) DeleteContractStates(ctx types.Context, networkInfo *NetworkConfig, manifest *UpgradeManifest) error {
	var contractsToWipe []string

	contractsToWipe = networkInfo.Contracts.Reconciliation.GetContracts(contractsToWipe)

	contractsToWipe = networkInfo.Contracts.Almanac.GetContracts(contractsToWipe)
	contractsToWipe = networkInfo.Contracts.AName.GetContracts(contractsToWipe)

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

type ReconciliationContractStateRecord struct {
	Paused bool `json:"paused"`
}
