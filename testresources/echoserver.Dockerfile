FROM docker.io/golang:1.13-alpine

WORKDIR /app

COPY echoserver.go .

CMD go run echoserver.go
