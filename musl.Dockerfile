# syntax=docker/dockerfile:1
#
# Static (musl) build of fetchd, producing a self-contained image that does NOT
# need the dynamic libwasmvm.<arch>.so libraries shipped alongside the binary.
#
# Differences to the (glibc) Dockerfile:
#   - Builder runs on Alpine (musl), not Debian (glibc).
#   - fetchd is statically linked against the pre-built static wasmvm archive
#     (libwasmvm_muslc.<arch>.a) selected by the `muslc` build tag.
#   - The runtime image is plain Alpine + the static binary: no .so copying.
#
# The build is fully self-contained w.r.t. Go module downloads: the host must
# provide a Go proxy-layout module cache (see the header of .dockerignore),
# passed as a NAMED BUILD CONTEXT so it is never copied into image layers:
#
#   docker build -f musl.Dockerfile \
#     --build-context modcache=.modcache \
#     --target hub -t fetchai/fetchd:musl .
#
# (the musl-hub target in docker-bake.hcl wires this up automatically). This
# avoids the need for a GitHub token (netrc) at build time even though the
# module graph pulls in private repositories (github.com/fetchai/priv_wasmd_sec,
# github.com/fetchai/priv_wasmvm_sec).
#
# Populate the module cache on the host before building, e.g.:
#
#   rm -rf .modcache && mkdir -p .modcache
#   cp -R "$(go env GOMODCACHE)/cache/download/." .modcache/
#
# (i.e. copy the *proxy layout* part of the Go module cache, the one rooted at
# GOMODCACHE/cache/download - that is what GOPROXY=file:///modcache expects.)
#
# Build (single platform): see the --build-context example above.
#
# Build (multi-platform, via buildx):
#
#   docker buildx build --platform linux/amd64,linux/arm64 \
#     -f musl.Dockerfile --build-context modcache=.modcache \
#     --target hub -t fetchai/fetchd:musl --push .

FROM golang:1.25.7-alpine3.23 AS builder

# build-base: gcc + musl-dev + make (needed for the static cgo/wasmvm link)
# git:      the Makefile derives VERSION/COMMIT via `git describe`/`git log`
#           (the build context includes .git, so do not add a .git entry to
#           .dockerignore for this image)
# xz:       to decompress the static wasmvm archive
# linux-headers: Linux kernel UAPI headers, needed by the ledger support
#           (github.com/zondax/hid compiles libusb sources that include
#           <linux/types.h> etc.); only used when LEDGER_ENABLED=true, but
#           installed unconditionally to keep the builder image uniform
RUN apk add --no-cache build-base git xz linux-headers

# Serve all Go modules from the pre-populated proxy cache, falling back to the
# public proxy/direct only if something is missing. GOSUMDB is disabled since
# the private modules are not in the public checksum database; their integrity
# is still enforced by the entries in go.sum (and the wasmvm archive checksum
# verified explicitly below). GOPRIVATE must NOT be set here: it would make
# `go` bypass GOPROXY (the file:// cache) for the private modules and attempt
# authenticated VCS fetches instead.
ENV GOPROXY=file:///modcache,https://proxy.golang.org,direct
ENV GOSUMDB=off

# Optional Go runtime setting, empty by default. Set via --build-arg only where
# needed, e.g. GODEBUG=asyncpreempt=0 when building linux/amd64 under CPU
# emulation (QEMU/Rosetta on Apple Silicon), where emulated signal delivery
# can crash the go toolchain (SIGSEGV in runtime.suspendG). On native builds
# (Linux CI) leave it unset. The ARG value is exposed as an environment
# variable to the RUN steps below; no ENV needed.
ARG GODEBUG=""

WORKDIR /fetchd

# The module cache is provided as the `modcache` named build context (see the
# header comment) and bind-mounted read-only at /modcache for the steps that
# need it; it is excluded from the main context via .dockerignore so
# `COPY . .` never copies ~5 GB into image layers.
#
# The Go module cache and build cache use BuildKit cache mounts, so the
# downloaded/extracted modules and compiled packages never bloat the builder
# image layers either (and persist across rebuilds on the same builder).
# The build cache is per-platform (the go build cache is content-keyed and
# arch-sensitive), the module cache is arch-independent and shared.
COPY go.mod go.sum ./
RUN --mount=type=bind,from=modcache,source=/,target=/modcache \
    --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Extract the static (musl) wasmvm archive shipped inside the module that
# satisfies github.com/CosmWasm/wasmvm/v3 (following any `replace` directives
# in go.mod) into /lib, verify it, and build. All of that lives in the
# Makefile (see the static-wasmvm-lib / install-static* targets below), so
# the same flow can be run outside Docker (e.g. on a bare Alpine CI runner).

# Optional: enable ledger support in the static build. Disabled by default:
# statically linking the ledger/HID USB stack on musl needs extra care, and
# the usual motivation for a static build is a minimal, reproducible runtime.
ARG LEDGER_ENABLED=true

# Which Makefile static-install target to run:
#   - install-static          : generic static build, no archive checksum pinning
#   - install-static-v0.15.1  : release build for v0.15.1, additionally asserts
#                               the wasmvm version and the SHA256 of the
#                               static wasmvm archive before building
# The extraction of libwasmvm_muslc.<arch>.a from the module and the optional
# checksum verification live in the Makefile, so they can be reused outside
# Docker (e.g. on a bare Alpine CI runner) as well.
ARG INSTALL_TARGET=install-static-v0.15.1

COPY . .

# Static build via the Makefile: the `muslc` build tag selects the static
# wasmvm cgo bindings (internal/api/link_muslc_*.go) instead of the dynamic
# .so ones, and the static external link keeps the Makefile's -buildmode=pie
# (musl fully supports static-PIE, unlike glibc). The Makefile's usual
# version ldflags (name, commit, tags, ...) are preserved.
# NOTE: ARG values are exposed as environment variables to RUN steps, hence
# the shell-style (not Make-style) $INSTALL_TARGET reference.
# The same cache mounts as the `go mod download` step above are used, so
# the modules and compiled packages persist across builds without landing in
# image layers.
RUN --mount=type=bind,from=modcache,source=/,target=/modcache,ro \
    --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    make ${INSTALL_TARGET} LEDGER_ENABLED=${LEDGER_ENABLED}

# Sanity check: the binary must be statically linked (note that busybox ldd
# misleadingly prints the musl loader line even for static binaries, so use
# `file` instead). Static-PIE is expected, since the Makefile forces PIE and
# musl supports static-PIE.
RUN file /go/bin/fetchd | grep -E "static-pie linked|statically linked"

# ##################################

FROM alpine:3.23 AS hub

# Runtime dependencies of entrypoint.sh (bash + curl + jq; sed is in busybox)
RUN apk add --no-cache bash jq curl

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

# Same as hub, plus the extra entrypoint scripts used by internal (GCR)
# deployments - mirrors the gcr stage of the (glibc) Dockerfile.
FROM hub AS gcr

COPY ./entrypoints/run-node.sh /usr/bin/run-node.sh
COPY ./entrypoints/run-tx-server.sh /usr/bin/run-tx-server.sh
