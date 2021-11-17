module github.com/fetchai/fetchd

go 1.16

require (
	github.com/CosmWasm/wasmd v0.20.0
	github.com/armon/go-metrics v0.3.9 // indirect
	github.com/cockroachdb/apd/v2 v2.0.2
	github.com/cosmos/cosmos-sdk v0.44.2
	github.com/gogo/protobuf v1.3.3
	github.com/golang/protobuf v1.5.2
	github.com/gorilla/mux v1.8.0
	github.com/grpc-ecosystem/grpc-gateway v1.16.0
	github.com/lib/pq v1.10.3 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/common v0.29.0 // indirect
	github.com/rakyll/statik v0.1.7
	github.com/regen-network/cosmos-proto v0.3.1
	github.com/spf13/cast v1.4.1
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	github.com/tendermint/tendermint v0.34.13
	github.com/tendermint/tm-db v0.6.4
	golang.org/x/crypto v0.0.0-20211115234514-b4de73f9ece8 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20211115160612-a5da7257a6f7
	google.golang.org/grpc v1.40.0
	google.golang.org/protobuf v1.27.1
	pgregory.net/rapid v0.4.7
)

// fix for "invalid Go type types.Dec for field ..." errors
// see: https://github.com/cosmos/cosmos-sdk/issues/8426
replace google.golang.org/grpc => google.golang.org/grpc v1.33.2

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.3-alpha.regen.1

replace github.com/cosmos/cosmos-sdk => github.com/fetchai/cosmos-sdk v0.0.0-20211112171545-a1b9f22f93df
