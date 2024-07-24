package app

import (
	"encoding/json"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"io/ioutil"
	"log"
	"os"
)

// ParseJSON recursively parses JSON data into a map.
func ParseJSON(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ReadJSONFile reads a JSON file and returns its content as a byte slice.
func ReadJSONFile(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	return byteValue, nil
}

func ReadGenesisFile(filePath string) (map[string]interface{}, error) {
	jsonData, err := ReadJSONFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read JSON file: %v", err)
	}

	parsedData, err := ParseJSON(jsonData)
	if err != nil {
		log.Fatalf("Failed to parse JSON data: %v", err)
	}

	return parsedData, nil
}

func GetBankBalances(genesis map[string]interface{}) (map[string][]sdk.Coin, error) {

	// Create the map to store balances by address
	balanceMap := make(map[string][]sdk.Coin)

	// Unsafe way to access and iterate over app_state -> bank -> balances
	appState := genesis["app_state"].(map[string]interface{})
	bank := appState["bank"].(map[string]interface{})
	balances := bank["balances"].([]interface{})

	for _, res := range balances {
		address := res.(map[string]interface{})["address"].(string)
		balance := res.(map[string]interface{})["coins"]

		// Create a list to store coins for this address
		var coins []sdk.Coin

		for _, coinsRes := range balance.([]interface{}) {
			denom := coinsRes.(map[string]interface{})["denom"].(string)
			amount := coinsRes.(map[string]interface{})["amount"].(string)
			// Convert amount to sdk.Int
			amountInt, ok := sdk.NewIntFromString(amount)
			if !ok {
				panic("Failed to convert amount to sdk.Int: %s" + amount)
			}
			// Append coin to the list
			coin := sdk.Coin{
				Denom:  denom,
				Amount: amountInt,
			}
			coins = append(coins, coin)
		}

		// Store the list of coins in the map
		balanceMap[address] = coins
	}

	return balanceMap, nil
}
