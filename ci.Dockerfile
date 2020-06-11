#syntax = docker/dockerfile:experimental
#export DOCKER_BUILDKIT=1

FROM golang:buster

ENV GOPRIVATE="github.com/fetchai/*"

WORKDIR /workspace/mcl
RUN --mount=type=ssh \
  mkdir -m 700 /root/.ssh && \
  touch -m 600 /root/.ssh/known_hosts && \
  git config --global url."git@github.com:".insteadOf https://github.com/ && \
  ssh-keyscan github.com > /root/.ssh/known_hosts

WORKDIR /workspace/cosmos-sdk
COPY . .
RUN --mount=type=ssh \
  make go-mod-cache && \
  make build
