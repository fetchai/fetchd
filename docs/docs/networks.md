# Networks

## Mainnet

Fetchhub is our mainnet v2 production network. 

<div class="admonition note">
  <p class="admonition-title">Note</p>
  <p>If when running the first time you have an error similar to: <b>panic: Any JSON doesn't have '@type'</b>, this mean you have installed the <b>Stargate version</b> of fetchd. Make sure you have switched to the proper branch as mentionned in the <b>Version field</b></p>
</div>

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


## Launchpad testnets

###  fetchhubtest

Fetchhubtest is our testnet v2, running similar versions than our production network. 

| Parameter      | Value                                                                                      |
| -------------- | ------------------------------------------------------------------------------------------ |
| Chain ID       | fetchhubtest-2                                                                             |
| Denomination   | atestfet                                                                                   |
| Decimals       | 18 (1testfet = 1000000000000000000atestfet)                                                |
| Version        | [v0.7.x (fetchcli >= v0.7.2)](https://github.com/fetchai/fetchd/tree/release/v0.7.x)       |
| RPC Endpoint   | https://rpc-fetchhubtest.fetch.ai:443                                                      |
| REST Endpoint  | https://rest-fetchhubtest.fetch.ai:443                                                     |
| Block Explorer | [https://explore-fetchhubtest.fetch.ai/](https://explore-fetchhubtest.fetch.ai/)           |
| Token Faucet   | Use block explorer                                                                         |
| Seed Node(s)   | 06da15abb82328a2fa7ba8b69925cf3fa73f1970@connect-fetchhubtest.t-v2-london-c.fetch-ai.com:36756 |


## Stargate testnets

### stargateworld

A new network, updated with latest patches from the Cosmos ecosystem. It is **not compatible** with the current mainnet version.
Note that with this version `fetchcli` has been removed and its functionality has been merged into `fetchd`. This means that with this version you will only need the single `fetchd` binary to either run a node or query the network.

| Parameter      | Value                                                                                      |
| -------------- | ------------------------------------------------------------------------------------------ |
| Chain ID       | stargateworld-2                                                                            |
| Denomination   | atestfet                                                                                   |
| Decimals       | 18 (1testfet = 1000000000000000000atestfet)                                                |
| Version        | [v0.8.x (fetchd >= v0.8.2)](https://github.com/fetchai/fetchd/tree/release/v0.8.x)                 |
| RPC Endpoint   | https://rpc-stargateworld.fetch.ai:443                                                     |
| GRPC Endpoint  | https://grpc-stargateworld.t-v2-london-c.fetch-ai.com:443                                  |
| REST Endpoint  | https://rest-stargateworld.fetch.ai:443                                                    |
| Block Explorer | [https://explore-stargateworld.fetch.ai/](https://explore-stargateworld.fetch.ai/)         |
| Token Faucet   | Use block explorer                                                                         |
| Seed Node(s)   | 0831c7f4cb4b12fe02b35cc682c7edb03f6df36c@connect-stargateworld.t-v2-london-c.fetch-ai.com:36656 |

### andromeda

A new network, updated with latest patches from the Cosmos ecosystem. It is **not compatible** with the current mainnet version.
Note that with this version `fetchcli` has been removed and its functionality has been merged into `fetchd`. This means that with this version you will only need the single `fetchd` binary to either run a node or query the network.

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
