# syntax=docker/dockerfile:1
FROM golang:1.25.7-bookworm

WORKDIR /src

COPY . .

# The netrc secret (id=netrc) is provided by CI (docker build --secret) and holds
# a GitHub token with read access to the private go module dependencies.
# See .github/workflows/build_and_test.yml
RUN --mount=type=secret,id=netrc,target=/root/.netrc \
    GOPRIVATE=github.com/fetchai/priv_wasmd_sec,github.com/fetchai/priv_wasmvm_sec \
    make go-mod-cache && \
    go mod download all && \
    make build
