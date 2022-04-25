# Multisig keys

This feature of `fetchd` allows users to securely control keys in a number of configurations. Using a threshold number K of maximum N keys, a user or group of users can set the minimum number of keys required to sign a transaction. Some examples of these configurations allow some useful features such as the choice of a spare key, where only one key is required to sign (K=1) but there are two keys available to do so. Another more complex example configuration is set out below.

## Creating a multisig key

The following represents the syntax and argument layout of the `fetchd` command to create a multisig key.

```
# Create a simple multisig key with a threshold of 1 as default
fetchd keys add <multisig_key_name> --multisig <list_of_key_names>

# Creating a multisig key with a higher threshold, K
fetchd keys add <multisig_key_name> --multisig <list_of_key_names> --multisig-threshold <threshold integer K>
```

### Example instantiation of a multisig key

This example represents a shared multisig key that could be used within a business amongst three account holders - where at least two of three (K=2) must sign off on each transaction.

```
# Create the three keys owned by the separate account holders
fetchd keys add fred
fetchd keys add ted
fetchd keys add ned

# Create the multisig key from keys above
fetchd keys add business_key --multisig fred,ted,ned --multisig-threshold 2
```

You will need the address of the business_key later in the example. Here just a reminder how to get it:

```
fetchd keys show -a business_key
```

## Signing and broadcasting multisig transactions

Transactions must be signed and broadcast before they are carried out.

In order to sign a multisig transaction, the transaction itself must not be immediately broadcast; but instead, the keyholders must each sign until a minimum threshold K signatures are present.

_For this example we will be performing the transaction on the [Dorado](https://explore-dorado.fetch.ai/) network and therefore will be using `atestfet` as the denomination, and a gas price of 1000000000atestfet (this should be changed depending on the actual currency and network used)._

### Multisig transaction example

```
# Create a key to represent a vendor that the business must pay
fetchd keys add vendor

# Generate a transaction as an output file to be signed by
# the keyholders, 'ted' and 'fred' in this example
fetchd tx bank send <business_key address> <vendor address> 1000atestfet --gas 90000 --gas-prices 1000000000atestfet --generate-only > transfer.json

# you'll get "account <address of business_key> not found" error for missing funds
# add funds to <address of business_key> using block explorer or by eg
curl -XPOST -H 'Content-Type: application/json' -d '{"address":"<address of business_key>"}' https://faucet-dorado.t-v2-london-c.fetch-ai.com/api/v3/claims

# This transaction file (transfer.json) is then made available for
# the first keyholder to sign, 'fred'
fetchd tx sign transfer.json --chain-id dorado-1 --from fred --multisig <address of business_key> > transfer_fredsigned.json

# This is repeated for 'ted'
fetchd tx sign transfer.json --chain-id dorado-1 --from ted --multisig <address of business_key> > transfer_tedsigned.json

# These two files are then collated together and used as inputs to the
# multisign command to create a fully signed transaction
fetchd tx multisign transfer.json business_key transfer_fredsigned.json transfer_tedsigned.json > signed_transfer.json

# Now that the transaction is fully signed, it may be broadcast
fetchd tx broadcast signed_transfer.json

# Now display the result of the transaction and confirm that the vendor has
# received payment
fetchd query bank balances <address of vendor>
```

It is important to note that this method of signing transactions can apply to all types of transaction.

### Other multisig transaction examples

```
# In order to create a staking transaction using a multisig key
# the same process as above can be used with the output file of this command
fetchd tx staking delegate <fetchvaloper address> 10000atestfet --from <address of business_key> --gas 200000 --gas-prices 1000000000atestfet --generate-only > stake.json

# The following command can also be used to create a withdrawal transaction for the
# rewards from staking when using a multisig key - this too must be signed as before
fetchd tx distribution withdraw-all-rewards --from <address of business_key> --gas 150000 --gas-prices 1000000000atestfet --generate-only > withdrawal.json
```
