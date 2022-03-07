# Networks

## Mainnet

The chain identifier of our production network is fetchhub-3.

| Parameter      | Value                                                                                      |
| -------------- | ------------------------------------------------------------------------------------------ |
| Chain ID       | fetchhub-3                                                                                 |
| Denomination   | afet                                                                                       |
| Decimals       | 18 (1fet = 1000000000000000000afet)                                                        |
| Version        | [v0.9.x (fetchcli >= v0.9.0)](https://github.com/fetchai/fetchd/tree/release/v0.9.x)       |
| RPC Endpoint   | <https://rpc-fetchhub.fetch.ai:443>                                                        |
| GRPC Endpoint  | <https://grpc-fetchhub.fetch.ai:443>                                                       |
| REST Endpoint  | <https://rest-fetchhub.fetch.ai:443>                                                       |
| Block Explorer | [https://explore-fetchhub.fetch.ai](https://explore-fetchhub.fetch.ai)                     |
| Token Faucet   | N/A                                                                                        |
| Seed Node(s)   | 5f3fa6404a67b664be07d0e133a00c1600967396@connect-fetchhub.fetch.ai:36756                   |

## Test Nets

### Capricorn

This network is running the same software as our mainnet (`fetchhub-3`), and is stable for deploying smart contracts and testing IBC.

| Parameter      | Value                                                                                      |
| -------------- | ------------------------------------------------------------------------------------------ |
| Chain ID       | capricorn-1                                                                                |
| Denomination   | atestfet                                                                                   |
| Decimals       | 18 (1testfet = 1000000000000000000atestfet)                                                |
| Min Gas Prices | 5000000000atestfet                                                                         |
| Version        | [v0.9.0 (fetchd >= v0.9.0)](https://github.com/fetchai/fetchd/releases/tag/v0.9.0)         |
| RPC Endpoint   | <https://rpc-capricorn.fetch.ai:443>                                                       |
| GRPC Endpoint  | <https://grpc-capricorn.fetch.ai:443>                                                      |
| REST Endpoint  | <https://rest-capricorn.fetch.ai:443>                                                      |
| Block Explorer | [https://explore-capricorn.fetch.ai/](https://explore-capricorn.fetch.ai/)                 |
| Token Faucet   | Use block explorer                                                                         |
| Seed Node(s)   | fec822ecf6e503a694a709ce663fd0c6da5fda3e@connect-capricorn.fetch.ai:36956                  |

### Dorado

This network is used for testing the future upgrade to mainnet.

| Parameter       | Value                                                                                      |
| --------------- | ------------------------------------------------------------------------------------------ |
| Chain ID        | dorado-1                                                                                   |
| Denomination    | atestfet                                                                                   |
| Decimals        | 18 (1testfet = 1000000000000000000atestfet)                                                |
| Min Gas Prices  | 1000000000atestfet                                                                         |
| Version         | [v0.10.x (fetchd >= v0.10.x)](https://github.com/fetchai/fetchd/releases/tag/v0.10.0-rc1)  |
| RPC Endpoint    | <https://rpc-dorado.fetch.ai:443>                                                          |
| GRPC Endpoint   | <https://grpc-dorado.fetch.ai:443>                                                         |
| REST Endpoint   | <https://rest-dorado.fetch.ai:443>                                                         |
| Block Explorer  | [https://explore-dorado.fetch.ai/](https://explore-dorado.fetch.ai/)                       |
| Ledger Explorer | [https://browse-dorado.fetch.ai/](https://browse-dorado.fetch.ai/)                         |
| Token Faucet    | Use block explorer                                                                         |
| Seed Node(s)    | b9b9717975b49a57e62ea93aa4480e091ae0660@connect-dorado.fetch.ai:36556,46d2f86a255ece3daf244e2ca11d5be0f16cb633@connect-dorado.fetch.ai:36557,066fc564979b1f3173615f101b62448ac7e00eb1@connect-dorado.fetch.ai:36558 |
