module github.com/fetchai/fetchd

go 1.13

require (
	github.com/CosmWasm/go-cosmwasm v0.10.0
	github.com/CosmWasm/wasmd v0.10.0
	github.com/cosmos/cosmos-sdk v0.39.1-0.20200727135228-9d00f712e334
	github.com/google/gofuzz v1.0.0
	github.com/gorilla/mux v1.7.4
	github.com/magiconair/properties v1.8.1
	github.com/otiai10/copy v1.0.2
	github.com/pkg/errors v0.9.1
	github.com/snikch/goodman v0.0.0-20171125024755-10e37e294daa
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.6.3
	github.com/stretchr/testify v1.6.1
	github.com/tendermint/go-amino v0.15.1
	github.com/tendermint/tendermint v0.33.7
	github.com/tendermint/tm-db v0.5.1
	gopkg.in/yaml.v2 v2.3.0
)

replace github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4

// this include a few extra debug helpers on top of cosmos v0.38.3 but original also works fine
replace github.com/cosmos/cosmos-sdk => github.com/fetchai/cosmos-sdk v0.7.1

replace github.com/tendermint/tendermint => github.com/fetchai/cosmos-consensus v0.6.1
