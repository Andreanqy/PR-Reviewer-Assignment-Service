# Builder
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Копируем модули и скачиваем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь проект
COPY . .

# Сборка бинарника
RUN go build -o server .

# Финальный образ
FROM alpine:3.18

WORKDIR /app

# Копируем бинарник
COPY --from=builder /app/server .

# Открываем порт
EXPOSE 8080

# Команда запуска
CMD ["./server"]
