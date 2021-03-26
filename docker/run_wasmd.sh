#!/bin/sh

if test -n "$1"; then
    # need -R not -r to copy hidden files
    cp -R "$1/.fetchd" /root
    cp -R "$1/.fetchcli" /root
fi

mkdir -p /root/log
fetchd start --rpc.laddr tcp://0.0.0.0:26657 --trace
