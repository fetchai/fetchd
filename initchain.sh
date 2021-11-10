#!/bin/bash

set -ex

CHAIN_ID="devchain"
MONIKER="devnode"

HOME="/tmp/${CHAIN_ID}"
FETCHD="./build/fetchd --home ${HOME}"

if [ -d "/tmp/${CHAIN_ID}" ]; then
    rm -rf "/tmp/${CHAIN_ID}"
fi

${FETCHD} init ${MONIKER} --chain-id ${CHAIN_ID}
sed -i 's/stake/atestfet/g' ${HOME}/config/genesis.json

${FETCHD} config keyring-backend test
${FETCHD} config chain-id ${CHAIN_ID}
${FETCHD} config node http://localhost:26657

ALICE_ADDR=$(${FETCHD} keys add alice --output json | jq -r '.address')
BOB_ADDR=$(${FETCHD} keys add bob --output json | jq -r '.address')

${FETCHD} add-genesis-account ${ALICE_ADDR} 10000000000000000000000atestfet
${FETCHD} add-genesis-account ${BOB_ADDR} 10000000000000000000000atestfet

${FETCHD} gentx alice 1000000000000000000atestfet --chain-id ${CHAIN_ID} \
    --moniker "devvalidator" \
    --commission-max-change-rate 0.01 \
    --commission-max-rate 1.0 \
    --commission-rate 0.07

${FETCHD} collect-gentxs

${FETCHD} start