FROM golang:1.26-alpine@sha256:8ac98ca534ac3f51e1f420a1dd2c15e74c75cfa0f23f3ad27eb5d7236c349a0c

ARG FOO

ENV FOO=$FOO

WORKDIR /app

COPY echoserver.go .

CMD go run echoserver.go
