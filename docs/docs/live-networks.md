# Networks

## Mainnet

The chain identifier of our production network is fetchhub-4.

| Parameter      | Value |
| -------------- | ----- |
| Chain ID       | fetchhub-4 |
| Block range    | 5,300,201 --> |
| Date range     | 05/04/2022 --> |
| Denomination   | afet |
| Decimals       | 18 (1fet = 1000000000000000000afet) |
| Version        | [v0.10.3](https://github.com/fetchai/fetchd/releases/tag/v0.10.3) up to block 6295500 <br/> [v0.10.4](https://github.com/fetchai/fetchd/releases/tag/v0.10.4) for blocks > 6295500 |
| RPC Endpoint   | <https://rpc-fetchhub.fetch.ai:443> |
| GRPC Endpoint  | <https://grpc-fetchhub.fetch.ai:443> |
| REST Endpoint  | <https://rest-fetchhub.fetch.ai:443> |
| Block Explorer | [https://explore-fetchhub.fetch.ai](https://explore-fetchhub.fetch.ai) |
| Token Faucet   | N/A |
| Genesis        | `curl https://raw.githubusercontent.com/fetchai/genesis-fetchhub/fetchhub-4/fetchhub-4/data/genesis_migrated_5300200.json --output ~/.fetchd/config/genesis.json` |
| Seed Node(s)   | 17693da418c15c95d629994a320e2c4f51a8069b@connect-fetchhub.fetch.ai:36456,a575c681c2861fe945f77cb3aba0357da294f1f2@connect-fetchhub.fetch.ai:36457,d7cda986c9f59ab9e05058a803c3d0300d15d8da@connect-fetchhub.fetch.ai:36458 |
| Snapshots      | <https://storage.googleapis.com/fetch-ai-mainnet-snapshots/fetchhub-4-pruned.tgz> <br /> <https://storage.googleapis.com/fetch-ai-mainnet-snapshots/fetchhub-4-full.tgz> <br /> <https://storage.googleapis.com/fetch-ai-mainnet-snapshots/fetchhub-4-archive.tgz> |

## Test Nets

### Dorado

This network is running the same major version of fetchd as our mainnet (`fetchhub-4`), possibly at a more recent minor version.

It is stable for deploying smart contracts and testing IBC.

| Parameter       | Value  |
| --------------- | ------ |
| Chain ID        | dorado-1 |
| Denomination    | atestfet |
| Decimals        | 18 (1testfet = 1000000000000000000atestfet) |
| Min Gas Prices  | 1000000000atestfet |
| Version         | [v0.10.3](https://github.com/fetchai/fetchd/releases/tag/v0.10.3) up to block 947800 <br/> [v0.10.4](https://github.com/fetchai/fetchd/releases/tag/v0.10.4) for blocks > 947800 and < 2198000 <br/> [v0.10.5-rc1](https://github.com/fetchai/fetchd/releases/tag/v0.10.5-rc1) for blocks > 2198000 |
| RPC Endpoint    | <https://rpc-dorado.fetch.ai:443> |
| GRPC Endpoint   | <https://grpc-dorado.fetch.ai:443> |
| REST Endpoint   | <https://rest-dorado.fetch.ai:443> |
| Block Explorer  | [https://explore-dorado.fetch.ai/](https://explore-dorado.fetch.ai/) |
| Ledger Explorer | [https://browse-dorado.fetch.ai/](https://browse-dorado.fetch.ai/) |
| Token Faucet    | Use block explorer |
| Genesis         | `curl https://storage.googleapis.com/fetch-ai-testnet-genesis/genesis-dorado-827201.json --output ~/.fetchd/config/genesis.json` |
| Seed Node(s)    | eb9b9717975b49a57e62ea93aa4480e091ae0660@connect-dorado.fetch.ai:36556,46d2f86a255ece3daf244e2ca11d5be0f16cb633@connect-dorado.fetch.ai:36557,066fc564979b1f3173615f101b62448ac7e00eb1@connect-dorado.fetch.ai:36558 |
| Snapshots       | <https://storage.googleapis.com/fetch-ai-testnet-snapshots/dorado-pruned.tgz> <br /> <https://storage.googleapis.com/fetch-ai-testnet-snapshots/dorado-full.tgz> <br /> <https://storage.googleapis.com/fetch-ai-testnet-snapshots/dorado-archive.tgz> |

### Eridanus

This network is running the next major version of fetchd that is currently being developed.

It is UNSTABLE, and may be missing functionality included in dorado/mainnet versions.

| Parameter       | Value  |
| --------------- | ------ |
| Chain ID        | eridanus-1 |
| Denomination    | atestfet |
| Decimals        | 18 (1testfet = 1000000000000000000atestfet) |
| Min Gas Prices  | 1000000000atestfet |
| Version         | [v0.11.x](https://github.com/fetchai/fetchd/tree/release/v0.11.x) |
| RPC Endpoint    | <https://rpc-eridanus.fetch.ai:443> |
| GRPC Endpoint   | <https://grpc-eridanus.fetch.ai:443> |
| REST Endpoint   | [https://rest-eridanus.fetch.ai:443](https://rest-eridanus.fetch.ai/cosmos/base/tendermint/v1beta1/node_info) |
| Block Explorer  | [https://explore-eridanus.fetch.ai/](https://explore-eridanus.fetch.ai/) |
| Ledger Explorer | [https://browse-eridanus.fetch.ai/](https://browse-eridanus.fetch.ai/) |
| Token Faucet    | Use block explorer for atestfet, or <br /> `curl -X POST -H 'Content-Type: application/json' -d '{"address":"fetch1myaddress"}' https://faucet-eridanus.fetch.ai/api/v3/claims` <br /> `curl -X POST -H 'Content-Type: application/json' -d '{"address":"fetch1myaddress"}' https://faucet-lrn-eridanus.fetch.ai/api/v3/claims` <br /> `curl -X POST -H 'Content-Type: application/json' -d '{"address":"fetch1myaddress"}' https://faucet-mobx-eridanus.fetch.ai/api/v3/claims` |
| Genesis         | `curl https://rpc-eridanus.fetch.ai/genesis | jq '.result.genesis' > ~/.fetchd/config/genesis.json `|
| Seed Node(s)    | b129b5a93e9bb32ec7a300735569abd278725046@connect-eridanus.fetch.ai:36656,ed866a34fc47c088163b539ce8c89e0334f90468@connect-eridanus.fetch.ai:36657,25d9a60cdb9c05169ab9665793d0031d5864fd02@connect-eridanus.fetch.ai:36658 |
| Snapshots       | <https://storage.googleapis.com/fetch-ai-testnet-snapshots/eridanus-pruned.tgz> <br /> <https://storage.googleapis.com/fetch-ai-testnet-snapshots/eridanus-full.tgz> <br /> <https://storage.googleapis.com/fetch-ai-testnet-snapshots/eridanus-archive.tgz> |
