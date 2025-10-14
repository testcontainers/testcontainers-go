FROM golang:1.24-alpine@sha256:fc2cff6625f3c1c92e6c85938ac5bd09034ad0d4bc2dfb08278020b68540dbb5

WORKDIR /app

COPY echoserver.go .

CMD go run echoserver.go
