
# CLI - Staking query

## Querying the current staking holdings of the validators

The following command can be used to retrieve the current staking holdings of all validators:

```bash
fetchcli query staking validators
```

On `agent-land` network, this will produce an output similar to the following, describing the status of all the existing validators:

```text
- |
  operatoraddress: fetchvaloper12xd8rgp2u0cwp8lnj2ndulpzad3y9m9f2r8lsx
  conspubkey: fetchvalconspub1zcjduepqpzmqj5g42cz4w2vjnt00z0jcalcl6sn4p2fynys6xjfy8xgzyzmswn296q
  jailed: false
  status: 2
  tokens: "196120000000000000000"
  delegatorshares: "200100000000000000000.000000000000000000"
  description:
    moniker: Bourne
    identity: ""
    website: ""
    security_contact: ""
    details: ""
  unbondingheight: 1714003
  unbondingcompletiontime: 2021-01-11T15:16:04.979759165Z
  commission:
    commission_rates:
      rate: "0.050000000000000000"
      max_rate: "0.100000000000000000"
      max_change_rate: "0.010000000000000000"
    update_time: 2020-09-01T18:30:00Z
  minselfdelegation: "50"
  producingblocks: false
- |
  operatoraddress: fetchvaloper1vf5wsxjkmjk4uv3nm2zjplw0y2f96rsjw8k7gv
  conspubkey: fetchvalconspub1zcjduepqnrnfqys8p78reemexhyg95f78n4xdat65u3xqnuvhmyz6qkx8gvqyxhf7n
  jailed: false
  status: 2
  tokens: "607240000000000000000"
  delegatorshares: "644902293967714528462.192013593882752757"
  description:
    moniker: Hunt
    identity: ""
    website: ""
    security_contact: ""
    details: ""
  unbondingheight: 1729520
  unbondingcompletiontime: 2021-01-12T15:35:20.868117941Z
  commission:
    commission_rates:
      rate: "0.050000000000000000"
      max_rate: "0.100000000000000000"
      max_change_rate: "0.010000000000000000"
    update_time: 2020-09-01T18:30:00Z
  minselfdelegation: "50"
  producingblocks: false
...
```

To obtain the same informations for a single validator, use the following command, providing the `operatoraddress` of the validator.

```bash
fetchcli query staking validator fetchvaloper12xd8rgp2u0cwp8lnj2ndulpzad3y9m9f2r8lsx --trust-node
```

A delegator will be particularly interested in the following keys:
- `commission/commission_rates/rate`: The commission rate on revenue charged to any delegator by the validator.
- `commission/commission_rates/max_change_rate`: The maximum daily increase of the validator's commission. This parameter cannot be changed by the validator operator.
- `commission/commission_rates/max_rate`: The maximum commission rate this validator can charge. This parameter cannot be changed by the validator operator.
- `minselfdelegation`: Minimum amount of Atoms the validator need to have bonded at all time. If the validator's self-bonded stake falls below this limit, their entire staking pool (i.e. all its delegators) will unbond. This parameter exists as a safeguard for delegators. Indeed, when a validator misbehaves, part of their total stake gets slashed. This included the validator's self-delegateds stake as well as their delegators' stake. Thus, a validator with a high amount of self-delegated Atoms has more skin-in-the-game than a validator with a low amount. The minimum self-bond amount parameter guarantees to delegators that a validator will never fall below a certain amount of self-bonded stake, thereby ensuring a minimum level of skin-in-the-game. This parameter can only be increased by the validator operator.

## Query the delegations to a validator

From a validator address, we can retrieve the list of delegations it received:

```bash
fetchcli query staking delegations-to fetchvaloper1cct4fhhksplu9m9wjljuthjqhjj93z0s97p3g7
```

Here is a sample of delegations `Bond` received on `agent-land`

```text
- delegation:
    delegator_address: fetch1xdr9y0e9z6t5mm6t0y7y9n3hlrgt6ctdzp6sjc
    validator_address: fetchvaloper1cct4fhhksplu9m9wjljuthjqhjj93z0s97p3g7
    shares: "28000000000000000000.000000000000000000"
  balance:
    denom: atestfet
    amount: "26898217871152184017"
- delegation:
    delegator_address: fetch1g0ktdwr4t9jj70tyyvdjf633n5tgh9e9lhctxl
    validator_address: fetchvaloper1cct4fhhksplu9m9wjljuthjqhjj93z0s97p3g7
    shares: "510145592112805227.559325919981657686"
  balance:
    denom: atestfet
    amount: "490071688666363222"
- delegation:
    delegator_address: fetch140a7u7hl3su8efz96qeu8gkzuslv4qaahhll69
    validator_address: fetchvaloper1cct4fhhksplu9m9wjljuthjqhjj93z0s97p3g7
    shares: "612174710535362191.906454201536168749"
  balance:
    denom: atestfet
    amount: "588086026399631946"
...
```

## Query a users rewards from their delegations

After having delegated some tokens to a validator, the user is eligible to a share of the rewards the validator collect.

To retrieve all the outstanding rewards for an address, issue the following command:

```bash
fetchcli query distribution rewards fetch1xdr9y0e9z6t5mm6t0y7y9n3hlrgt6ctdzp6sjc
```

This address having delegated tokens to 3 validators on `agent-land`, it produces the following output:

```text
rewards:
- validator_address: fetchvaloper12xd8rgp2u0cwp8lnj2ndulpzad3y9m9f2r8lsx
  reward:
  - denom: atestfet
    amount: "36507421287940463196.704176667226831706"
- validator_address: fetchvaloper108hhutnylgz09acca2ljde8dp6huhsu67hn8v7
  reward:
  - denom: atestfet
    amount: "15234979036841842481.187320204232358848"
- validator_address: fetchvaloper1cct4fhhksplu9m9wjljuthjqhjj93z0s97p3g7
  reward:
  - denom: atestfet
    amount: "19967189966226410985.053631472920723989"
total:
- denom: atestfet
  amount: "71709590291008716662.945128344379914543"
```

Rewards can also be filtered for a given validator, like `Bourne` here:

```bash
fetchcli query distribution rewards fetch1xdr9y0e9z6t5mm6t0y7y9n3hlrgt6ctdzp6sjc fetchvaloper12xd8rgp2u0cwp8lnj2ndulpzad3y9m9f2r8lsx
```

we now get only the reward from this validator:

```text
- denom: atestfet
  amount: "36507421287940463196.704176667226831706"
```

## Withdrawing rewards

In order to transfer rewards to the wallet, the following command can be used:

```bash
fetchcli tx distribution withdraw-rewards fetchvaloper12xd8rgp2u0cwp8lnj2ndulpzad3y9m9f2r8lsx --from myKey
```

It requires the validator address from where the reward is withdrawn, and the name of the account private key having delegated tokens to the validator.

When having delegated tokens to multiple validators, all rewards can be claimed in a single command:

```bash
fetchcli tx distribution withdraw-all-rewards --from myKey
```
