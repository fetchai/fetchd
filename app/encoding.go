package app

import (
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/fetchai/fetchd/app/params"
	"github.com/fetchai/fetchd/crypto/keys/bls12381"
)

// MakeEncodingConfig creates an EncodingConfig for testing
func MakeEncodingConfig() params.EncodingConfig {
	encodingConfig := params.MakeEncodingConfig()
	std.RegisterLegacyAminoCodec(encodingConfig.Amino)
	std.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	ModuleBasics.RegisterLegacyAminoCodec(encodingConfig.Amino)
	ModuleBasics.RegisterInterfaces(encodingConfig.InterfaceRegistry)

	bls12381.RegisterInterfaces(encodingConfig.InterfaceRegistry)
	bls12381.RegisterAmino(encodingConfig.Amino)
	bls12381.RegisterAmino(legacy.Cdc) // Deprecated, but still needed by keyring as cosmos-sdk v0.45.4. Removed in v0.46.x

	return encodingConfig
}
