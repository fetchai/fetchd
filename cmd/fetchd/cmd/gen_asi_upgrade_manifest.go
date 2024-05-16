package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/types"
	config2 "github.com/tendermint/tendermint/config"
	"os"
	"path"
)

type ASIUpgradeTransfer struct {
	From   string      `json:"from"`
	Amount types.Coins `json:"amount"`
}

type ASIUpgradeTransfers struct {
	Transfer []ASIUpgradeTransfer `json:"transfer"`
	To       string               `json:"to"`
}

type ASIUpgradeSupplyMint struct {
	LandingAddress string      `json:"landing_address"`
	Amount         types.Coins `json:"amount"`
	NewSupplyTotal types.Coins `json:"new_supply_total"`
}

type ASIUpgradeManifest struct {
	SupplyMint     []ASIUpgradeSupplyMint `json:"supply_mint"`
	IBC            *ASIUpgradeTransfers   `json:"ibc,omitempty"`
	Reconciliation *ASIUpgradeTransfers   `json:"reconciliation,omitempty"`
}

func SaveASIManifest(manifest *ASIUpgradeManifest, config *config2.Config) error {
	var serialisedManifest []byte
	var err error
	if serialisedManifest, err = json.MarshalIndent(manifest, "", "\t"); err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	var f *os.File
	const manifestFilename = "asi_upgrade_manifest.json"
	genesisFilePath := config.GenesisFile()
	manifestFilePath := path.Join(path.Dir(genesisFilePath), manifestFilename)
	if f, err = os.Create(manifestFilePath); err != nil {
		return fmt.Errorf("failed to create file \"%s\": %w", manifestFilePath, err)
	}
	defer f.Close()

	if _, err = f.Write(serialisedManifest); err != nil {
		return fmt.Errorf("failed to write manifest to the \"%s\" file : %w", manifestFilePath, err)
	}

	return nil
}
