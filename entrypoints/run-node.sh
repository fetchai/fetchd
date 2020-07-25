#!/usr/bin/env bash
sleep ${TIMESLEEPNODE}

if $EPHEMERALNODE
then
  echo "Creating ephermeral node"
  sleep 5

  wasmd init $MONIKER --chain-id ${CHAINID}

  curl http://sentry-lb:26657/genesis? | jq .result.genesis > ~/.wasmd/config/genesis.json

  sed -i  's/allow_duplicate_ip = false/allow_duplicate_ip = true/' ~/.wasmd/config/config.toml
  sed -i  's/prometheus = false/prometheus = true/' ~/.wasmd/config/config.toml
  sed -i  "s/external_address.*/external_address = \"$P2PADDRESS\"/" ~/.wasmd/config/config.toml
  # Modify mempool size
  sed -i  's/size = 5000/size = 50000/' ~/.wasmd/config/config.toml
  sed -i  's/cache_size = 10000/cache_size = 50000/' ~/.wasmd/config/config.toml

  wasmd start --p2p.laddr tcp://127.0.0.1:26656 --rpc.laddr tcp://127.0.0.1:26657 ${P2PPEX} ${PERSPEERS} ${PRIVPEERS} ${SEEDMODE} ${SEEDS} ${PRUNING}
else
  VALIDATOR_STATE_FILE="/root/.wasmd/data/priv_validator_state.json"

  # Copy readonly values from configmap dir to /root/.gaiad/config
  mkdir -p /root/.wasmd/config
  cp /root/wasm-temp-config/* /root/.wasmd/config/
  cp /root/secret-temp-config/* /root/.wasmd/config/
  chmod 644 /root/.wasmd/config/*

  # Set the correct moniker in the config.toml
  sed -i "s/tempmoniker/$MONIKER/g" ~/.wasmd/config/config.toml
  sed -i "s/tempexternal/$P2PADDRESS/g" ~/.wasmd/config/config.toml

  ##
  ## Create priv_validator_state.json if it does not exist
  ##
  if [ ! -f "$VALIDATOR_STATE_FILE" ];
  then
    echo "$VALIDATOR_STATE_FILE not found"
    echo "---"
    echo "Creating priv_validator_state.json"
    echo '{' >> $VALIDATOR_STATE_FILE
    echo '  "height": "0",' >> $VALIDATOR_STATE_FILE
    echo '  "round": "0",' >> $VALIDATOR_STATE_FILE
    echo '  "step": 0' >> $VALIDATOR_STATE_FILE
    echo '}' >> $VALIDATOR_STATE_FILE
    chmod 666 $VALIDATOR_STATE_FILE
  fi

  wasmd start --p2p.laddr tcp://127.0.0.1:26656 --rpc.laddr tcp://127.0.0.1:26657 ${P2PPEX} ${PERSPEERS} ${PRIVPEERS} ${SEEDMODE} ${SEEDS} ${PRUNING}
fi