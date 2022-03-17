FROM golang:1.17-buster

WORKDIR /src

COPY . .

RUN make go-mod-cache && \
  make build
