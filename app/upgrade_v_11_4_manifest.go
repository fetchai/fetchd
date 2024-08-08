package app

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types"
	"os"
	"path"
)

const manifestFilenameBase = "upgrade_manifest.json"

type UpgradeManifest struct {
	Reconciliation *UpgradeReconciliation `json:"reconciliation,omitempty"`
	Contracts      *Contracts             `json:"contracts,omitempty"`
	IBC            *UpgradeIBCTransfers   `json:"ibc,omitempty"`
	Minting        *UpgradeMinting        `json:"minting,omitempty"`
}

func NewUpgradeManifest() *UpgradeManifest {
	return &UpgradeManifest{
		Contracts: &Contracts{},
	}
}

type Contracts struct {
	StateCleaned   []string                `json:"contracts_state_cleaned,omitempty"`
	AdminUpdated   []ContractValueUpdate   `json:"contracts_admin_updated,omitempty"`
	LabelUpdated   []ContractValueUpdate   `json:"contracts_label_updated,omitempty"`
	VersionUpdated []ContractVersionUpdate `json:"version_updated,omitempty"`
}

type ContractValueUpdate struct {
	Address string `json:"address"`
	From    string `json:"from"`
	To      string `json:"to"`
}

type ContractVersionUpdate struct {
	Address string           `json:"address"`
	From    *ContractVersion `json:"from,omitempty"`
	To      *ContractVersion `json:"to"`
}

type ValueUpdate struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type UpgradeIBCTransfer struct {
	From      string      `json:"from"`
	ChannelID string      `json:"channel_id"`
	Amount    types.Coins `json:"amount"`
}

type UpgradeMint struct {
	to     string      `json:"to"`
	Amount types.Coins `json:"amount"`
}

type UpgradeIBCTransfers struct {
	Transfers                   []UpgradeIBCTransfer `json:"transfer"`
	To                          string               `json:"to"`
	AggregatedTransferredAmount types.Coins          `json:"aggregated_transferred_amount"`
	NumberOfTransfers           int                  `json:"number_of_transfers"`
}

type UpgradeReconciliation struct {
	Transfers     *UpgradeReconciliationTransfers     `json:"transfers,omitempty"`
	ContractState *UpgradeReconciliationContractState `json:"contract_state,omitempty"`
}

type UpgradeReconciliationTransfer struct {
	From    string      `json:"from"`
	EthAddr string      `json:"eth_addr"`
	Amount  types.Coins `json:"amount"`
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

type UpgradeMinting struct {
	Mints                  []UpgradeMint `json:"mint"`
	AggregatedMintedAmount types.Coins   `json:"aggregated_minted_amount"`
	NumberOfMints          int           `json:"number_of_mints"`
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
