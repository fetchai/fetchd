package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	authmiddleware "github.com/cosmos/cosmos-sdk/x/auth/middleware"

	"github.com/fetchai/fetchd/x/blsgroup"
)

type Keeper struct {
	key storetypes.StoreKey

	groupKeeper blsgroup.GroupKeeper
	accKeeper   blsgroup.AccountKeeper

	router *authmiddleware.MsgServiceRouter
}

var _ blsgroup.MsgServer = Keeper{}

func NewKeeper(storeKey storetypes.StoreKey, cdc codec.Codec, router *authmiddleware.MsgServiceRouter, groupKeeper blsgroup.GroupKeeper, accKeeper blsgroup.AccountKeeper) Keeper {
	return Keeper{
		key:         storeKey,
		router:      router,
		groupKeeper: groupKeeper,
		accKeeper:   accKeeper,
	}
}
