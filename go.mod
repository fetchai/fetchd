module github.com/fetchai/fetchd

go 1.13

require (
	github.com/CosmWasm/wasmd v0.9.1
	github.com/cosmos/cosmos-sdk v0.38.3
	github.com/otiai10/copy v1.0.2
	github.com/pkg/errors v0.9.1
	github.com/snikch/goodman v0.0.0-20171125024755-10e37e294daa
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.6.3
	github.com/stretchr/testify v1.6.1
	github.com/tendermint/go-amino v0.15.1
	github.com/tendermint/tendermint v0.33.9
	github.com/tendermint/tm-db v0.5.2
)

replace github.com/keybase/go-keychain => github.com/99designs/go-keychain v0.0.0-20191008050251-8e49817e8af4

// this include a few extra debug helpers on top of cosmos v0.38.3 but original also works fine
replace github.com/cosmos/cosmos-sdk => github.com/fetchai/cosmos-sdk v0.16.5-0.20210430155815-57e85bdd6d02

replace github.com/tendermint/tendermint => github.com/fetchai/cosmos-consensus v0.16.3
