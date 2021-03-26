#!/bin/bash

docker build -t fetchai/fetchd:test --target hub .
docker build -t local-fetchd:test --target gcr .