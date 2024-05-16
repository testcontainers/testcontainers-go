FROM golang:bullseye

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod tidy

COPY . .

RUN go build -o ./pg_test

CMD /app/pg_test