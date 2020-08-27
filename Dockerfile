<<<<<<< HEAD
FROM golang:1.14-buster as builder

# Set up dependencies
ENV PACKAGES jq curl wget jq file make git libgmp-dev gcc g++ swig libboost-serialization-dev

RUN apt-get update && \
    apt-get install -y $PACKAGES && \
    git clone https://github.com/herumi/mcl && cd mcl && make install && ldconfig

WORKDIR /cosmwasm
COPY . .
RUN make install

# ##################################

FROM debian:buster as hub
=======
# docker build . -t cosmwasm/wasmd:latest
# docker run --rm -it cosmwasm/wasmd:latest /bin/sh
FROM cosmwasm/go-ext-builder:0.8.2-alpine AS rust-builder

RUN apk add git

# copy all code into /code
WORKDIR /code
COPY go.* /code/

# download all deps
RUN go mod download github.com/CosmWasm/go-cosmwasm

# build go-cosmwasm *.a and install it
RUN export GO_WASM_DIR=$(go list -f "{{ .Dir }}" -m github.com/CosmWasm/go-cosmwasm) && \
    cd ${GO_WASM_DIR} && \
    cargo build --release --features backtraces --example muslc && \
    mv ${GO_WASM_DIR}/target/release/examples/libmuslc.a /lib/libgo_cosmwasm_muslc.a


# --------------------------------------------------------
FROM cosmwasm/go-ext-builder:0.8.2-alpine AS go-builder

RUN apk add git
# NOTE: add these to run with LEDGER_ENABLED=true
# RUN apk add libusb-dev linux-headers

WORKDIR /code
COPY . /code/

COPY --from=rust-builder /lib/libgo_cosmwasm_muslc.a /lib/libgo_cosmwasm_muslc.a

# force it to use static lib (from above) not standard libgo_cosmwasm.so file
RUN LEDGER_ENABLED=false BUILD_TAGS=muslc make build
# we also (temporarily?) build the testnet binaries here
RUN LEDGER_ENABLED=false BUILD_TAGS=muslc make build-coral
RUN LEDGER_ENABLED=false BUILD_TAGS=muslc make build-gaiaflex

# --------------------------------------------------------
FROM alpine:3.12

COPY --from=go-builder /code/build/wasmd /usr/bin/wasmd
COPY --from=go-builder /code/build/wasmcli /usr/bin/wasmcli

# testnet
COPY --from=go-builder /code/build/coral /usr/bin/coral
COPY --from=go-builder /code/build/corald /usr/bin/corald
COPY --from=go-builder /code/build/gaiaflex /usr/bin/gaiaflex
COPY --from=go-builder /code/build/gaiaflexd /usr/bin/gaiaflexd
>>>>>>> v0.10.0

# Set up dependencies
ENV PACKAGES jq curl libgmpxx4ldbl libboost-serialization1.67.0

RUN apt-get update && \
    apt-get install -y $PACKAGES

COPY --from=builder /go/pkg/mod/github.com/\!cosm\!wasm/go-cosmwasm@v*/api/libgo_cosmwasm.so /usr/lib/libgo_cosmwasm.so
COPY --from=builder /go/bin/fetchcli /usr/bin/fetchcli
COPY --from=builder /go/bin/fetchd /usr/bin/fetchd
COPY --from=builder /usr/local/lib/libmcl.so /usr/lib
COPY entrypoints/entrypoint.sh /usr/bin/entrypoint.sh

VOLUME /root/.fetchd
VOLUME /root/secret-temp-config

ENTRYPOINT [ "/usr/bin/entrypoint.sh" ]
EXPOSE 1317
EXPOSE 26656
EXPOSE 26657
STOPSIGNAL SIGTERM

# ##################################

FROM hub as gcr

<<<<<<< HEAD
COPY ./entrypoints/run-node.sh /usr/bin/run-node.sh
COPY ./entrypoints/run-server.sh /usr/bin/run-server.sh
=======
CMD ["/usr/bin/wasmd version"]
>>>>>>> v0.10.0
