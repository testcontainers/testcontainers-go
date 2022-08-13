ARG GO_VERSION=1.19

FROM docker.io/golang:${GO_VERSION}-alpine as tools

ENV CGO_ENABLED=0

RUN go install github.com/go-delve/delve/cmd/dlv@latest

FROM docker.io/golang:${GO_VERSION}-alpine

ENV CGO_ENABLED=0

WORKDIR /work

COPY --from tools /go/bin/dlv /usr/local/bin/
COPY go.* ./

RUN go mod download -x && \
    apk add -U docker-compose make