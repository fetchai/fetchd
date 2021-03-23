# Setting up a Validator Node

This guide assumes that you have successfuly [installed](../../building/), configured and [connected](../../joining-a-testnet/) your validator to the desired network.

## Creating a validator

To create a validator on the network you will need to send a transaction to the network bonding / staking your FET tokens. This process registers you as a validator and if you are one of the chosen validators you will start to produce blocks.

```bash
fetchcli tx staking create-validator \
  --amount=<the amount to bond> \
  --pubkey=$(fetchd tendermint show-validator) \
  --moniker="choose a moniker" \
  --chain-id=<chain_id> \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="<the min self delegation>" \
  --from=<key_name>
```

** Beacon World Example **

In the case of Beaconworld the minimum self delegation amount (the amount of FET that the validator themselves must stake) is 1000TESTFET (i.e. 1000000000000000000000atestfet).

Before trying to create a validator you should verify that you have this required minimum number of tokens available beforehand. The easiest way to do this is via the [CLI](../../cli-tokens/).

Here is an sample of a typical command line command that will register the node as running the validator.

```bash
fetchcli tx staking create-validator \
  --amount=1000000000000000000000atestfet \
  --pubkey=fetchvalconspub1zcjduepqa8ldmfnt6avct9x5h7269jjfmv2l22pnejezz37vh7syajfyku6stlj0s9 \
  --moniker="my-test-validator" \
  --chain-id=beaconworld-3 \
  --commission-rate="0.10" \
  --commission-max-rate="0.20" \
  --commission-max-change-rate="0.01" \
  --min-self-delegation="1000000000000000000000" \
  --from=beaonworld-test-key
```

## Editing a validator

Over time it is possible that validators will want to adjust various settings about their nodes. This can be simple things like the associated website for a validator or more consequential actions like altering the commission rate.

In either case, should a validator choose to make this update they would send an "edit-validator" transaction to the network. These can be created in a similar way to the "create-validator" transactions as shown below:

```bash
fetchcli tx staking edit-validator
  --moniker="choose a moniker" \
  --website="https://fetch.ai" \
  --details="To infinity and beyond!" \
  --chain-id=<chain_id> \
  --commission-rate="0.10"
  --from=<key_name>
```

## Unbonding a validator


When / if a validator wants to stop being a validator for any reason, they can unbond some or all of their staked FET. This is done with the following command.

```bash
./build/fetchcli tx staking unbond \
  <validator operator address> \
  <amount to remove> \
  --from <key name>
```

An example of the command is given in the following example:

```bash
./build/fetchcli tx staking unbond \
  fetchvaloper1jqqwdch3jmzlmj4tjfn67s3sqm9elkd3wrpspf \
  1000000000000000000000atestfet \
  --from beaonworld-test-key
```

** Note **

Validators' obligations continue until the end of the aeon (which is typically 100 blocks or ~8 minutes depending on the configuration). It is therefore important that after a validator unbonds their stake they must leave their node up and running for 2 complete aeons before switching off. Failure to do so is treated as malicious behaviour and will result in stake being slashed.
