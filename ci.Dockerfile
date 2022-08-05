FROM golang:1.18-buster

WORKDIR /src

COPY . .

RUN make go-mod-cache && \
  make build
