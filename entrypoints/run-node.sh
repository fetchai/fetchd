#!/usr/bin/env bash

set -eo pipefail

if [ "${MAINTENANCE}" == "true" ];
then
  echo "Sleeping for 24 hours"
  sleep 86400
else
  VALIDATOR_STATE_FILE="/root/.fetchd/data/priv_validator_state.json"
  VALIDATOR_STATE_DIR="/root/.fetchd/data"

  # Copy readonly values from configmap dir to /root/.fetchd/config
  if [ ! -d "/root/.fetchd/config" ]; then
    mkdir -p "/root/.fetchd/config"
  fi
  cp /root/wasm-temp-config/* /root/.fetchd/config/
  cp /root/secret-temp-config/* /root/.fetchd/config/
  chmod 644 /root/.fetchd/config/*

  # Set the correct moniker in the config.toml
  sed -i "s/tempmoniker/$MONIKER/g" ~/.fetchd/config/config.toml
  sed -i "s/tempexternal/$P2PADDRESS/g" ~/.fetchd/config/config.toml

  # Genesis usually comes from /root/wasm-temp-config/genesis.json, which is populated from a configmap
  # Some genesis might not fit there (when over 1MB), so as an alternative, OVERWRITE_GENESIS_URL environment 
  # can be specified to pull the genesis from the URL it contains.
  if [ -n "${OVERWRITE_GENESIS_URL}" ];
  then
      echo "Overwritting genesis.json from ${OVERWRITE_GENESIS_URL}"
      curl -o ~/.fetchd/config/genesis.json "${OVERWRITE_GENESIS_URL}"
      if [ $? -ne 0 ]; then
          echo "failed to download genesis.json"
          exit 1
      fi
  fi

  ##
  ## Create priv_validator_state.json if it does not exist
  ##
  if [ ! -f "$VALIDATOR_STATE_FILE" ];
  then
    mkdir -p "$VALIDATOR_STATE_DIR"
    echo "$VALIDATOR_STATE_FILE not found"
    echo "---"
    echo "Creating priv_validator_state.json"
    echo '{' >> "$VALIDATOR_STATE_FILE"
    echo '  "height": "0",' >> "$VALIDATOR_STATE_FILE"
    echo '  "round": 0,' >> "$VALIDATOR_STATE_FILE"
    echo '  "step": 0' >> "$VALIDATOR_STATE_FILE"
    echo '}' >> "$VALIDATOR_STATE_FILE"
    chmod 666 "$VALIDATOR_STATE_FILE"
  fi

  fetchd start
fi
