# syntax=docker/dockerfile:1

##
## Build
##
FROM golang:1.18.7-alpine AS build

ENV PATH="${GOPATH}/bin:${PATH}"
ENV LLVL=info

COPY . /d-voting
WORKDIR /d-voting/dela/cli/crypto
RUN go install

WORKDIR /d-voting/cli/memcoin/
RUN go install

EXPOSE 2001
WORKDIR /d-voting
# CMD memcoin --config /tmp/node1 start --port 2001 





