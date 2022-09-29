FROM golang:1.15-alpine as builder
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
