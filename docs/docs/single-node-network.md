# Running a Single Node Network

Especially for things like contract development, it can be very useful to be able to run a single node network for
testing. This document will outline the steps that are required in order to configure a `fetchd` network of 1 node.

## Network Setup

These steps only need to be done once in order to setup the local network.

**Step 1 - Build the ledger from source**

Follow the [build instructions](../building/) in order to compile the latest version of the ledger.

**Step 2 - Remove any existing networks**

Since we are starting a new network we need to remove any local files that we have in our system from a previous network

   `rm -Rf ~/.fetchd`

**Step 3 - Create an initial genesis**

Create the initial genesis file (`~/.fetchd/config/genesis.json`) with the following command:

   `fetchd init --chain-id localnet-1 my-local-node-name`

* `localnet-1` is the chain id 
* `my-local-node-name` is the moniker for the node

If you want to make any updates to the genesis, it is a good opportunity to make these updates now.

**Step 4 - Create your validator key**

In the following steps we will need to create the public/private keypair for our node. 

To create a new key called "validator" use the following command.

   `fetchd keys add validator`

* `validator` is the name of the key in the keyring

For more information checkout the complete [documentation on keys](../cli-keys/).

**Step 5 - Adding the validator to the network**

To set the initial state for the network use the following command. This allocates `100000000000000000000` `stake` tokens
to the validator which can be bonded. 

   `fetchd add-genesis-account validator 100000000000000000000stake`

`stake` is the default test token denomination in the cosmos ecosystem, but you could use `afet`, `BTC` etc.

**Step 6 - Generating a validator transaction**

To get your validator to sign the genesis block (and to agree that this is the correct genesis starting point) use the
following command.

   `fetchd gentx validator 100000000000000000000stake --chain-id localnet-1`

* `validator` here is the name that you have given to the key

**Step 7 - Building the complete genesis**

To build final genesis configuration for the network run the following command

   `fetchd collect-gentxs`

After running this command the network is successfully configured and you have computed the final genesis configuration
for the network.

## Running the local node

To run the network use the following command.

    `fetchd start`

## Resetting the network

Often you will want to clear out all the data from the network and start again. To do that in a local network simply
run the following command:

    `fetchd tendermint unsafe-reset-all`

This resets the chain back to genesis, you **DO NOT** need to perform the network setup steps again. After running this
command you can simply run the `fetchd start` command again.

