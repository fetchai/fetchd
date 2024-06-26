package app

import (
	"bufio"
	"bytes"
	_ "embed"
	"encoding/binary"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types"
	"strings"
)

//go:embed reconciliation_data.csv
//go:embed reconciliation_data_testnet.csv
var reconciliationDataTestnet []byte
var reconciliationData []byte
var reconciliationBalancesKey = prefixStringWithLength("balances")

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

func (app *App) ProcessReconciliation(ctx types.Context, networkInfo *NetworkConfig) error {
	records := readInputReconciliationData(reconciliationData)

	transfers, err := app.WithdrawReconciliationBalances(ctx, networkInfo, records)
	if err != nil {
		return fmt.Errorf("error withdrawing reconciliation balances: %v", err)
	}

	err = app.ReplaceReconciliationContractState(ctx, networkInfo, transfers)
	if err != nil {
		return fmt.Errorf("error replacing reconciliation contract state: %v", err)
	}

	return nil
}

func (app *App) WithdrawReconciliationBalances(ctx types.Context, networkInfo *NetworkConfig, records [][]string) ([]ReconciliationTransfer, error) {
	transfers := make([]ReconciliationTransfer, 0)
	landingAddr, err := types.AccAddressFromBech32(networkInfo.ReconciliationInfo.TargetAddress)
	if err != nil {
		return nil, err
	}

	if !app.AccountKeeper.HasAccount(ctx, landingAddr) {
		return nil, fmt.Errorf("landing address does not exist")
	}

	for _, record := range records {
		recordAddr, err := types.AccAddressFromBech32(record[2])
		if err != nil {
			return nil, err
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
			return nil, err
		}

		transfer := ReconciliationTransfer{
			EthAddr: record[0],
			From:    record[2],
			Amount:  recordBalanceCoins,
		}
		transfers = append(transfers, transfer)
	}

	return transfers, nil
}

func (app *App) ReplaceReconciliationContractState(ctx types.Context, networkInfo *NetworkConfig, reconciliationTransfers []ReconciliationTransfer) error {
	_, _, prefixStore, err := app.GetContractParams(ctx, networkInfo.Contracts.Reconciliation.Addr)
	if err != nil {
		return err
	}

	for _, transfer := range reconciliationTransfers {
		key, value := reconciliationContractStateBalancesRecord(transfer.EthAddr, transfer.Amount)
		if key == nil {
			continue
		}

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

type ReconciliationTransfer struct {
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
}

type Reconciliation struct {
	Addr     string
	NewAdmin *string
	NewLabel *string
}
