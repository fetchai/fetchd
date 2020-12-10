#!/usr/bin/env bash
set -e

if [ ! $BOOTSTRAP == "" ];
then
  echo "Fetching configuration for $BOOTSTRAP network"
  bootstrapurl="https://raw.githubusercontent.com/fetchai/networks-$BOOTSTRAP/master/bootstrap/boostrap.json"
  export CHAINID=$(curl -sS $bootstrapurl  | (jq -r .chainid))
  export ARGS=$(curl -sS $bootstrapurl  | (jq -r .args))
fi

CHECK_FILE="/root/secret-temp-config/config/config.toml"

if [ -f "$CHECK_FILE" ];
then
  echo "Provided node configuration files in /root/secret-temp-config"
  mkdir /root/.fetchd
  cp -R /root/secret-temp-config/* /root/.fetchd/
  curl https://rpc-${CHAINID}.fetch.ai/genesis? | jq .result.genesis > ~/.fetchd/config/genesis.json
  sed -i  's/allow_duplicate_ip = false/allow_duplicate_ip = true/' ~/.fetchd/config/config.toml
  fetchd start --p2p.laddr tcp://0.0.0.0:26656 --rpc.laddr tcp://0.0.0.0:26657 ${ARGS}
else
  echo "Node configuration files have not been provided"
  echo "Moniker : ${MONIKER}"
  echo "Chain ID: ${CHAINID}"
  echo "Args    : ${ARGS}"

  fetchd init ${MONIKER} --chain-id ${ARGS}
  curl https://rpc-${CHAINID}.fetch.ai/genesis? | jq .result.genesis > ~/.fetchd/config/genesis.json
  sed -i  's/allow_duplicate_ip = false/allow_duplicate_ip = true/' ~/.fetchd/config/config.toml
  fetchd start --p2p.laddr tcp://0.0.0.0:26656 --rpc.laddr tcp://0.0.0.0:26657 ${ARGS}
fi
