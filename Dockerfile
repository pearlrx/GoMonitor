# -----------------------------
# 1. Stage: Builder
# -----------------------------
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o collector ./cmd/collector

# -----------------------------
# 2. Stage: Create app
# -----------------------------
FROM debian:bookworm-slim

WORKDIR /app

COPY --from=builder /app/collector .

COPY config.yaml .
COPY migrations ./migrations

RUN apt-get update && apt-get install -y tzdata && rm -rf /var/lib/apt/lists/*

CMD ["./collector"]