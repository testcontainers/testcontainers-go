FROM golang:1.24-alpine@sha256:fc2cff6625f3c1c92e6c85938ac5bd09034ad0d4bc2dfb08278020b68540dbb5 as builder
WORKDIR /app
COPY . .
RUN mkdir -p dist
RUN go build -o ./dist/server main.go

FROM alpine
WORKDIR /app
COPY --from=builder /app/dist/server .
EXPOSE 6443
CMD ["/app/server"]
