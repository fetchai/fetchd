# CLI - Introduction

The command line client provides all of the capabilities for interacting with the fetch ledger such as creating addresses, sending transactions and the governance capabilities. Before starting with the command line client you need to follow the installation instructions [here](building.md)

## Connecting to a network

While some users will want to connect a node to the network and sync the entire blockchain, for many however, it is quicker and easier to connect directly to existing publically available nodes.

### Connecting to fetchhub mainnet 

To connect to the mainnet run the following configuration steps:

```bash
fetchd config chain-id fetchhub-3
fetchd config node https://rpc-fetchhub.fetch.ai:443
```

### Connecting to dorado network

To connect to the dorado network run the following configuration steps:

```bash
fetchd config chain-id dorado-1
fetchd config node https://rpc-dorado.fetch.ai:443
```

Checkout the [Network Information](../networks/) page for more detailed information on the available networks.
