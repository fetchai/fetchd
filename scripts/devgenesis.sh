#!/bin/bash

set -euo pipefail

usage() {
  cat <<EOF
Usage: $(basename "${BASH_SOURCE[0]}") path/to/exported/genesis.json [node_home_dir]

This script updates an exported genesis from a running chain
to be used to run on a single validator local node.
It will take the first validator and jail all the others
and replace the validator pubkey and the nodekey with the one 
found in the node_home_dir folder

if unspecified, node_home_dir default to the ~/.fetchd/ folder. 
this folder must exists and contains the files created by the "fetchd init" command.

The updated genesis will be written under node_home_dir/config/genesis.json, allowing
the local chain to be started with:

fetchd --home <node_home_dir> unsafe-reset-all && fetchd --home <node_home_dir> start

EOF
  exit
}

if [ "$#" -lt 1 ]; then
    usage
fi

GEN_FILE="$1"
if [ ! -f "${GEN_FILE}" ]; then
    usage
fi

OUT_HOMEDIR="${2:-~/.fetchd/}"

if [ ! -f "${OUT_HOMEDIR}/config/priv_validator_key.json" ]; then
    echo "cannot find file ${OUT_HOMEDIR}/config/priv_validator_key.json"
    exit 1
fi

echo "Found ${OUT_HOMEDIR}/config/priv_validator_key.json"

GEN_CONTENT=$(cat "${GEN_FILE}")

NEW_HEXADDR=$(jq -r '.address' "${OUT_HOMEDIR}/config/priv_validator_key.json")
echo "- new address: ${NEW_HEXADDR}"
NEW_PUBKEY=$(jq -r '.pub_key.value' "${OUT_HOMEDIR}/config/priv_validator_key.json")
echo "- new pubkey: ${NEW_PUBKEY}"
NEW_TMADDR=$(fetchd --home "${OUT_HOMEDIR}" tendermint show-address)
echo "- new tendermint address: ${NEW_TMADDR}"

VAL_INFOS=$(jq '.app_state.staking.validators[0]' <(echo "${GEN_CONTENT}"))
if [ -z "${VAL_INFOS}" ]; then
    echo "genesis file does not contains any validators"
    exit 1
fi

VAL_OPADDR=$(jq -r '.operator_address' <(echo "${VAL_INFOS}"))
VAL_PUBKEY=$(jq -r '.consensus_pubkey.key' <(echo "${VAL_INFOS}"))
VAL_ADDR=$(jq -r --arg VAL_PUBKEY "${VAL_PUBKEY}" '.validators[] | select(.pub_key.value == $VAL_PUBKEY).address' <(echo "${GEN_CONTENT}"))

#
# replace selected validator by current node one 
#
echo "Replacing validator ${VAL_OPADDR}..."
GEN_CONTENT=$(sed "s#${VAL_ADDR}#${NEW_HEXADDR}#g" <(echo "${GEN_CONTENT}"))
GEN_CONTENT=$(sed "s#${VAL_PUBKEY}#${NEW_PUBKEY}#g" <(echo "${GEN_CONTENT}"))

#
# set .app_state.slashing.signing_infos to contains only our validator signing infos
#
echo "Updating signing infos..."
GEN_CONTENT=$(jq --arg TMADDR "${NEW_TMADDR}" '.app_state.slashing.signing_infos = [
{
    "address": $TMADDR,
    "validator_signing_info": {
        "address": $TMADDR,
        "index_offset": "0",
        "jailed_until": "1970-01-01T00:00:00Z",
        "missed_blocks_counter": "0",
        "start_height": "0",
        "tombstoned": false
    }
}]' <(echo "${GEN_CONTENT}"))


#
# update bonded and not bonded pools value to make invariant checks happy
#
# pool addresses are static:
# bonded pool: fetch1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3xxqtmq
# not bonded pool: fetch1tygms3xhhs3yv487phx3dw4a95jn7t7ljxu6d5
#

echo "Updating bonded and not bonded token pool values..."

VALTOKENS=$(jq -r '.tokens' <(echo "${VAL_INFOS}"))
VALPOWER=$(echo "${VALTOKENS}/10^18" | bc)

BONDED_TOKENS=$(jq -r '.app_state.bank.balances[] |  select(.address == "fetch1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3xxqtmq").coins[] | select(.denom == "afet").amount' <(echo "${GEN_CONTENT}"))
NOTBONDED_TOKENS=$(jq -r '.app_state.bank.balances[] |  select(.address == "fetch1tygms3xhhs3yv487phx3dw4a95jn7t7ljxu6d5").coins[] | select(.denom == "afet").amount' <(echo "${GEN_CONTENT}"))
NOTBONDED_TOKENS=$(echo "${NOTBONDED_TOKENS} + ${BONDED_TOKENS} - ${VALTOKENS}" | bc)

GEN_CONTENT=$(jq -r --arg TOKENS "${VALTOKENS}" '(.app_state.bank.balances[] | select(.address == "fetch1fl48vsnmsdzcv85q5d2q4z5ajdha8yu3xxqtmq") | (.coins[] | select(.denom == "afet")).amount) = $TOKENS' <(echo "${GEN_CONTENT}"))
GEN_CONTENT=$(jq -r --arg TOKENS "${NOTBONDED_TOKENS}" '(.app_state.bank.balances[] | select(.address == "fetch1tygms3xhhs3yv487phx3dw4a95jn7t7ljxu6d5") | (.coins[] | select(.denom == "afet")).amount) = $TOKENS' <(echo "${GEN_CONTENT}"))

#
# removes all .validators but the one we work with
#
echo "Removing other validators from initchain..."
GEN_CONTENT=$(jq --arg HEXADDR "${NEW_HEXADDR}" '.validators = [(.validators[] | select(.address == $HEXADDR))]' <(echo "${GEN_CONTENT}"))

#
# set .app_state.staking.last_validator_powers to contains only our validator 
#
echo "Updating last voting power..."
GEN_CONTENT=$(jq --arg POWER "${VALPOWER}" --arg ADDR "${VAL_OPADDR}" '.app_state.staking.last_validator_powers = [{
    "address": $ADDR,
    "power": $POWER
}]' <(echo "${GEN_CONTENT}"))

#
# jail everyone but our validator
#
echo "Jail other validators..."
GEN_CONTENT=$(jq --arg ADDR "${VAL_OPADDR}" '(.app_state.staking.validators[] | select(.operator_address != $ADDR ) | .status) = "BOND_STATUS_UNBONDING"' <(echo "${GEN_CONTENT}"))
GEN_CONTENT=$(jq --arg ADDR "${VAL_OPADDR}" '(.app_state.staking.validators[] | select(.operator_address != $ADDR ) | .jailed) = true' <(echo "${GEN_CONTENT}"))


echo "${GEN_CONTENT}" | jq > "${OUT_HOMEDIR}/config/genesis.json"
echo "Done! Wrote new genesis at ${OUT_HOMEDIR}/config/genesis.json"
echo "You can now start the chain:"
echo
echo "fetchd --home ${OUT_HOMEDIR} unsafe-reset-all && fetchd --home ${OUT_HOMEDIR} start"
echo
