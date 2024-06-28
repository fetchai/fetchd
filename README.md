# Fetch.ai fetchd repository

This repository contains the source code for validators on the Fetch network. The source is based on the [wasmd](https://github.com/CosmWasm/wasmd) variant of the Cosmos-SDK, which includes a virtual machine that compiles to WebAssembly. It contains Fetch.ai-specific updates required for the test networks and future mainnet, including a decentralized random beacon (DRB) and a novel, compact multi-signatures scheme. Versions of this repository are not currently synchronised with either wasmd or the Cosmos-SDK. Please refer to the [releases](https://github.com/fetchai/fetchd/releases) section for the compatibility with upstream versions.

**Note**: Requires [Go 1.18+](https://golang.org/dl/)

## Supported Systems

The supported systems are limited by the dlls created in [`go-cosmwasm`](https://github.com/CosmWasm/go-cosmwasm). In particular, **we only support MacOS and Linux**.

## Quick Start

### Building and testing the project

First, install golang >= v1.18 (follow the guide from [https://golang.org/dl/](https://golang.org/dl/)) and execute the following commands:

```bash
# make sure you have the following packages:
apt-get update && apt-get install -y make gcc

# install fetchd. This will output the binary in ~/go/bin/ folder by default.
make install
```

You should now have `fetchd` successfully installed in your path. You can check this with the following command:

```bash
which fetchd
```

This should return a path such as `~/go/bin/fetchd` (might be different depending on your actual go installation).

> If you get no output, or an error such as `which: no fetchd in ...`, possible cause can either be that `make install` failed with some errors or that your go binary folder (default: ~/go/bin) is not in your `PATH`.
>
> To add the ~/go/bin folder to your PATH, add this line at the end of your ~/.bashrc:
>
>```bash
>export PATH=$PATH:~/go/bin
>```
>
>and reload it with:
>
>```bash
>source ~/.bashrc
>```

You can also verify that you are running the correct version

```bash
fetchd version
```

This should print a version number that must be compatible with the network you're connecting to (see the [network page](https://docs.fetch.ai/ledger_v2/live-networks/) for the list of supported versions per network).

Alternatively, you can also build without installing the binary with:

```bash
make build
```

The fetchd binary will be available under `./build/fetchd`.

## Run a simple local test network

The easiest way to get started with a simple network is to run the [docker-compose](https://docs.docker.com/compose/). The details of this can be found [here](https://github.com/fetchai/fetchd/blob/master/docker-compose.yml). By default it will launch a small 3 validator nodes network.

## Resources

1. [Website](https://fetch.ai/)
2. [Documentation](https://docs.fetch.ai/ledger_v2/)
3. [Discord Server](https://discord.gg/fetchai)
4. [Blog](https://fetch.ai/blog)
5. [Community Telegram Group](https://t.me/fetch_ai)
