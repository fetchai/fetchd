package verifiablecredential_test

import (
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/fetchai/fetchd/testutil"
	tmjson "github.com/tendermint/tendermint/libs/json"

	"testing"

	"github.com/stretchr/testify/require"
	abcitypes "github.com/tendermint/tendermint/abci/types"

	"github.com/fetchai/fetchd/app"
	"github.com/tendermint/tendermint/libs/log"

	dbm "github.com/tendermint/tm-db"
)

func TestCreateModuleInApp(t *testing.T) {
	app := app.New(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		make(map[int64]bool),
		app.DefaultNodeHome,
		0,
		app.MakeEncodingConfig(),
		simapp.EmptyAppOptions{},
	)

	genesisState := testutil.GenesisStateWithSingleValidator(t, app)
	stateBytes, err := tmjson.Marshal(genesisState)
	require.NoError(t, err)

	app.InitChain(
		abcitypes.RequestInitChain{
			AppStateBytes: stateBytes,
			ChainId:       "test-chain-id",
		},
	)

	require.NotNil(t, app.VcKeeper)
}
