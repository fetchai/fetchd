# Joining a testnet

In order to join the test network you will need to have the correct version of the fetchd ledger available on your system. 

## Using a local version

Assuming that you have following the [installation guide](../building/). You should now have `fetchd` successfully installed in your path. You can check this with the following command:

```bash
which fetchd
```

You can also verify that you are running the correct version for the [network](../networks/).

```bash
fetchd version
```

### Configuring the server `fetchd`


Initialize fetchd by running command. This setups a default / empty genesis configuration.

```bash
fetchd init <Moniker-name> --chain-id <chain id>
```

Execute the following command to download the latest the genesis file:

```bash
curl <rpc url>/genesis | jq .result.genesis > ~/.fetchd/config/genesis.json`
```

Finally connect fetchd to the network by getting it to connect to a seed node for the given network.

```bash
fetchd start --p2p.seeds=<network seed peers>
```

**Stargate Example**

Less abstractly then, if a user wants to connect to the Stargate test net for example. You would need to run the following steps:


```bash
# init
fetchd init my-first-fetch-node --chain-id stargateworld-1

# genesis
curl https://rpc-stargateworld.fetch.ai/genesis? | jq .result.genesis > ~/.fetchd/config/genesis.json

# start
fetchd start --p2p.seeds=0831c7f4cb4b12fe02b35cc682c7edb03f6df36c@connect-stargateworld.t-v2-london-c.fetch-ai.com:36656
```
