#!/usr/bin/env bash
sleep ${TIMESLEEPNODE}

if [ $MAINTENANCE == "True" ];
then
  echo "Sleeping for 24 hours"
  sleep 86400
else
  if $EPHEMERALNODE
  then
    echo "Creating ephermeral node"
    sleep 5

    fetchd init $MONIKER --chain-id ${CHAINID}

    curl http://sentry-lb:26657/genesis? | jq .result.genesis > ~/.fetchd/config/genesis.json

    sed -i  's/allow_duplicate_ip = false/allow_duplicate_ip = true/' ~/.fetchd/config/config.toml
    sed -i  's/prometheus = false/prometheus = true/' ~/.fetchd/config/config.toml
    sed -i  "s/external_address.*/external_address = \"$P2PADDRESS\"/" ~/.fetchd/config/config.toml
    # Modify mempool size
    sed -i  's/size = 5000/size = 50000/' ~/.fetchd/config/config.toml
    sed -i  's/cache_size = 10000/cache_size = 50000/' ~/.fetchd/config/config.toml

    fetchd start --p2p.laddr tcp://127.0.0.1:26656 --rpc.laddr tcp://127.0.0.1:26657 ${P2PPEX} ${PERSPEERS} ${PRIVPEERS} ${SEEDMODE} ${SEEDS} ${PRUNING}
  else
    VALIDATOR_STATE_FILE="/root/.fetchd/data/priv_validator_state.json"
    VALIDATOR_STATE_DIR="/root/.fetchd/data"

    # Copy readonly values from configmap dir to /root/.gaiad/config
    mkdir -p /root/.fetchd/config
    cp /root/wasm-temp-config/* /root/.fetchd/config/
    cp /root/secret-temp-config/* /root/.fetchd/config/
    chmod 644 /root/.fetchd/config/*

    # Set the correct moniker in the config.toml
    sed -i "s/tempmoniker/$MONIKER/g" ~/.fetchd/config/config.toml
    sed -i "s/tempexternal/$P2PADDRESS/g" ~/.fetchd/config/config.toml

    ##
    ## Create priv_validator_state.json if it does not exist
    ##
    if [ ! -f "$VALIDATOR_STATE_FILE" ];
    then
      mkdir $VALIDATOR_STATE_DIR
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

    fetchd start --p2p.laddr tcp://127.0.0.1:26656 --rpc.laddr tcp://127.0.0.1:26657 ${P2PPEX} ${PERSPEERS} ${PRIVPEERS} ${SEEDMODE} ${SEEDS} ${PRUNING}
  fi
fi
