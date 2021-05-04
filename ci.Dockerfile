FROM golang:1.15-buster

WORKDIR /src

COPY . .

RUN make go-mod-cache && \
  make build
