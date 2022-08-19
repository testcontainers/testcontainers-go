ARG GO_VERSION=1.19
FROM docker.io/golang:${GO_VERSION}-alpine

ENV CGO_ENABLED=0

WORKDIR /work

COPY go.* ./

RUN go mod download -x && \
    apk add -U docker-compose