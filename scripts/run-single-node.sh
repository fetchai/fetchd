#!/bin/bash

# You should run a `make install` before running this script to ensure the
# fetch binary is up to date

# Create the validator keys
fetchcli keys add fooValidator

# Init the chain config files
fetchd init --chain-id=testing testing

# Clear any old state
fetchd unsafe-reset-all

# Create a genesis account using the fooValidator public key/address
fetchd add-genesis-account $(fetchcli keys show fooValidator -a) 1000000000000000000000stake

# Create a transaction which will make a validator with the fooValidator public key
fetchd gentx --name=fooValidator

# Combine all previous 
fetchd collect-gentxs

fetchd start
