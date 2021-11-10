package module

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/fetchai/fetchd/types"
	abci "github.com/tendermint/tendermint/abci/types"
)

type InitGenesisHandler func(ctx types.Context, cdc codec.Marshaler, data json.RawMessage) ([]abci.ValidatorUpdate, error)
type ExportGenesisHandler func(ctx types.Context, cdc codec.Marshaler) (json.RawMessage, error)
