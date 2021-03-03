
# CLI - Staking query

## Querying the current staking holdings of the validators

The following command can be used to retrieve the current staking holdings of all validators:

```bash
fetchcli query staking validators
```

On `beacon-world` network, this will produce an output similar to the following, describing the status of all the existing validators:

```text
- |
  operatoraddress: fetchvaloper1z72rph6l5j6ex83n4urputykawcqg6t98xul2w
  conspubkey: fetchvalconspub1zcjduepq3urw6c6u0zvqmde4vr4gmy56nnq57shdhg56jynpu8n3s74hrm0q0mzqrx
  jailed: false
  status: 2
  tokens: "1000000000000000000000"
  delegatorshares: "1000000000000000000000.000000000000000000"
  description:
    moniker: validator5
    identity: ""
    website: ""
    security_contact: ""
    details: ""
  unbondingheight: 0
  unbondingcompletiontime: 1970-01-01T00:00:00Z
  commission:
    commission_rates:
      rate: "0.050000000000000000"
      max_rate: "0.100000000000000000"
      max_change_rate: "0.010000000000000000"
    update_time: 2021-02-12T12:41:25.579730119Z
  minselfdelegation: "1000000000000000000000"
  producingblocks: true
- |
  operatoraddress: fetchvaloper1ysc8n5uspv4698nyk8u75lx98uu92zt7m3udw8
  conspubkey: fetchvalconspub1zcjduepqmxr8gmcs6pwuxpsma264ax59wxtxd3vchrcv2c06deq9986kwt3s0wsk6n
  jailed: false
  status: 2
  tokens: "1000000000000000000000"
  delegatorshares: "1000000000000000000000.000000000000000000"
  description:
    moniker: validator2
    identity: ""
    website: ""
    security_contact: ""
    details: ""
  unbondingheight: 0
  unbondingcompletiontime: 1970-01-01T00:00:00Z
  commission:
    commission_rates:
      rate: "0.050000000000000000"
      max_rate: "0.100000000000000000"
      max_change_rate: "0.010000000000000000"
    update_time: 2021-02-03T13:00:00Z
  minselfdelegation: "1000000000000000000000"
  producingblocks: true
...
```

To obtain the same informations for a single validator, use the following command, providing the `operatoraddress` of the validator.

```bash
fetchcli query staking validator fetchvaloper1z72rph6l5j6ex83n4urputykawcqg6t98xul2w
```

A delegator will be particularly interested in the following keys:
- `commission/commission_rates/rate`: The commission rate on revenue charged to any delegator by the validator.
- `commission/commission_rates/max_change_rate`: The maximum daily increase of the validator's commission. This parameter cannot be changed by the validator operator.
- `commission/commission_rates/max_rate`: The maximum commission rate this validator can charge. This parameter cannot be changed by the validator operator.
- `minselfdelegation`: Minimum amount of Atoms the validator need to have bonded at all time. If the validator's self-bonded stake falls below this limit, their entire staking pool (i.e. all its delegators) will unbond. This parameter exists as a safeguard for delegators. Indeed, when a validator misbehaves, part of their total stake gets slashed. This included the validator's self-delegateds stake as well as their delegators' stake. Thus, a validator with a high amount of self-delegated Atoms has more skin-in-the-game than a validator with a low amount. The minimum self-bond amount parameter guarantees to delegators that a validator will never fall below a certain amount of self-bonded stake, thereby ensuring a minimum level of skin-in-the-game. This parameter can only be increased by the validator operator.

## Query the delegations made to a validator

From a validator address, we can retrieve the list of delegations it received:

```bash
fetchcli query staking delegations-to fetchvaloper1z72rph6l5j6ex83n4urputykawcqg6t98xul2w
```

Here is a sample of delegations `validator5` received on `beacon-world`:

