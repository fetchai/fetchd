FROM golang:1.14-buster

WORKDIR /src

COPY . .

RUN make go-mod-cache && \
  make build
