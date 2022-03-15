package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types/errors"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/tendermint/tendermint/types"
)

func loadAppStateFromGenesis(genesisPath string) (genDoc *types.GenesisDoc, appState genutiltypes.AppMap, err error) {
	genDoc, err = types.GenesisDocFromFile(genesisPath)
	if err != nil {
		return nil, genutiltypes.AppMap{}, fmt.Errorf("failed to load genesis file at %q: %w", genesisPath, err)
	}
	if err := json.Unmarshal(genDoc.AppState, &appState); err != nil {
		return nil, genutiltypes.AppMap{}, errors.Wrap(err, "failed to JSON unmarshal initial genesis state")
	}
	return genDoc, appState, nil
}
