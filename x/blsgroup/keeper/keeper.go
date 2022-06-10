package keeper

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/fetchai/fetchd/x/blsgroup"
)

type Keeper struct {
	key storetypes.StoreKey

	cdc codec.Codec

	groupKeeper blsgroup.GroupKeeper
	accKeeper   blsgroup.AccountKeeper

	router *baseapp.MsgServiceRouter
}

func NewKeeper(storeKey storetypes.StoreKey, cdc codec.Codec, router *baseapp.MsgServiceRouter, groupKeeper blsgroup.GroupKeeper, accKeeper blsgroup.AccountKeeper) Keeper {
	return Keeper{
		key:         storeKey,
		cdc:         cdc,
		router:      router,
		groupKeeper: groupKeeper,
		accKeeper:   accKeeper,
	}
}
