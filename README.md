# Fetch ai fetchd 

## BLS signature (curve BLS12-381) support for accounts

### Running demo with local FetchD node
Install fetchd
```
git clone https://github.com/kitounliu/fetchd.git
cd fetchd
git checkout bls
make install

```


Setup FetchD node
```s
# Initialise test chain

# Clean state 
rm -rf ~/.fetch*

# SETUP LOCAL CHAIN
# Initialize the genesis.json file that will help you to bootstrap the network
fetchd init --chain-id=testing testing

export KEYRING="--keyring-backend test --keyring-dir $HOME/.fetchd"

# Create a key to hold your validator account
fetchd keys add validator $KEYRING

# Add validator to genesis block and give him some stake
fetchd add-genesis-account $(fetchd keys show validator -a $KEYRING) 100000000000000000000000stake
							    
# Generate the transaction that creates your validator
fetchd gentx validator  10000000000000000000000stake --chain-id testing $KEYRING

# Add the generated bonding transaction to the genesis file
fetchd collect-gentxs

# Enable rest-api
sed -i '/^\[api\]$/,/^\[/ s/^enable = false/enable = true/' ~/.fetchd/config/app.toml

# Configure
fetchd config chain-id testing
fetchd config keyring-backend test

# run the node
fetchd start
```


To create an account using bls public/private key
```s
# Create users and give them some stake
fetchd keys add alice $KEYRING --algo bls12381
fetchd tx bank send $(fetchd keys show validator -a $KEYRING) $(fetchd keys show alice -a $KEYRING) 10000000stake --chain-id testing

fetchd keys add bob $KEYRING --algo bls12381
fetchd tx bank send $(fetchd keys show alice -a $KEYRING) $(fetchd keys show bob -a $KEYRING) 1000000stake  --chain-id testing


# Check if funds were transfered from validator to alice and from alice to bob
fetchd query bank balances $(fetchd keys show -a alice $KEYRING)
fetchd query bank balances $(fetchd keys show -a bob $KEYRING)
```



