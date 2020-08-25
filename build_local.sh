#!/bin/bash

docker build -t fetchai/wasmd:test --target hub .
docker build -t local-wasmd:test --target gcr .
