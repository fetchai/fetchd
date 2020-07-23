# Fetch.ai Wasmd Fork

This repository is the fork from the original [wasmd](https://github.com/CosmWasm/wasmd) which is the first implementation of a cosmos zone with wasm smart contracts enabled. It contains Fetch.ai specific updates required for the test networks. For this reason it is versioned independantly. Please refer to the [releases](https://github.com/fetchai/cosmos-sdk/releases) section for the compatiblity with upstream versions.

This code was forked from the `cosmos/gaia` repository and the majority of the codebase is the same as `gaia`.

**Note**: Requires [Go 1.13+](https://golang.org/dl/)

**Compatibility**: Last merge from `cosmos/gaia` was `090c545347b03e59415a18107a0a279c703c8f40` (Jan 23, 2020)

The `v0.7.2` release is based on `cosmos-sdk` `v0.38.3` and should be used in any public networks (as that fixed a security issue).

## Supported Systems

The supported systems are limited by the dlls created in [`go-cosmwasm`](https://github.com/CosmWasm/go-cosmwasm). In particular, **we only support MacOS and Linux**. Even for Linux, we only support `glibc 2.12+`. This includes all known supported distributions using glibc (CentOS 7 uses 2.12, obsolete Debian Jessy uses 2.19). **We do not currently support `muslc` systems, such as Alpine Linux**, which is popular in docker distributions.

There is some work done in `go-cosmwasm` to extend support, but it is rather tricky with Rust build flags.

## Stability

**This is alpha software, do not run on a production system.** Notably, we currently provide **no migration path** not even "dump state and restart" to move to future versions. At **beta** we will begin to offer migrations and better backwards compatibility guarantees.

With the `v0.6.0` tag, we are entering semver. That means anything with `v0.6.x` tags is compatible with each other, and everything with `v0.7.x` tags is compatible with each other. We will have a series of minor version updates prior to `v1.0.0`, where we offer strong backwards compatibility guarantees. In particular, work has begun in the `cosmwasm` library on `v0.8`, which will change many internal APIs, in order to allow adding other languages for writing smart contracts. We hope to stabilize much of this well before `v1`, but we are still in the process of learning from real-world use-cases

## Encoding

We use standard cosmos-sdk encoding (amino) for all sdk Messages. However, the message body sent to all contracts, as well as the internal state is encoded using JSON. Cosmwasm allows arbitrary bytes with the contract itself responsible for decodng. For better UX, we often use `json.RawMessage` to contain these bytes, which enforces that it is valid json, but also give a much more readable interface.  If you want to use another encoding in the contracts, that is a relatively minor change to wasmd but would currently require a fork. Please open in issue if this is important for your use case.

## Quick Start

```
make install
make test
```
if you are using a linux without X or headless linux, look at [this article](https://ahelpme.com/linux/dbusexception-could-not-get-owner-of-name-org-freedesktop-secrets-no-such-name) or [#31](https://github.com/cosmwasm/wasmd/issues/31#issuecomment-577058321).

To set up a single node testnet, [look at the deployment documentation](./docs/deploy-testnet.md).

If you want to deploy a whole cluster, [look at the network scripts](./networks/README.md).

## Dockerized

We provide a docker image to help with test setups. There are two modes to use it

Build:  `docker build  -t cosmwasm/wasmd:manual .`  or pull from dockerhub

### Dev server

Bring up a local node with a test account containing tokens

This is just designed for local testing/CI - DO NOT USE IN PRODUCTION

```sh
docker volume rm -f wasmd_data

# pass password (one time) as env variable for setup, so we don't need to keep typing it
# add some addresses that you have private keys for (locally) to give them genesis funds
docker run --rm -it \
    -e PASSWORD=xxxxxxxxx \
    --mount type=volume,source=wasmd_data,target=/root \
    cosmwasm/wasmd-demo:latest ./setup.sh cosmos1pkptre7fdkl6gfrzlesjjvhxhlc3r4gmmk8rs6

# This will start both wasmd and wasmcli rest-server, only wasmcli output is shown on the screen
docker run --rm -it -p 26657:26657 -p 26656:26656 -p 1317:1317 \
    --mount type=volume,source=wasmd_data,target=/root \
    cosmwasm/wasmd-demo:latest ./run_all.sh

# view wasmd logs in another shell
docker run --rm -it \
    --mount type=volume,source=wasmd_data,target=/root,readonly \
    cosmwasm/wasmd-demo:latest ./logs.sh
```

### CI

For CI, we want to generate a template one time and save to disk/repo. Then we can start a chain copying the initial state, but not modifying it. This lets us get the same, fresh start every time.

```sh
# Init chain and pass addresses so they are non-empty accounts
rm -rf ./template && mkdir ./template
docker run --rm -it \
    -e PASSWORD=xxxxxxxxx \
    --mount type=bind,source=$(pwd)/template,target=/root \
    cosmwasm/wasmd-demo:latest ./setup.sh cosmos1pkptre7fdkl6gfrzlesjjvhxhlc3r4gmmk8rs6

sudo chown -R $(id -u):$(id -g) ./template

# FIRST TIME
# bind to non-/root and pass an argument to run.sh to copy the template into /root
# we need wasmd_data volume mount not just for restart, but also to view logs
docker volume rm -f wasmd_data
docker run --rm -it -p 26657:26657 -p 26656:26656 -p 1317:1317 \
    --mount type=bind,source=$(pwd)/template,target=/template \
    --mount type=volume,source=wasmd_data,target=/root \
    cosmwasm/wasmd-demo:latest ./run_all.sh /template

# RESTART CHAIN with existing state
docker run --rm -it -p 26657:26657 -p 26656:26656 -p 1317:1317 \
    --mount type=volume,source=wasmd_data,target=/root \
    cosmwasm/wasmd-demo:latest ./run_all.sh

# view wasmd logs in another shell
docker run --rm -it \
    --mount type=volume,source=wasmd_data,target=/root,readonly \
    cosmwasm/wasmd-demo:latest ./logs.sh
```


