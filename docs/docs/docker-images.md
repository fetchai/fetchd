# Docker Images

For every Fetchd release a corresponding docker images is released. The full list of images and tags can be found at:

[https://hub.docker.com/r/fetchai/fetchd](https://hub.docker.com/r/fetchai/fetchd)

# Supported Tags & Dockerfiles

Current Versions:

* [fetchhub, 0.7, 0.7.x](https://github.com/fetchai/fetchd/blob/v0.7.2/Dockerfile) (used for fetchhub mainnetv2 networks)
* [beaconworld, 0.6, 0.6.x](https://github.com/fetchai/fetchd/blob/v0.6.4/Dockerfile) (used for beaconworld test networks)
* [agentworld, 0.2, 0.2.x](https://github.com/fetchai/fetchd/blob/v0.2.7/Dockerfile) (used for agentworld test networks)

# Quick Reference

* Support documentation available at the [Fetch.ai Documenation](https://docs.fetch.ai/)

* Where to file issues: [https://github.com/fetchai/fetchd/issues](https://github.com/fetchai/fetchd/issues)

* Maintained by: [The Fetch.ai Ledger Team](https://github.com/fetchai/fetchd)

* Supported architectures: amd64

# How to use this image

## Starting a node

To start a fetchd node on `fetchhub`, simply run:

```
docker run \
    -v $(pwd)/fetchd:/root/.fetchd \
    -v $(pwd)/fetchcli:/root/.fetchcli \
    -e MONIKER=<your_moniker> \
    fetchai/fetchd:fetchhub
```

Replace `<your_moniker>` with any name of your choice to identify your node on the network.
The 2 volumes are used to export the fetchd and fetchcli keys, configuration and runtime data out of the container. 

To connect to other networks, simply swap the `fetchai/fetchd:fetchhub` image with `fetchai/fetchd:beaconworld` or `fetchai/fetchd:agentworld`.

## Using fetchcli

You can invoke the fetchcli binary directly by setting the `--entrypoint` parameter:

```
docker run \
    -v $(pwd)/fetchd:/root/.fetchd \
    -v $(pwd)/fetchcli:/root/.fetchcli \
    --entrypoint fetchcli \
    fetchai/fetchd:fetchhub <commands>
```

### Example

List keys: 

```
docker run \
    -v $(pwd)/fetchd:/root/.fetchd \
    -v $(pwd)/fetchcli:/root/.fetchcli \
    --entrypoint fetchcli \
    fetchai/fetchd:fetchhub keys list
```

List validators:

```
docker run \
    -v $(pwd)/fetchd:/root/.fetchd \
    -v $(pwd)/fetchcli:/root/.fetchcli \
    --entrypoint fetchcli \
    fetchai/fetchd:fetchhub query staking validators
```

## Customize fetchd

You can specify additionnal environment variables to connect to a custom network:

```
docker run \
    -v $(pwd)/fetchd:/root/.fetchd \
    -v $(pwd)/fetchcli:/root/.fetchcli \
    -e MONIKER=<your_moniker> \
    -e RPC_ENDPOINT=<rpc_endpoint>
    -e SEEDS=<seed_ids>
    fetchai/fetchd:fetchhub
```

Or fine tune the flags passed to `fetchd start`:

```
docker run \
    -v $(pwd)/fetchd:/root/.fetchd \
    -v $(pwd)/fetchcli:/root/.fetchcli \
    -e MONIKER=<your_moniker> \
    fetchai/fetchd:fetchhub --p2p.pex --p2p.laddr 0.0.0.0:2665
```

# License

View [license information](https://github.com/fetchai/fetchd/blob/master/LICENSE) for the software contained in this image.

As with all Docker images, these likely also contain other software which may be under other licenses (such as Bash, etc from the base distribution, along with any direct or indirect dependencies of the primary software being contained).

As for any pre-built image usage, it is the image user's responsibility to ensure that any use of this image complies with any relevant licenses for all software contained within.
