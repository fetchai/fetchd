#!/usr/bin/env bash
set -ueEo pipefail

FETCH_ADDR_REGEX="^fetch[0-9a-z]{39}"
RECONCILIATION_CONTRACT="${1-""}"

if [[ ! $RECONCILIATION_CONTRACT =~ $FETCH_ADDR_REGEX ]]; then
  echo -e "usage: query_all_registrations.sh CONTRACT
    Where CONTRACT is the address of the reconciliation contract to query."
  exit 1
fi

# Ensure reconciliation contract is paused
RECONCILIATION_IS_PAUSED=$(fetchd query wasm contract-state smart "$RECONCILIATION_CONTRACT" '{"query_pause_status":{}}' --output json | jq -r '.data.paused')
if [[ "$RECONCILIATION_IS_PAUSED" == "false" ]]; then
  echo "Aborting!: reconciliation contract \"$RECONCILIATION_CONTRACT\" is not paused"
  exit 1
fi

# Query reconciliation registrations
fetchd query wasm contract-state smart "$RECONCILIATION_CONTRACT" '{"query_all_registrations":{}}' --output json | jq '.data.registrations'