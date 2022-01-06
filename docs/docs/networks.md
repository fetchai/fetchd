# Networks

## Mainnet

Fetchhub is our mainnet v2 production network.

| Parameter      | Value                                                                                      |
| -------------- | ------------------------------------------------------------------------------------------ |
| Chain ID       | fetchhub-2                                                                                 |
| Denomination   | afet                                                                                       |
| Decimals       | 18 (1fet = 1000000000000000000afet)                                                        |
| Version        | [v0.8.x (fetchcli >= v0.8.7)](https://github.com/fetchai/fetchd/tree/release/v0.8.x)               |
| RPC Endpoint   | https://rpc-fetchhub.fetch.ai:443                                                      |
| REST Endpoint  | https://rest-fetchhub.fetch.ai:443                                                     |
| Block Explorer | [https://explore-fetchhub.fetch.ai](https://explore-fetchhub.fetch.ai)                     |
| Token Faucet   | N/A                                                                                        |
| Seed Node(s)   | c240ca38b990a7d3f25621eb848d0222d5b50278@connect-fetchhub2.m-v2-london-c.fetch-ai.com:36656 |

## Test Nets

### Stargateworld

This network is running the same software as our mainnet (`fetchhub-2`), and is stable for deploying smart contracts and testing.

| Parameter      | Value                                                                                      |
| -------------- | ------------------------------------------------------------------------------------------ |
| Chain ID       | stargateworld-3                                                                            |
| Denomination   | atestfet                                                                                   |
| Decimals       | 18 (1testfet = 1000000000000000000atestfet)                                                |
| Version        | [v0.8.x (fetchd >= v0.8.2)](https://github.com/fetchai/fetchd/tree/release/v0.8.x)                 |
| RPC Endpoint   | https://rpc-stargateworld.fetch.ai:443                                                     |
| GRPC Endpoint  | https://grpc-stargateworld.t-v2-london-c.fetch-ai.com:443                                  |
| REST Endpoint  | https://rest-stargateworld.fetch.ai:443                                                    |
| Block Explorer | [https://explore-stargateworld.fetch.ai/](https://explore-stargateworld.fetch.ai/)         |
| Token Faucet   | Use block explorer                                                                         |
| Seed Node(s)   | 0831c7f4cb4b12fe02b35cc682c7edb03f6df36c@connect-stargateworld.t-v2-london-c.fetch-ai.com:36656 |

### Andromeda

This network is also a clone of mainnet but is used for internal testing and network upgrades.

| Parameter      | Value                                                                                      |
| -------------- | ------------------------------------------------------------------------------------------ |
| Chain ID       | andromeda-1                                                                                |
| Denomination   | atestfet                                                                                   |
| Decimals       | 18 (1testfet = 1000000000000000000atestfet)                                                |
| Version        | [v0.8.x (fetchd >= v0.8.2)](https://github.com/fetchai/fetchd/tree/release/v0.8.x)                 |
| RPC Endpoint   | https://rpc-andromeda.fetch.ai:443                                                         |
| REST Endpoint  | https://rest-andromeda.fetch.ai:443                                                        |
| Block Explorer | [https://explore-andromeda.fetch.ai/](https://explore-andromeda.fetch.ai/)                 |
| Token Faucet   | Use block explorer                                                                         |
| Seed Node(s)   | f14fc7f2e6e2fabe9b11406333252f30973e0af1@connect-andromeda.fetch.ai:36856                  |

### Capricorn

This test-net is used for testing the future upgrade to mainnet. It supports Cosmwasm v1.0 release candidate smart contracts.

| Parameter      | Value                                                                                      |
| -------------- | ------------------------------------------------------------------------------------------ |
| Chain ID       | capricorn-1                                                                                |
| Denomination   | atestfet                                                                                   |
| Decimals       | 18 (1testfet = 1000000000000000000atestfet)                                                |
| Version        | [v0.9.0-rc1 (fetchd >= v0.9.0-rc1)](https://github.com/fetchai/fetchd/tree/integration/capricorn)      |
| RPC Endpoint   | https://rpc-capricorn.fetch.ai:443                                                         |
| REST Endpoint  | https://rest-capricorn.fetch.ai:443                                                        |
| Block Explorer | [https://explore-capricorn.fetch.ai/](https://explore-capricorn.fetch.ai/)                 |
| Token Faucet   | Use block explorer                                                                         |
