FROM golang:1.14-buster

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
    swig && \
  wget https://github.com/herumi/mcl/archive/v1.05.tar.gz && \
  tar xvf v1.05.tar.gz && cd mcl-1.05 && \
  make install && ldconfig

WORKDIR /workspace/fetchd
COPY . .

RUN make go-mod-cache && make build