```text
- delegation:
    delegator_address: fetch1z72rph6l5j6ex83n4urputykawcqg6t9zzruef
    validator_address: fetchvaloper1z72rph6l5j6ex83n4urputykawcqg6t98xul2w
    shares: "1000000000000000000000.000000000000000000"
  balance:
    denom: atestfet
    amount: "1000000000000000000000"
- delegation:
    delegator_address: fetch15fn3meky8ktfry3qm73xkpjckzw4dazxpfx34m
    validator_address: fetchvaloper1z72rph6l5j6ex83n4urputykawcqg6t98xul2w
    shares: "100000.000000000000000000"
  balance:
    denom: atestfet
    amount: "100000"
```

## Query the redelegations

Delegators can choose to redelegate the tokens they already delegated from one validator to another. Redelegation takes effect immediately, without any waiting period, but the tokens can't be redelegated until the initial redelegation complete, in ~21 days (the unlocking time is indicated by the `redelegationentry/completion_time` field in the outputs below).


To obtains the list of redelegations made from a validator, use the following command:

```bash
fetchcli query staking redelegations-from fetchvaloper1z72rph6l5j6ex83n4urputykawcqg6t98xul2w
```

Which produce an output similar to the following, where delegator `fetch15fn3meky8ktfry3qm73xkpjckzw4dazxpfx34m` issued 2 redelegations from `fetchvaloper1z72rph6l5j6ex83n4urputykawcqg6t98xul2w` to `fetchvaloper122veneudkzyalay6gusvrhhpp0560mparpanvu`:

```text
fetchcli query staking redelegations-from fetchvaloper1z72rph6l5j6ex83n4urputykawcqg6t98xul2w
- redelegation:
    delegator_address: fetch15fn3meky8ktfry3qm73xkpjckzw4dazxpfx34m
    validator_src_address: fetchvaloper1z72rph6l5j6ex83n4urputykawcqg6t98xul2w
    validator_dst_address: fetchvaloper122veneudkzyalay6gusvrhhpp0560mparpanvu
    entries: []
  entries:
  - redelegationentry:
      creation_height: 291037
      completion_time: 2021-03-24T14:24:38.973444629Z
      initial_balance: "50000"
      shares_dst: "50000.000000000000000000"
    balance: "50000"
  - redelegationentry:
      creation_height: 291133
      completion_time: 2021-03-24T14:33:43.425472866Z
      initial_balance: "10000"
      shares_dst: "10000.000000000000000000"
    balance: "10000"
```

Similarly, the list of redelegations issued by a delegator can be obtained with the following:

```bash
fetchcli query staking redelegations fetch15fn3meky8ktfry3qm73xkpjckzw4dazxpfx34m
```

## Query the user rewards

After having delegated some tokens to a validator, the user is eligible to a share of the rewards the validator collect.

To retrieve all the outstanding rewards for an address, issue the following command:

```bash
fetchcli query distribution rewards fetch15fn3meky8ktfry3qm73xkpjckzw4dazxpfx34m
```

This address having delegated tokens to 2 validators on `beacon-world`, it produces the following output:

```text
rewards:
- validator_address: fetchvaloper1z72rph6l5j6ex83n4urputykawcqg6t98xul2w
  reward:
  - denom: atestfet
    amount: "0.000000000000200000"
- validator_address: fetchvaloper1ysc8n5uspv4698nyk8u75lx98uu92zt7m3udw8
  reward:
  - denom: atestfet
    amount: "0.000000000001000000"
total:
- denom: atestfet
  amount: "0.000000000001200000"
```

Rewards can also be filtered for a given validator, like `validator5` here:

```bash
fetchcli query distribution rewards fetch15fn3meky8ktfry3qm73xkpjckzw4dazxpfx34m fetchvaloper1z72rph6l5j6ex83n4urputykawcqg6t98xul2w
```

we now get only the reward from this validator:

```text
- denom: atestfet
  amount: "0.000000000000200000"
```

## Withdrawing rewards

In order to transfer rewards to the wallet, the following command can be used:

```bash
fetchcli tx distribution withdraw-rewards fetchvaloper1z72rph6l5j6ex83n4urputykawcqg6t98xul2w --from myKey
```

It requires the validator address from where the reward is withdrawn, and the name of the account private key having delegated tokens to the validator.

When having delegated tokens to multiple validators, all rewards can be claimed in a single command:

```bash
fetchcli tx distribution withdraw-all-rewards --from myKey
```
