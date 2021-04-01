# Building the Ledger

## Prerequisites

- Go 1.14+ (installation instructions available [here](https://golang.org/dl/))

### Ubuntu 20.04 / Debian 10

```bash
# Run the following command if using Ubuntu
sudo apt-get install libgmp-dev swig

# Download and install the MCL libraries
cd ~
wget https://github.com/herumi/mcl/archive/v1.05.tar.gz
tar xvf v1.05.tar.gz
cd mcl-1.05
sudo make install
sudo ldconfig
```

### MacOS

```bash
# Run the following command if using OS X
brew install swig gmp

# Download and install the MCL libraries
cd ~
wget https://github.com/herumi/mcl/archive/v1.05.tar.gz
tar xvf v1.05.tar.gz
cd mcl-1.05
sudo make install
sudo ldconfig
```

## Building the code

Download the latest released version from github and build the project using the following commands:

    git clone https://github.com/fetchai/fetchd.git && cd fetchd

Then build the code with the command:

    make build

This will generate the following binaries:

- `./build/fetchcli` - This is the command line client that is useful for interacting with the network
- `./build/fetchd` - This is the block chain node daemon and can be configured to join the network

For non-developer users we recommend that the user installs the binaries into their system. This can be done with the following command:

    sudo make install

This will install the binaries in the directory specified by your `$GOBIN` environment variable.

### Boost Dependencies (Only for beaconworld versions `v0.5x` and `v0.6.x`)

Currently the code requires that the user compiles the code with at least version 1.67 of Boost Serialisation library. Failure to do so will mean that users will not be able to sync with the blockchain. This limitation will be resolved in the near future (with the Boost dependency being removed completely).

To verify which libraries you have linked against use the following commands

**Ubuntu 20.04 / Debian 10**

```bash
ldd ./build/fetchd
```

**MacOS**

```bash
otool -L ./build/fetchd
```
