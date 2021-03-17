# Joining a testnet

In order to join the test network you will need to have the correct version of the fetchd ledger available on your system. For users who just want to connect to the network we recommend that you use the docker images. A number of
the setup steps have been automated to it is very quick to get started.

Alternatively if you plan to run a validator node then it will make more sense in the long term for you are familiar with operating a local version of the software.

## Using the docker image

Much more information is available from the [Docker Images](../docker-images/) pages however, to join a desired network you can simply run the docker images with the following command:

    docker run -e MONIKER=<insert node name here> -e NETWORK=<network name> fetchai/fetchd:0.5

**Beacon World Example**

To connect to the beacon world testnet users would simply need to run the following command

	docker run -e MONIKER=my-first-fetch-node -e NETWORK=beaconworld fetchai/fetchd:0.5

## Using a local version

Assuming that you have following the [installation guide](../building/). You should now have `fetchd` and `fetchcli` successfully installed in your path. You can check this with the following command:

```bash
which fetchd
which fetchcli
```

You can also verify that you are running the correct version for the [network](../networks/).

```bash
fetchd version
fetchcli version
```

### Configuring the client `fetchcli`

In general to configure the CLI to point at a given network it needs as a minimum the following configuration values

```bash
fetchcli config chain-id <chain-id>
fetchcli config trust-node false
fetchcli config node <rpc url>
```

**Beacon World Example**

In the case of the beacon world network this would be as follows:

```bash
fetchcli config chain-id beaconworld-2
fetchcli config trust-node false
fetchcli config node https://rpc-beaconworld.fetch.ai:443
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

**Beacon World Example**

Less abstractly then, if a user wants to connect to the beacon world test net for example. You would need to run the following steps:


```bash
# init
fetchd init my-first-fetch-node --chain-id beaconworld-2

# genesis
curl https://rpc-beaconworld.fetch.ai/genesis? | jq .result.genesis > ~/.fetchd/config/genesis.json

# start
fetchd start --p2p.seeds=29478d48b65482f4d627cad034281ebefb6fdc13@connect-beaconworld.fetch.ai:36656
```
