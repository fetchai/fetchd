#!/bin/sh
#set -o errexit -o nounset -o pipefail

PASSWORD=${PASSWORD:-1234567890}
STAKE=${STAKE_TOKEN:-ustake}
FEE=${FEE_TOKEN:-ucosm}
CHAIN_ID=${CHAIN_ID:-testing}
MONIKER=${MONIKER:-node001}

fetchd init --chain-id "$CHAIN_ID" "$MONIKER"
# staking/governance token is hardcoded in config, change this
sed -i "s/\"stake\"/\"$STAKE\"/" "$HOME"/.fetchd/config/genesis.json
if ! fetchcli keys show validator; then
  (echo "$PASSWORD"; echo "$PASSWORD") | fetchcli keys add validator
fi
# hardcode the validator account for this instance
echo "$PASSWORD" | fetchd add-genesis-account validator "1000000000$STAKE,1000000000$FEE"
# (optionally) add a few more genesis accounts
for addr in "$@"; do
  echo $addr
  fetchd add-genesis-account "$addr" "1000000000$STAKE,1000000000$FEE"
done
# submit a genesis validator tx
(echo "$PASSWORD"; echo "$PASSWORD"; echo "$PASSWORD") | fetchd gentx --name validator --amount "250000000$STAKE"
fetchd collect-gentxs
