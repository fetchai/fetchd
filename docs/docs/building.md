# Building the Ledger

## Prerequisites

- Go 1.14+ (installation instructions available [here](https://golang.org/dl/]))

### Ubuntu

```bash
# Run the following command if using Ubuntu
sudo apt-get install libgmp-dev swig libboost-all-dev

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
brew install swig boost gmp

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
