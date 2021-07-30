module github.com/fetchai/fetchd

go 1.16

require (
	github.com/CosmWasm/wasmd v0.16.0
	github.com/cosmos/cosmos-sdk v0.42.7
	github.com/gorilla/mux v1.8.0
	github.com/prometheus/client_golang v1.10.0
	github.com/rakyll/statik v0.1.7
	github.com/spf13/cast v1.3.1
	github.com/spf13/cobra v1.1.3
	github.com/tendermint/tendermint v0.34.11
	github.com/tendermint/tm-db v0.6.4
)

// fix for "invalid Go type types.Dec for field ..." errors
// see: https://github.com/cosmos/cosmos-sdk/issues/8426
replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace github.com/cosmos/cosmos-sdk => github.com/fetchai/cosmos-sdk v0.17.4

replace github.com/tendermint/tendermint => github.com/fetchai/tendermint v1.0.0
