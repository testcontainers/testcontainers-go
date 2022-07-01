FROM docker.io/golang:1.13-alpine

ARG FOO

ENV FOO=$FOO

WORKDIR /app

COPY echoserver.go .

CMD go run echoserver.go
