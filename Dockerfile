FROM golang:1.18-buster as builder

# Set up dependencies
ENV PACKAGES jq curl wget jq file make git

RUN apt-get update && \
    apt-get install -y $PACKAGES

WORKDIR /cosmwasm

COPY . .

RUN make install

RUN ARCH=`uname -m` && ln -s /go/pkg/mod/github.com/\!cosm\!wasm/wasmvm@v*/api/libwasmvm.${ARCH}.so /usr/lib/libwasmvm.${ARCH}.so

# ##################################

FROM debian:buster as hub

# Set up dependencies
ENV PACKAGES jq curl

RUN apt-get update && \
    apt-get install -y $PACKAGES

COPY --from=builder /usr/lib/libwasmvm.*.so /usr/lib/
COPY --from=builder /go/bin/fetchd /usr/bin/fetchd
COPY entrypoints/entrypoint.sh /usr/bin/entrypoint.sh

VOLUME /root/.fetchd
VOLUME /root/secret-temp-config

WORKDIR /root

ENTRYPOINT [ "/usr/bin/entrypoint.sh" ]
EXPOSE 1317
EXPOSE 26656
EXPOSE 26657
STOPSIGNAL SIGTERM

# ##################################

FROM hub as gcr

COPY ./entrypoints/run-node.sh /usr/bin/run-node.sh
COPY ./entrypoints/run-tx-server.sh /usr/bin/run-tx-server.sh

# ##################################

FROM hub as localnet

COPY ./entrypoints/run-localnet.sh /usr/bin/run-localnet.sh

ENTRYPOINT [ "/usr/bin/run-localnet.sh" ]

# ##################################

FROM hub as localnet-setup

RUN apt-get update && apt-get install -y python3

COPY ./entrypoints/run-localnet-setup.py /usr/bin/run-localnet-setup.py

ENV PYTHONUNBUFFERED=1

ENTRYPOINT [ "/usr/bin/run-localnet-setup.py" ]
CMD []

