FROM golang:buster

WORKDIR /workspace/mcl
RUN apt-get update && \
  apt-get install -y \
    curl \
    wget \
    jq \
    file \
    make \
    git \
    libgmp-dev \
    gcc \
    g++ \
    swig \
    libboost-serialization-dev && \
  git clone https://github.com/herumi/mcl && cd mcl && make install && ldconfig

WORKDIR /workspace/cosmos-sdk
COPY . .
RUN make go-mod-cache && make build
