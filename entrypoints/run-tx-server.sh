#!/usr/bin/env bash
sleep ${TIMESLEEPLCD}

fetchcli config chain-id ${CHAINID}
fetchcli config output json
fetchcli config indent true
fetchcli config trust-node ${TRUSTNODE}

fetchcli config node ${NODE}

fetchcli tx fmtd --port 8090