#!/bin/sh
#set -euo pipefail

mkdir -p /root/log
touch /root/log/fetchd.log
./run_fetchd.sh $1 >> /root/log/fetchd.log &

sleep 4
echo Starting Rest Server...

./run_rest_server.sh
