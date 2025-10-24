# -----------------------------
# 1. Stage: Builder
# -----------------------------
FROM golang:1.24 AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем go.mod и go.sum и качаем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь исходный код
COPY . .

# Собираем бинарник
RUN go build -o collector ./cmd/collector

# -----------------------------
# 2. Stage: Минимальный образ
# -----------------------------
FROM debian:bookworm-slim

# Рабочая директория
WORKDIR /app

# Копируем бинарник из builder stage
COPY --from=builder /app/collector .

# Копируем конфиг и миграции в контейнер
COPY config.yaml .
COPY migrations ./migrations

# Устанавливаем tzdata для корректного времени
RUN apt-get update && apt-get install -y tzdata && rm -rf /var/lib/apt/lists/*

# Команда запуска контейнера
CMD ["./collector"]