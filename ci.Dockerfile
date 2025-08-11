FROM golang:1.24.3-bookworm

WORKDIR /src

COPY . .

RUN make go-mod-cache && \
  make build
