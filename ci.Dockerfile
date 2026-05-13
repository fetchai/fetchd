FROM golang:1.25.7-bookworm

WORKDIR /src

COPY . .

RUN make go-mod-cache && \
    go mod download all && \
    make build
