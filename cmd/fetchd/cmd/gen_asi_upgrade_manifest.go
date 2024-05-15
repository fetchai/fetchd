package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types"
	"os"
)

type ASIUpgradeTransfer struct {
	From   string      `json:"from"`
	Amount types.Coins `json:"amount"`
}

type ASIUpgradeTransfers struct {
	Transfer []ASIUpgradeTransfer `json:"transfer"`
	To       string               `json:"to"`
}

type ASIUpgradeManifest struct {
	IBC            *ASIUpgradeTransfers `json:"ibc"`
	Reconciliation *ASIUpgradeTransfers `json:"reconciliation"`
}

func SaveASIManifest(manifest *ASIUpgradeManifest) error {
	var serialised_manifest []byte
	var err error
	if serialised_manifest, err = json.MarshalIndent(manifest, "", "\t"); err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	var f *os.File
	const manifestFilePath = "asi_upgrade_manifest.json"
	if f, err = os.Create(manifestFilePath); err != nil {
		return fmt.Errorf("failed to create file \"%s\": %w", manifestFilePath, err)
	}
	defer f.Close()
	if _, err = f.Write(serialised_manifest); err != nil {
		return fmt.Errorf("failed to write manifest to the \"%s\" file : %w", manifestFilePath, err)
	}

	return nil
}
