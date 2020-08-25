FROM golang:1.14-buster as builder

# Set up dependencies
ENV PACKAGES jq curl wget jq file make git

RUN apt-get update && \
    apt-get install -y $PACKAGES

WORKDIR /cosmwasm

COPY . .

RUN make install

# ##################################

FROM debian:buster as hub

# Set up dependencies
ENV PACKAGES jq curl

RUN apt-get update && \
    apt-get install -y $PACKAGES

COPY --from=builder /go/pkg/mod/github.com/\!cosm\!wasm/go-cosmwasm@v*/api/libgo_cosmwasm.so /usr/lib/libgo_cosmwasm.so
COPY --from=builder /go/bin/wasmcli /usr/bin/wasmcli
COPY --from=builder /go/bin/wasmd /usr/bin/fwasm
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

COPY ./entrypoints/run-node.sh /usr/bin/run-node.sh
COPY ./entrypoints/run-server.sh /usr/bin/run-server.sh
