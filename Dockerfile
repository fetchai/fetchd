FROM golang:1.14-buster as builder

RUN apt-get update && \
    apt-get install -y jq curl wget jq file make git

WORKDIR /cosmwasm

COPY . .

RUN make install

# ##################################

FROM debian:buster as fetchd

ARG DEFAULT_RPC_ENDPOINT=""
ARG DEFAULT_SEEDS=""

ENV RPC_ENDPOINT="${DEFAULT_RPC_ENDPOINT}"
ENV SEEDS="${DEFAULT_SEEDS}"

RUN apt-get update && \
    apt-get install -y jq curl

COPY --from=builder /go/pkg/mod/github.com/\!cosm\!wasm/go-cosmwasm@v*/api/libgo_cosmwasm.so /usr/lib/libgo_cosmwasm.so
COPY --from=builder /go/bin/fetchcli /usr/bin/fetchcli
COPY --from=builder /go/bin/fetchd /usr/bin/fetchd
COPY entrypoints/entrypoint.sh /usr/bin/entrypoint.sh

VOLUME /root/.fetchd
VOLUME /root/.fetchcli

EXPOSE 1317
EXPOSE 26656
EXPOSE 26657

ENTRYPOINT [ "/usr/bin/entrypoint.sh" ]
STOPSIGNAL SIGTERM

# ##################################

FROM fetchd as gcr

COPY ./entrypoints/run-node.sh /usr/bin/run-node.sh
COPY ./entrypoints/run-tx-server.sh /usr/bin/run-tx-server.sh

# ##################################

FROM fetchd as localnet

COPY ./entrypoints/run-localnet.sh /usr/bin/run-localnet.sh

ENTRYPOINT [ "/usr/bin/run-localnet.sh" ]

# ##################################

FROM fetchd as localnet-setup

RUN apt-get update && apt-get install -y python3

COPY ./entrypoints/run-localnet-setup.py /usr/bin/run-localnet-setup.py

ENV PYTHONUNBUFFERED=1

ENTRYPOINT [ "/usr/bin/run-localnet-setup.py" ]
CMD []
