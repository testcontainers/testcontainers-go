FROM golang:1.25-alpine@sha256:8e02eb337d9e0ea459e041f1ee5eece41cbb61f1d83e7d883a3e2fb4862063fa as builder
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
