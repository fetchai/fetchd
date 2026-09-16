FROM golang:1.25.7-bookworm AS builder

# Set up dependencies
ENV PACKAGES="jq curl wget jq file make git"

RUN apt-get update && \
    apt-get install -y $PACKAGES

WORKDIR /cosmwasm

COPY . .

# Optional Go runtime setting, empty by default. Set via --build-arg only where
# needed, e.g. GODEBUG=asyncpreempt=0 when building linux/amd64 under CPU
# emulation (QEMU/Rosetta on Apple Silicon), where emulated signal delivery
# can crash the go toolchain (SIGSEGV in runtime.suspendG). On native builds
# (Linux CI) leave it unset. The ARG value is exposed as an environment
# variable to the RUN steps below; no ENV needed.
ARG GODEBUG=""

# GitHub token for private repos (GOPRIVATE) is mounted from the host at build
# time; it is never baked into the image layers. Passed via buildx/bake as
# secret id=netrc (a .netrc file readable by git over HTTPS).
RUN --mount=type=secret,id=netrc,target=/root/.netrc make install

RUN GOPATH="$(go env GOPATH)"
# Resolve the wasmvm module's actual cache directory via go list -m, so this
# works regardless of any `replace` directives in go.mod (e.g. the private
# github.com/fetchai/priv_wasmvm_sec replacement) or if they are later dropped.
RUN ARCH=$(uname -m) && \
    WASMVM_DIR=$(go list -m -f '{{.Dir}}' github.com/CosmWasm/wasmvm/v3) && \
    ln -s "${WASMVM_DIR}/internal/api/libwasmvm.${ARCH}.so" /usr/lib/libwasmvm.${ARCH}.so && \
    test -e /usr/lib/libwasmvm.${ARCH}.so || \
    { echo "ERROR: /usr/lib/libwasmvm.${ARCH}.so is a dangling symlink;" >&2; \
      echo "       go list could not resolve github.com/CosmWasm/wasmvm/v3 —" >&2; \
      echo "       check go.mod and GOPRIVATE settings." >&2; exit 1; }

# ##################################

FROM debian:bookworm-slim AS hub

# Set up dependencies
ENV PACKAGES="jq curl"

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

FROM hub AS gcr

COPY ./entrypoints/run-node.sh /usr/bin/run-node.sh
COPY ./entrypoints/run-tx-server.sh /usr/bin/run-tx-server.sh

# ##################################

FROM hub AS localnet

COPY ./entrypoints/run-localnet.sh /usr/bin/run-localnet.sh

ENTRYPOINT [ "/usr/bin/run-localnet.sh" ]

# ##################################

FROM hub AS localnet-setup

RUN apt-get update && apt-get install -y python3

COPY ./entrypoints/run-localnet-setup.py /usr/bin/run-localnet-setup.py

ENV PYTHONUNBUFFERED=1

ENTRYPOINT [ "/usr/bin/run-localnet-setup.py" ]
CMD []

