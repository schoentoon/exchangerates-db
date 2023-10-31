FROM docker.io/library/golang:alpine AS builder

RUN apk add gcc musl-dev

WORKDIR /app

COPY go.mod go.sum /app/

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o /bin/exchangerates-db ./cmd/exchangerates-db/...

FROM alpine:latest

COPY --from=builder /bin/exchangerates-db /bin/exchangerates-db

CMD ["/bin/exchangerates-db"]
