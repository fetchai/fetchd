# Fetch.ai fetchd repository

This repository contains the source code for validators on the Fetch network. The source is based on the [wasmd](https://github.com/CosmWasm/wasmd) variant of the Cosmos-SDK, which includes a virtual machine that compiles to WebAssembly. It contains Fetch.ai-specific updates required for the test networks and future mainnet, including a decentralized random beacon (DRB) and a novel, compact multi-signatures scheme. Versions of this repository are not currently syncrhonised with either wasmd or the Cosmos-SDK. Please refer to the [releases](https://github.com/fetchai/fetchd/releases) section for the compatiblity with upstream versions.

**Note**: Requires [Go 1.18+](https://golang.org/dl/)

## Supported Systems

The supported systems are limited by the dlls created in [`go-cosmwasm`](https://github.com/CosmWasm/go-cosmwasm). In particular, **we only support MacOS and Linux**.
For linux, the default is to build for glibc, and we cross-compile with CentOS 7 to provide
backwards compatibility for `glibc 2.12+`. This includes all known supported distributions
using glibc (CentOS 7 uses 2.12, obsolete Debian Jessy uses 2.19).

As of `0.5.x` we support `muslc` Linux systems, in particular **Alpine linux**,
which is popular in docker distributions. Note that we do **not** store the
static `muslc` build in the repo, so you must compile this yourself, and pass `-tags muslc`.
Please look at the [`Dockerfile`](./Dockerfile) for an example of how we build a static Go
binary for `muslc`. (Or just use this Dockerfile for your production setup).

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
>```
>export PATH=$PATH:~/go/bin
>```
>
>and reload it with:
>
>```
>source ~/.bashrc
>```

You can also verify that you are running the correct version 

```bash
fetchd version
```

This should print a version number that must be compatible with the network you're connecting to (see the [network page](../networks/) for the list of supported versions per network).

If instead you have an error: `Error: failed to parse log level (main:info,state:info,:error): Unknown Level String: 'main:info,state:info,:error', defaulting to NoLevel`, this means you had a pre-stargate version of fetchd (<= v0.7.x), and just installed a stargate version (>= v0.8.x), you'll need to remove the previous configuration files with:

```bash
rm ~/.fetchd/config/app.toml ~/.fetchd/config/config.toml
```

Alternatively, you can also build without installing the binary with:

```bash
make build
```

The fetchd binary will be available under `./build/fetchd`.

## Run a simple local test network

The easiest way to get started with a simple network is to run the [docker-compose](https://docs.docker.com/compose/). The details of this can be found [here](https://github.com/fetchai/fetchd/blob/master/docker-compose.yml). By default it will launch a small 3 validator nodes network.

## Resources

1. [Website](https://fetch.ai/)
2. [Documenation](https://docs.fetch.ai/ledger_v2/)
3. [Discord Server](https://discord.gg/UDzpBFa)
4. [Blog](https://fetch.ai/blog)
5. [Community Website](https://community.fetch.ai/)
6. [Community Telegram Group](https://t.me/fetch_ai)
