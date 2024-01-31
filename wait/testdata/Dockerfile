FROM golang:1.18-alpine@sha256:77f25981bd57e60a510165f3be89c901aec90453fd0f1c5a45691f6cb1528807 as builder
WORKDIR /app
COPY . .
RUN mkdir -p dist
RUN go build -o ./dist/server main.go

FROM alpine
WORKDIR /app
COPY --from=builder /app/tls.pem /app/tls-key.pem ./
COPY --from=builder /app/dist/server .
EXPOSE 6443
CMD ["/app/server"]
