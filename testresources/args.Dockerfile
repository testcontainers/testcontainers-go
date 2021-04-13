FROM golang:1.13-alpine

ARG FOO

ENV FOO=$FOO

WORKDIR /app

COPY argsserver.go .

CMD go run argsserver.go
