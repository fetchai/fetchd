#!/usr/bin/env bash

set -eo pipefail

if [ "${MAINTENANCE}" == "true" ];
then
  echo "Sleeping for 24 hours"
  sleep 86400
else
  NODE_HOME="/root/.fetchd"
  VALIDATOR_STATE_FILE="${NODE_HOME}/data/priv_validator_state.json"
  VALIDATOR_STATE_DIR="${NODE_HOME}/data"
  WASM_STATE_DIR="${NODE_HOME}/wasm"
  
  # when EXTERNAL_DATA_DIR is provided, it will replace default state directories
  # by symlinks, effectively moving their effective location to EXTERNAL_DATA_DIR.
  if [ -d "${EXTERNAL_DATA_DIR}" ]; then
    mkdir -p ${EXTERNAL_DATA_DIR}/{data,wasm}
    test -L ${VALIDATOR_STATE_DIR} || ln -s "${EXTERNAL_DATA_DIR}/data/" "${VALIDATOR_STATE_DIR}"
    test -L ${WASM_STATE_DIR} || ln -s "${EXTERNAL_DATA_DIR}/wasm/" "${WASM_STATE_DIR}"
  fi

  # Copy readonly values from configmap dir to ${NODE_HOME}/config
  if [ ! -d "${NODE_HOME}/config" ]; then
    mkdir -p "${NODE_HOME}/config"
  fi
  
  cp /root/wasm-temp-config/* ${NODE_HOME}/config/  || true
  cp /root/secret-temp-config/* ${NODE_HOME}/config/ || true

  # Set the correct moniker in the config.toml
  sed -i "s/tempmoniker/$MONIKER/g" ${NODE_HOME}/config/config.toml
  sed -i "s/tempexternal/$P2PADDRESS/g" ${NODE_HOME}/config/config.toml

  # Genesis usually comes from /root/wasm-temp-config/genesis.json, which is populated from a configmap
  # Some genesis might not fit there (when over 1MB), so as an alternative, OVERWRITE_GENESIS_URL environment 
  # can be specified to pull the genesis from the URL it contains.
  if [ -n "${OVERWRITE_GENESIS_URL}" ];
  then
      echo "Overwritting genesis.json from ${OVERWRITE_GENESIS_URL}"
      curl -o ${NODE_HOME}/config/genesis.json "${OVERWRITE_GENESIS_URL}"
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
    mkdir -p "$VALIDATOR_STATE_DIR" || true
    echo "$VALIDATOR_STATE_FILE not found"
    echo "---"
    echo "Creating priv_validator_state.json"
    echo '{' >> "$VALIDATOR_STATE_FILE"
    echo '  "height": "0",' >> "$VALIDATOR_STATE_FILE"
    echo '  "round": 0,' >> "$VALIDATOR_STATE_FILE"
    echo '  "step": 0' >> "$VALIDATOR_STATE_FILE"
    echo '}' >> "$VALIDATOR_STATE_FILE"
  fi

  fetchd start
fi
