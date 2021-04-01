# Docker Images

For every Fetchd release a corresponding docker images is released. The full list of images and tags can be found at:

[https://hub.docker.com/r/fetchai/fetchd](https://hub.docker.com/r/fetchai/fetchd)

# Supported Tags & Dockerfiles

Current Versions:

* [0.7.2, 0.7](https://github.com/fetchai/fetchd/blob/v0.7.2/Dockerfile) (used for fetchhub mainnetv2 networks)
* [0.6.4, 0.6](https://github.com/fetchai/fetchd/blob/v0.6.4/Dockerfile) (used for beaconworld test networks)
* [0.2.7, 0.2](https://github.com/fetchai/fetchd/blob/v0.2.7/Dockerfile) (used for agentworld test networks)

# Quick Reference

* Support documentation available at the [Fetch.ai Documenation](https://docs.fetch.ai/)

* Where to file issues: [https://github.com/fetchai/fetchd/issues](https://github.com/fetchai/fetchd/issues)

* Maintained by: [The Fetch.ai Ledger Team](https://github.com/fetchai/fetchd)

* Supported architectures: amd64

# How to use this image

## Connecting to a test network

Connecting a node to the test network is easy. In its simpliest configuration the docker container can be run with just a couple of environment variables as shown below:

    docker run -e MONIKER=<insert node name here> -e NETWORK=<network name> fetchai/fetchd:0.5

However, users will almost certainly want to mount a storage volume into the container so that the node does not need to resync from genesis everytime. This can be done by adding the following volume path:

    docker run -e MONIKER=<insert node name here> -e NETWORK=<network name> -v /path/for/data:/root/.fetchd fetchai/fetchd:0.5

For example connecting a node to the beaconworld testnet can be done with the following command:

    docker run -e MONIKER=my-first-fetch-node -e NETWORK=beaconworld -v $(pwd)/my-first-fetch-node-data:/root/.fetchd fetchai/fetchd:0.5

# License

View [license information](https://github.com/fetchai/fetchd/blob/master/LICENSE) for the software contained in this image.

As with all Docker images, these likely also contain other software which may be under other licenses (such as Bash, etc from the base distribution, along with any direct or indirect dependencies of the primary software being contained).

As for any pre-built image usage, it is the image user's responsibility to ensure that any use of this image complies with any relevant licenses for all software contained within.
