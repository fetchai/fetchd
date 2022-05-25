package network

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	dbm "github.com/tendermint/tm-db"

	"github.com/fetchai/fetchd/app"
	"github.com/fetchai/fetchd/app/params"
	"github.com/fetchai/fetchd/crypto/hd"
)

// DefaultConfig override the cosmos-sdk default test network config
// with our custom encoding config, and keyring options to enable BLS support
func DefaultConfig() network.Config {
	encodingConfig := app.MakeEncodingConfig()

	cfg := network.DefaultConfig()
	cfg.Codec = encodingConfig.Codec
	cfg.TxConfig = encodingConfig.TxConfig
	cfg.InterfaceRegistry = encodingConfig.InterfaceRegistry
	cfg.LegacyAmino = encodingConfig.Amino

	cfg.AppConstructor = NewAppConstructor(encodingConfig)
	cfg.KeyringOptions = append(cfg.KeyringOptions, func(options *keyring.Options) {
		options.SupportedAlgos = append(options.SupportedAlgos, hd.Bls12381)
	})

	return cfg
}

func NewAppConstructor(encodingCfg params.EncodingConfig) network.AppConstructor {
	return func(val network.Validator) servertypes.Application {
		return app.New(
			val.Ctx.Logger, dbm.NewMemDB(), nil, true, make(map[int64]bool), val.Ctx.Config.RootDir, 0,
			encodingCfg,
			simapp.EmptyAppOptions{},
			baseapp.SetPruning(storetypes.NewPruningOptionsFromString(val.AppConfig.Pruning)),
			baseapp.SetMinGasPrices(val.AppConfig.MinGasPrices),
		)
	}
}
