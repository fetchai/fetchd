#!/usr/bin/env bash
sleep ${TIMESLEEPLCD}

wasmcli config chain-id ${CHAINID}
wasmcli config output json
wasmcli config indent true
wasmcli config trust-node ${TRUSTNODE}

wasmcli config node ${NODE}

wasmcli rest-server --laddr tcp://127.0.0.1:1317
