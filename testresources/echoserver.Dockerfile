FROM golang:1.13-alpine

WORKDIR /app

COPY echoserver.go .

RUN apk add git

RUN go get -u github.com/gin-gonic/gin

ENV GIN_MODE=release

CMD go run echoserver.go
