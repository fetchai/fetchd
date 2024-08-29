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
	IBC                *UpgradeIBCTransfers       `json:"ibc,omitempty"`
	Minting            *UpgradeMinting            `json:"minting,omitempty"`
	MoveGenesisBalance *UpgradeMoveGenesisBalance `json:"move_genesis_balance,omitempty"`
	Delegate           *UpgradeDelegate           `json:"delegate,omitempty"`
	MoveMintedBalance  *UpgradeMoveMintedBalance  `json:"move_minted_balance,omitempty"`
}

func NewUpgradeManifest() *UpgradeManifest {
	return &UpgradeManifest{}
}

type UpgradeIBCTransfer struct {
	From      string      `json:"from"`
	ChannelID string      `json:"channel_id"`
	Amount    types.Coins `json:"amount"`
}

type UpgradeBalanceMovement struct {
	From   string      `json:"from"`
	To     string      `json:"to"`
	Amount types.Coins `json:"amount"`
	Memo   string      `json:"memo"`
}

type UpgradeIBCTransfers struct {
	Transfers                   []UpgradeIBCTransfer `json:"transfer"`
	To                          string               `json:"to"`
	AggregatedTransferredAmount types.Coins          `json:"aggregated_transferred_amount"`
	NumberOfTransfers           int                  `json:"number_of_transfers"`
}

type UpgradeMinting struct {
	Mints                  []UpgradeBalanceMovement `json:"mint"`
	AggregatedMintedAmount types.Coins              `json:"aggregated_minted_amount"`
	NumberOfMints          int                      `json:"number_of_mints"`
}

type UpgradeMoveGenesisBalance struct {
	Movements             []UpgradeBalanceMovement `json:"movements"`
	AggregatedMovedAmount types.Coins              `json:"aggregated_minted_amount"`
	NumberOfMovements     int                      `json:"number_of_mints"`
}

type UpgradeDelegate struct {
	Delegations               []UpgradeDelegation `json:"delegation"`
	AggregatedDelegatedAmount *types.Int          `json:"aggregated_minted_amount"`
	NumberOfDelegations       int                 `json:"number_of_delegations"`
}

type UpgradeDelegation struct {
	OriginalDelegator string    `json:"original_delegator"`
	Validator         string    `json:"validator"`
	Delegator         string    `json:"delegator"`
	Tokens            types.Int `json:"tokens"`
	NewShares         types.Dec `json:"new_shares"`
}

type UpgradeMoveMintedBalance struct {
	Movements []UpgradeBalanceMovement `json:"movements"`
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
