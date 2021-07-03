# Networks

## Mainnet v2: fetchhub

Fetchhub is our mainnet v2 production network. 

| Parameter      | Value                                                                                      |
| -------------- | ------------------------------------------------------------------------------------------ |
| Chain ID       | fetchhub-1                                                                                 |
| Denomination   | afet                                                                                       |
| Decimals       | 18 (1fet = 1000000000000000000afet)                                                        |
| Version        | v0.7.x (fetchcli >= v0.7.2)                                                                |
| RPC Endpoint   | https://rpc-fetchhub.fetch-ai.com:443                                                      |
| REST Endpoint  | https://rest-fetchhub.fetch-ai.com:443                                                     |
| Block Explorer | [https://explore-fetchhub.fetch.ai](https://explore-fetchhub.fetch.ai)                     |
| Token Faucet   | N/A                                                                                        |
| Seed Node(s)   | 9f9774d88bb6ff9f43b395baf0e7a4baba27dec6@connect-fetchhub.m-v2-london-c.fetch-ai.com:36856 |

## Testnet v2: fetchhubtest

Fetchhubtest is our testnet v2, running similar versions than our production network. 

| Parameter      | Value                                                                                      |
| -------------- | ------------------------------------------------------------------------------------------ |
| Chain ID       | fetchhubtest-2                                                                             |
| Denomination   | atestfet                                                                                   |
| Decimals       | 18 (1testfet = 1000000000000000000atestfet)                                                |
| Version        | v0.7.x (fetchcli >= v0.7.2)                                                                |
| RPC Endpoint   | https://rpc-fetchhubtest.fetch.ai:443                                                      |
| REST Endpoint  | https://rest-fetchhubtest.fetch.ai:443                                                     |
| Block Explorer | [https://explore-fetchhubtest.fetch.ai/](https://explore-fetchhubtest.fetch.ai/)           |
| Token Faucet   | Use block explorer                                                                         |
| Seed Node(s)   | 06da15abb82328a2fa7ba8b69925cf3fa73f1970@connect-fetchhubtest.t-v2-london-c.fetch-ai.com:36756 |

# Upcomming networks

## Stargate testnet: stargateworld

A new network, updated with latest patches from the Cosmos ecosystem. It is **not compatible** with the current mainnet version.
Note that with this version `fetchcli` has been removed and its functionality has been merged into `fetchd`. This means that with this version you will only need the single `fetchd` binary to either run a node or query the network.

| Parameter      | Value                                                                                      |
| -------------- | ------------------------------------------------------------------------------------------ |
| Chain ID       | stargateworld-1                                                                            |
| Denomination   | atestfet                                                                                   |
| Decimals       | 18 (1testfet = 1000000000000000000atestfet)                                                |
| Version        | v0.8.x (fetchd >= v0.8.0-rc5)                                                              |
| RPC Endpoint   | https://rpc-stargateworld.fetch.ai:443                                                     |
| REST Endpoint  | https://rest-stargateworld.fetch.ai:443                                                    |
| Block Explorer | [https://explore-stargateworld.fetch.ai/](https://explore-stargateworld.fetch.ai/)         |
| Token Faucet   | Use block explorer                                                                         |
| Seed Node(s)   | 0831c7f4cb4b12fe02b35cc682c7edb03f6df36c@connect-stargateworld.t-v2-london-c.fetch-ai.com:36656 |

# Deprecated networks

These networks are still up and kept for backward compatibility until everything is migrated off from them.

## Agent Land

Agent Land is our stable, public testnet for the Fetch Ledger v2. As such most developers will be interacting with this testnet. This is specifically designed and supported for autonomous economic agent development. There are other testnets, such as those supporting our unique DRB (decentralized random beacon) and other exciting technologies. When we come to the mainnet, all of these testnets will become one: a single network supporting all the new features.

| Parameter      | Value                                                                      |
| -------------- | -------------------------------------------------------------------------- |
| Chain ID       | agent-land                                                                 |
| Denomination   | atestfet                                                                   |
| Decimals       | 18 (1testfet = 1000000000000000000atestfet)                                |
| Version        | v0.2.x (fetchcli <= v0.2.7)                                                |
| RPC Endpoint   | https://rpc-agent-land.fetch.ai:443                                        |
| REST Endpoint  | https://rest-agent-land.fetch.ai:443                                       |
| Block Explorer | [https://explore-agent-land.fetch.ai](https://explore-agent-land.fetch.ai) |
| Token Faucet   | Use block explorer                                                         |

You can read more detailed information on [Github](https://github.com/fetchai/networks-agentland).

## **Incentivized Testnet Phase 1: Agent World**

The Agent World incentivized test network is phase 1 of our journey to Mainnet v2. Check out the [Incentivised Testnets](../../i_nets/) for more information.

| Parameter      | Value                                                                        |
| -------------- | ---------------------------------------------------------------------------- |
| Chain ID       | agentworld-1                                                                 |
| Denomination   | atestfet                                                                     |
| Decimals       | 18 (1testfet = 1000000000000000000atestfet)                                  |
| Version        | v0.2.x (fetchcli <= v0.2.7)                                                  |
| RPC Endpoint   | https://rpc-agentworld.fetch.ai:443                                          |
| REST Endpoint  | https://rest-agentworld.fetch.ai:443                                         |
| Block Explorer | [https://explore-agentworld.fetch.ai/](https://explore-agentworld.fetch.ai/) |
| Token Faucet   | n/a                                                                          |


# Read more / archives about Fetch.ai's networks and blockchain

[How to join the Fetch.ai Mainnet 2.0 and Take it to the Next Level](https://fetch.ai/how-to-join-the-fetch-ai-mainnet-2-0-and-take-it-to-the-next-level/)

[Join the Fetch.ai Journey into Decentralized AI]
(https://fetch.ai/join-the-fetch-ai-journey-into-decentralised-ai/)

[Pathway to Mainnet v2.0–2021 Q1 roadmap](https://fetch.ai/pathway-to-mainnet-v2-0-2021-q1-roadmap/)

[Build a global decentralized multi-agent system: Join and participate in our testnet program](https://fetch.ai/build-a-global-decentralized-multi-agent-system-join-and-participate-in-our-testnet-program/)

[Build with Fetch.ai: Participate in the Incentivized Testnet program launching October 22nd](https://fetch.ai/build-with-fetch-ai-participate-in-the-incentivized-testnet-program-launching-october-22nd/)

[Launching our Random Number Beacon on Binance Smart Chain](https://fetch.ai/launching-our-random-number-beacon-on-binance-smart-chain/)

[Revealing the Fetch.ai interoperability vision with our new virtual machine](https://fetch.ai/fetch-ai-announces-major-interoperability-update-to-its-network-enabling-fetch-ai-fet-technology-to-be-delivered-across-multiple-blockchains/)

[Boötes: Building on the foundations of our mainnet](https://fetch.ai/bootes-building-on-the-foundations-of-our-mainnet/)

[Lightning Fast Consensus Becomes a Reality Thanks to Fetch.ai Breakthrough](https://fetch.ai/lightning-fast-consensus-becomes-a-reality-thanks-to-fetch-ai-breakthrough/)

[Lifting the hood on Mainnet: the good bits!](https://fetch.ai/lifting-the-hood-on-mainnet-the-good-bits/)

[Achievement unlocked: mainnet release](https://fetch.ai/achievement-unlocked-mainnet-release/)

[Aries: One small step to mainnet](https://fetch.ai/aries-one-small-step-to-mainnet/)

[Ara: Making smart contract transactions smarter than ever](https://fetch.ai/ara-making-smart-contract-transactions-smarter-than-ever/)

[How to learn Etch: A tool for Solidity developers](https://fetch.ai/how-to-learn-etch-a-tool-for-solidity-developers/)

[Fetch.ai launches beta mainnet with Aquila release](https://fetch.ai/fetch-ai-launches-beta-mainnet-with-aquila-release/)

[Fetch.ai’s latest release speaks your language](https://fetch.ai/fetch-ais-latest-release-speaks-your-language/)

[The Future of Consensus: Proof-of-Stake with Unpermissioned Delegation](https://fetch.ai/the-future-of-consensus-proof-of-stake-with-unpermissioned-delegation/)

[Fetch.ai’s alpha mainnet Apus: the foundation for the future](https://fetch.ai/fetch-ais-alpha-mainnet-apus-the-foundation-for-the-future/)

[Fetch.ai Public Test Network: Foundation Release](https://fetch.ai/fetch-ai-public-test-network-foundation-release/)

[Fetch.ai Ledger Benchmarking II — Single Lane Performance](https://fetch.ai/fetch-ai-ledger-benchmarking-ii-single-lane-performance/)

[Removing the honesty box from the economy with an ANVIL](https://fetch.ai/removing-the-honesty-box-from-the-economy-with-an-anvil-2/)

[Smart Contracts for Smart Markets](https://fetch.ai/smart-contracts-for-smart-markets-2/)

[Fetch.ai Ledger Benchmarking I — Overview and Architecture](https://fetch.ai/fetch-ai-ledger-benchmarking-i-overview-and-architecture/)

[Introducing the Fetch.ai Virtual Machine](https://fetch.ai/introducing-the-fetch-ai-virtual-machine/)

[Design of the Fetch.ai Scalable Ledger](https://fetch.ai/design-of-the-fetch-ai-scalable-ledger/)

[Understanding the Fetch.ai network: a guide](https://fetch.ai/understanding-the-fetch-ai-network-a-guide/)

[Synergetic Smart Contracts with Fetch.ai](https://fetch.ai/synergetic-smart-contracts-with-fetch-ai/)


