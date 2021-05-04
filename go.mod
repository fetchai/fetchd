module github.com/fetchai/fetchd

go 1.15

require (
	github.com/CosmWasm/wasmd v0.16.0-alpha1
	github.com/cosmos/cosmos-sdk v0.42.4
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/tendermint/tendermint v0.34.10
	github.com/tendermint/tm-db v0.6.4
)

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace github.com/cosmos/cosmos-sdk => github.com/fetchai/cosmos-sdk v0.16.5-0.20210504123449-62c9ebcbabc0
