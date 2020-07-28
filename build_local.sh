#!/bin/bash

docker build -t fetchai/wasmd:test -f Dockerfile.hub .
docker build -t local-wasmd:test -f Dockerfile.gcr . --build-arg VERSION=test