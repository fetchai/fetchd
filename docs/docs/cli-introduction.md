# CLI - Introduction

The command line client provides all of the capabilities for interacting with the fetch ledger such as creating addresses, sending transactions and the governance capabilities. Before starting with the command line client you need to follow the installation instructions [here](building.md)

## Connecting to a network

While some users will want to connect a node to the network and sync the entire blockchain, for many however, it is quicker and easier to connect directly to existing publically available nodes.

### Connecting to fetchhub mainnet v2 network

To connect to the mainnet v2 network run the following configuration steps:

```bash
fetchcli config chain-id fetchhub-1
fetchcli config trust-node false
fetchcli config node https://rpc-fetchhub.fetch-ai.com:443
```

### Connecting to stargateworld network

On Stargate networks, the `fetchcli` binary **does not exists anymore**. Instead, every commands are now provided by `fetchd`. Beside this, the syntax is generaly the same. To connect to the stargateworld network run the following configuration steps:

```bash
fetchd config chain-id stargateworld-1
fetchd config node https://rpc-stargateworld.fetch-ai.com:443
```

### Connecting to Agent Land network

To connect to the agent land network run the following configuration steps:

```bash
fetchcli config chain-id agent-land
fetchcli config trust-node false
fetchcli config node https://rpc-agent-land.fetch.ai:443
```

### Connecting to Agent World network

To connect to the agent world network run the following configuration steps:

```bash
fetchcli config chain-id agentworld-1
fetchcli config trust-node false
fetchcli config node https://rpc-agentworld.prod.fetch-ai.com:443
```

Checkout the [Network Information](../networks/) page for more detailed information on the available networks.
