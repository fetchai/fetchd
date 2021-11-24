# BLS signatures

The BLS algorithm can be selected when creating new keys and signing transactions. BLS supported keys are particularly useful for the increasing the efficiency of multi-signature transactions at the cost of simplicity in verification.

It affords the users a shorter and yet still robust grouping of each signature from party members without the sacrifice of security on each multi-signature transaction.

## Creating BLS keys
Creating BLS keys is straightforward in comparison to normal key instantiation with the addition of one extra parameter to the command. This example will show the additional flag required in comparison to a key with the standard algorithm; with 'bls12381' in place of the default 'secp256k1'.

### Example
```
# Create a normal key
fetchd keys add Ron

# Create a key capable of BLS signed transactions
fetchd keys add Tom_BLS --algo bls12381

# 'Ron' can be assumed to define implicitly --algo secp256k1 by default
```

## BLS Transactions and signatures
After creating this BLS key, transactions can be carried out between two keys using the different algorithms.


### Example
*Ensure that the 'Ron' key has some funds before performing this example.*

```
# Perform a normal transfer of funds from Ron to Tom_BLS
fetchd tx bank send <address_of_Ron> <address_of_Tom_BLS> 1000test

# This should provide a breakdown of the transaction parameters, including the gas fees
# Keep note of these fees

# Check that these funds have been transferred to Tom_BLS
fetchd query bank balaces <address_of_Tom_BLS>

# Perform a BLS signed transaction from Tom_BLS to Ron
fetchd tx bank send <address_of_Tom_BLS> <address_of_Ron> 1000test

# Compare the difference between information printed from each transaction and observe
# the difference in gas costs (?)

# Now assure funds were successfully transferred back to Ron through a BLS signed transaction 
fetchd query bank balaces <address_of_Tom_BLS>
fetchd query bank balaces <address_of_Ron>
```
