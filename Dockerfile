# Build stage
FROM golang:1.25 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o url-shortener ./cmd/url-shortener/main.go

# Run stage
FROM debian:bookworm-slim

WORKDIR /root/

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/url-shortener .
COPY --from=builder /app/.env.example .env

VOLUME /root/storage

CMD ["./url-shortener"]
