# Building the Ledger

## Prerequisites

- Go 1.14+ (installation instructions available [here](https://golang.org/dl/]))

## Building the code

Download the latest released version from github and build the project using the following commands:

    git clone https://github.com/fetchai/fetchd.git -b release/v0.2.x && cd fetchd

Then build the code with the command:

    make build

This will generate the following binaries:

- `./build/fetchcli` - This is the command line client that is useful for interacting with the network
- `./build/fetchd` - This is the block chain node daemon and can be configured to join the network

For non-developer users we recommend that the user installs the binaries into their system. This can be done with the following command:

    sudo make install

This will install the binaries in the directory specified by your `$GOBIN` environment variable.
