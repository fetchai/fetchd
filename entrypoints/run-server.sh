#!/usr/bin/env bash
sleep ${TIMESLEEPLCD}

fetchcli config chain-id ${CHAINID}
fetchcli config output json
fetchcli config indent true
fetchcli config trust-node ${TRUSTNODE}

fetchcli config node ${NODE}

fetchcli rest-server --laddr tcp://127.0.0.1:1317 $@

