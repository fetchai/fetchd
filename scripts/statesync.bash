#!/bin/bash
# microtick and bitcanna contributed significantly here.
# Pebbledb state sync script for evmos
set -uxe

# Set Golang environment variables.
export GOPATH=~/go
export PATH=$PATH:~/go/bin

# Install Juno with pebbledb 
go mod edit -replace github.com/tendermint/tm-db=github.com/notional-labs/tm-db@136c7b6
go mod tidy
go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=pebbledb' -tags pebbledb ./...

# NOTE: ABOVE YOU CAN USE ALTERNATIVE DATABASES, HERE ARE THE EXACT COMMANDS
# go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=rocksdb' -tags rocksdb ./...
# go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=badgerdb' -tags badgerdb ./...
# go install -ldflags '-w -s -X github.com/cosmos/cosmos-sdk/types.DBBackend=boltdb' -tags boltdb ./...

# Initialize chain.
fetchd init test

# Get Genesis
wget https://storage.googleapis.com/fetch-ai-mainnet-v2-genesis/genesis-fetchhub4.json
mv genesis-fetchhub4.json ~/.fetchd/config/genesis.json

# Get "trust_hash" and "trust_height".
INTERVAL=1000
LATEST_HEIGHT="$(curl -s https://evmos-rpc.polkachu.com/block | jq -r .result.block.header.height)"
BLOCK_HEIGHT="$((LATEST_HEIGHT-INTERVAL))"
TRUST_HASH="$(curl -s "https://evmos-rpc.polkachu.com/block?height=$BLOCK_HEIGHT" | jq -r .result.block_id.hash)"

# Print out block and transaction hash from which to sync state.
echo "trust_height: $BLOCK_HEIGHT"
echo "trust_hash: $TRUST_HASH"

# Export state sync variables.
export FETCHD_STATESYNC_ENABLE=true
export FETCHD_P2P_MAX_NUM_OUTBOUND_PEERS=200
export FETCHD_STATESYNC_RPC_SERVERS="https://rpc-juno-ia.notional.ventures:443,https://juno-rpc.polkachu.com:443"
export FETCHD_STATESYNC_TRUST_HEIGHT=$BLOCK_HEIGHT
export FETCHD_STATESYNC_TRUST_HASH=$TRUST_HASH

# Fetch and set list of seeds from chain registry.
FETCHD_P2P_SEEDS="$(curl -s https://raw.githubusercontent.com/cosmos/chain-registry/master/fetchhub/chain.json | jq -r '[foreach .peers.seeds[] as $item (""; "\($item.id)@\($item.address)")] | join(",")')"
export FETCHD_P2P_SEEDS

# Start chain.
fetchd start --x-crisis-skip-assert-invariants --db_backend pebbledb
