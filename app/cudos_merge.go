package app

import (
	"encoding/json"
	"fmt"
	"os"
)

func ReadGenesisFile(filePath string) (*GenesisState, error) {
	// Read the file into a byte slice
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading genesis file: %v", err)
	}

	// Parse the JSON data
	var genesisState GenesisState
	err = json.Unmarshal(fileData, &genesisState)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling genesis file: %v", err)
	}

	return &genesisState, nil
}
