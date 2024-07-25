package app

import (
	"encoding/json"
	"fmt"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func ProcessAccounts(app *App, genesisState map[string]json.RawMessage) error {

	authGenState := authtypes.GetGenesisStateFromAppState(app.appCodec, genesisState)

	accs, err := authtypes.UnpackAccounts(authGenState.Accounts)
	if err != nil {
		panic(fmt.Sprintf("failed to get accounts from any: %w", err))
	}

	// Handle accounts

	print(accs)
	return nil
}
