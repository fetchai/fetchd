#!/bin/sh
#set -euo pipefail

fetchcli rest-server --laddr tcp://0.0.0.0:1317 --trust-node --cors
