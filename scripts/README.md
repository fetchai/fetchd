# Running a private beacon land testnet

This document outlines the steps that are required to run a beacon land (fetchd) testnet, highlighting the differences from vanilla tendermint/the cosmos sdk.

## Quickstart (locally)
To run a network consisting of a single node there is the script ./scripts/run-single-node.sh - this will create a single validator network running locally.

To run:
```
$ make install
$ ./scripts/run-single-node.sh
```

The script will ask for a password at several points. This is as the keys for the validator are password protected in your keychain.

The files for the network will be created at ~/.fetchd/ - so this needs to be deleted to restart the network. Alternatively, it is possible to run 

```
$ fetchd unsafe-reset-all
```

This should only be done when you are certain of what you are doing, as it will delete the data permanently.

Once the run script has been run for the first time, the network can continue running with the command 
```
$ fetchd start
```

The output should look like the following:

```
I[2020-11-04|15:44:37.916] starting ABCI with Tendermint                module=main
E[2020-11-04|15:44:38.225] Updated with new DKG ID 0                    module=slotProtocol
I[2020-11-04|15:44:43.380] Executed block                               module=state height=5 validTxs=0 invalidTxs=0
I[2020-11-04|15:44:43.393] Committed state                              module=state height=5 txs=0 appHash=5...E entropy=NotPresent nextAeonStart=-1
I[2020-11-04|15:44:48.477] Executed block                               module=state height=6 validTxs=0 invalidTxs=0
I[2020-11-04|15:44:48.489] Committed state                              module=state height=6 txs=0 appHash=D...8 entropy=NotPresent nextAeonStart=-1
I[2020-11-04|15:44:53.576] Executed block                               module=state height=7 validTxs=0 invalidTxs=0
I[2020-11-04|15:44:53.589] Committed state                              module=state height=7 txs=0 appHash=2...2 entropy=NotPresent nextAeonStart=-1
```

As can be seen, the blockchain has the concept of entropy and ‘aeons’. Entropy is fetch’s extension to the chain which generates a cryptographically random number on every block (also known as the random beacon). The generation of the entropy requires all of the validators to participate, so is not necessarily always available (for example as can be seen here during network start). 

The chain should be able to be queried on your localhost machine by going to:

http://localhost:26657 

In a web browser. To see the entropy, you can query any block which has generated it (since 
there is only one validator here once it starts generating entropy it won’t fail to generate again)

http://localhost:26657/block?height=46

The result of this query is json which should include block: entropy: group_signature, which is the entropy.

Other information about the chain can be queried using the `fetchcli` tool, which was installed and used during the execution of the quickstart script.

For example, while the chain is running locally, the current validator can be queried with the following command:

```
$ fetchcli query staking validators
```

This will give up to date information on the running validator, whether they are producing blocks etc.

It can be seen that the validator that was created with the key `fooValidator` by viewing the key:
```
$ fetchcli keys show fooValidator -a
```

Which shows the same address (with differing prefix and checksum so the begin and end are different) as the result of querying the validators.

Other useful commands: (while the validator is running)
```
$ fetchcli query account fooValidator - query the funds the validator has
$ fetchcli query block --trust-node
```
