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
FROM debian:buster

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
