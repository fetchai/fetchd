#!/bin/bash

docker build -t fetchai/fetchd:test -f Dockerfile.hub .
docker build -t local-fetchd:test -f Dockerfile.gcr . --build-arg VERSION=test