# Fetch.ai fetchd Fork

This repository is the fork from the original [fetchd](https://github.com/fetchai/fetchd). It contains Fetch.ai specific updates required for the test networks. For this reason it is versioned independantly. Please refer to the [releases](https://github.com/fetchai/cosmos-sdk/releases) section for the compatiblity with upstream versions.

This repository hosts `Wasmd`, the first implementation of a cosmos zone with wasm smart contracts enabled.

This code was forked from the `cosmos/gaia` repository as a basis and then we added `x/wasm` and cleaned up 
many gaia-specific files. However, the `wasmd` binary should function just like `gaiad` except for the
addition of the `x/wasm` module.

**Note**: Requires [Go 1.13+](https://golang.org/dl/)

## Supported Systems

The supported systems are limited by the dlls created in [`go-cosmwasm`](https://github.com/CosmWasm/go-cosmwasm). In particular, **we only support MacOS and Linux**. 
For linux, the default is to build for glibc, and we cross-compile with CentOS 7 to provide
backwards compatibility for `glibc 2.12+`. This includes all known supported distributions
using glibc (CentOS 7 uses 2.12, obsolete Debian Jessy uses 2.19). 

As of `0.9.0` we support `muslc` Linux systems, in particular **Alpine linux**,
which is popular in docker distributions. Note that we do **not** store the
static `muslc` build in the repo, so you must compile this yourself, and pass `-tags muslc`.
Please look at the [`Dockerfile`](./Dockerfile) for an example of how we build a static Go
binary for `muslc`. (Or just use this Dockerfile for your production setup).


## Stability

**This is alpha software, do not run on a production system.** Notably, we currently provide **no migration path** not even "dump state and restart" to move to future versions. At **beta** we will begin to offer migrations and better backwards compatibility guarantees.

With the `v0.6.0` tag, we entered semver. That means anything with `v0.6.x` tags is compatible with each other, 
and everything with `v0.7.x` tags is compatible with each other. 
We will have a series of minor version updates prior to `v1.0.0`, 
where we offer strong backwards compatibility guarantees. 

We released `v0.10` on  July 31, 2020, and have a `v1.0` feature freeze now.
We will be performing bug fixes, security patches, and performance improvements,
but the API should be stable and compatible with `v1.0`.

Thank you to all projects who have run this code in your testnets and
given feedback to improve stability.

## Encoding

We use standard cosmos-sdk encoding (amino) for all sdk Messages. However, the message body sent to all contracts, 
as well as the internal state is encoded using JSON. Cosmwasm allows arbitrary bytes with the contract itself 
responsible for decodng. For better UX, we often use `json.RawMessage` to contain these bytes, which enforces that it is
valid json, but also give a much more readable interface.  If you want to use another encoding in the contracts, that is
a relatively minor change to wasmd but would currently require a fork. Please open in issue if this is important for 
your use case.

## Quick Start

```
# Ubuntu/Debian based linux distro dependencies
apt install swig libboost-all-dev libgmp-dev

cd ~/Downloads
wget https://github.com/herumi/mcl/archive/v1.05.tar.gz
tar xvf v1.05.tar.gz 
cd mcl-1.05
make install
ldconfig

# OS X dependencies
brew install swig boost gmp

cd ~/Downloads
wget https://github.com/herumi/mcl/archive/v1.05.tar.gz
tar xvf v1.05.tar.gz 
cd mcl-1.05
make install

```
After installing the reuired dependencies, install golang and execute the following commands:
```
make install
make test
```
if you are using a linux without X or headless linux, look at [this article](https://ahelpme.com/linux/dbusexception-could-not-get-owner-of-name-org-freedesktop-secrets-no-such-name) or [#31](https://github.com/fetchai/fetchd/issues/31#issuecomment-577058321).

To set up a single node testnet, [look at the deployment documentation](./docs/deploy-testnet.md).

If you want to deploy a whole cluster, [look at the network scripts](./networks/README.md).

## Dockerized

We provide a docker image to help with test setups. There are two modes to use it

Build: `docker build -t fetchai/fetchd:latest .`  or pull from dockerhub

### Dev server

Bring up a local node with a test account containing tokens

This is just designed for local testing/CI - do not use these scripts in production.
Very likely you will assign tokens to accounts whose mnemonics are public on github.

```sh
docker volume rm -f fetchd_data

# pass password (one time) as env variable for setup, so we don't need to keep typing it
# add some addresses that you have private keys for (locally) to give them genesis funds
docker run --rm -it \
    -e PASSWORD=xxxxxxxxx \
    --mount type=volume,source=fetchd_data,target=/root \
    fetchai/fetchd:latest ./setup.sh cosmos1pkptre7fdkl6gfrzlesjjvhxhlc3r4gmmk8rs6

# This will start both fetchd and fetchcli rest-server, only fetchcli output is shown on the screen
docker run --rm -it -p 26657:26657 -p 26656:26656 -p 1317:1317 \
    --mount type=volume,source=fetchd_data,target=/root \
    fetchai/fetchd:latest ./run_all.sh

# view fetchd logs in another shell
docker run --rm -it \
    --mount type=volume,source=fetchd_data,target=/root,readonly \
    fetchai/fetchd:latest ./logs.sh
```

### CI

For CI, we want to generate a template one time and save to disk/repo. Then we can start a chain copying the initial state, but not modifying it. This lets us get the same, fresh start every time.

```sh
# Init chain and pass addresses so they are non-empty accounts
rm -rf ./template && mkdir ./template
docker run --rm -it \
    -e PASSWORD=xxxxxxxxx \
    --mount type=bind,source=$(pwd)/template,target=/root \
    fetchai/fetchd:latest ./setup.sh cosmos1pkptre7fdkl6gfrzlesjjvhxhlc3r4gmmk8rs6

sudo chown -R $(id -u):$(id -g) ./template

# FIRST TIME
# bind to non-/root and pass an argument to run.sh to copy the template into /root
# we need fetchd_data volume mount not just for restart, but also to view logs
docker volume rm -f fetchd_data
docker run --rm -it -p 26657:26657 -p 26656:26656 -p 1317:1317 \
    --mount type=bind,source=$(pwd)/template,target=/template \
    --mount type=volume,source=fetchd_data,target=/root \
    fetchai/fetchd:latest ./run_all.sh /template

# RESTART CHAIN with existing state
docker run --rm -it -p 26657:26657 -p 26656:26656 -p 1317:1317 \
    --mount type=volume,source=fetchd_data,target=/root \
    fetchai/fetchd:latest ./run_all.sh

# view fetchd logs in another shell
docker run --rm -it \
    --mount type=volume,source=fetchd_data,target=/root,readonly \
    fetchai/fetchd:latest ./logs.sh
```

## Runtime flags

We provide a number of variables in `app/app.go` that are intended to be set via `-ldflags -X ...`
compile-time flags. This enables us to avoid copying a new binary directory over for each small change
to the configuration.

Available flags:

* `-X github.com/CosmWasm/wasmd/app.CLIDir=.coral` - set the config directory for the cli (default `~/.fetchcli`) 
* `-X github.com/CosmWasm/wasmd/app.NodeDir=.corald` - set the config/data directory for the node (default `~/.fetchd`)
* `-X github.com/CosmWasm/wasmd/app.Bech32Prefix=coral` - set the bech32 prefix for all accounts (default `cosmos`)
* `-X github.com/CosmWasm/wasmd/app.ProposalsEnabled=true` - enable all x/wasm governance proposals (default `false`)
* `-X github.com/CosmWasm/wasmd/app.EnableSpecificProposals=MigrateContract,UpdateAdmin,ClearAdmin` - 
    enable a subset of the x/wasm governance proposal types (overrides `ProposalsEnabled`)

Examples:

* [`fetchd`](./Makefile#L50-L55) is a generic, permissionless version using the `cosmos` bech32 prefix
* [`gaiaflex`](./Makefile#L78-L87) is a generic, *permissioned* version using the `cosmos` bech32 prefix
* [`coral`](./Makefile#L63-L71) is a permissionless version designed for a specific testnet, with a `coral` bech32 prefix

## Genesis Configuration

@alpe we should document all the genesis config for x/wasm even more.

Tip: if you want to lock this down to a permisisoned network, the following script can edit the genesis file
to only allow permissioned use of code upload or instantiating. (Make sure you set `app.ProposalsEnabled=true`
in this binary):

`sed -i 's/permission": "Everybody"/permission": "Nobody"/'  .../config/genesis.json`
