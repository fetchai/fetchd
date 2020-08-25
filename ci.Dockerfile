FROM golang:buster

WORKDIR /src

COPY . .

RUN make go-mod-cache && \
  make build
