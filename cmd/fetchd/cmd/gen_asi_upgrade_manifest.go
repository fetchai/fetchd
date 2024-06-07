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

type ASIUpgradeSupply struct {
	LandingAddress       string      `json:"landing_address"`
	MintedAmount         types.Coins `json:"minted_amount"`
	ResultingSupplyTotal types.Coins `json:"resulting_supply_total"`
}

type ValueUpdate struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type MainParams struct {
	GenesisTime   *ValueUpdate      `json:"genesis_time,omitempty"`
	ChainID       *ValueUpdate      `json:"chain_id,omitempty"`
	AddressPrefix *ValueUpdate      `json:"address_prefix,omitempty"`
	Supply        *ASIUpgradeSupply `json:"supply,omitempty"`
}

type ASIUpgradeManifest struct {
	Main           *MainParams          `json:"main,omitempty"`
	IBC            *ASIUpgradeTransfers `json:"ibc,omitempty"`
	Reconciliation *ASIUpgradeTransfers `json:"reconciliation,omitempty"`
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
