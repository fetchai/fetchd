#!/usr/bin/env bash

set -eo pipefail

if [ -z "${MONIKER}" ]; then
    echo "Please specify the desired moniker for your node (-e MONIKER=<your_moniker>)"
    exit 1
fi

ARGS="$@"
if [ ! -z "${SEEDS}" ]; then
    ARGS="--p2p.seeds ${SEEDS} ${ARGS}"
fi

if [ ! -f ~/.fetchd/config/genesis.json ] && [ ! -z "${RPC_ENDPOINT}" ]; then 
    echo "Setting up fetchd using ${RPC_ENDPOINT}..."
    GENESIS=$(curl -s "${RPC_ENDPOINT}/genesis" | jq '.result.genesis')
    chain_id=$(echo ${GENESIS} | jq -r '.chain_id')

    fetchd init "${MONIKER}" --chain-id "${chain_id}"
    echo "${GENESIS}" > ~/.fetchd/config/genesis.json
    # ensure configuration if correct
    sed -i 's/allow_duplicate_ip = false/allow_duplicate_ip = true/' ~/.fetchd/config/config.toml

    # Configure CLI    
    fetchcli config chain-id ${chain_id}
    fetchcli config trust-node false
    fetchcli config node ${RPC_ENDPOINT}
else
    chain_id=$(cat ~/.fetchd/config/genesis.json | jq -r '.chain_id')
fi

echo "Moniker...: ${MONIKER}"
echo "Chain ID..: ${chain_id}"
echo "RPC Url...: ${RPC_ENDPOINT}"
echo "Args......: ${ARGS}"


# run the daemon
exec fetchd start ${ARGS}
