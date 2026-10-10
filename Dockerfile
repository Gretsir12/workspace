# Stage 1: Сборка (Builder)
FROM golang:alpine AS builder

WORKDIR /app

# Кэшируем зависимости
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Собираем бинарник с отключением CGO и отладочной информации
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/backend ./cmd/

# Stage 2: Финальный минимальный образ
FROM alpine:latest

# Добавляем корневые сертификаты (нужны для HTTPS-запросов) и часовые пояса
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Копируем только бинарный файл из этапа сборки
COPY --from=builder /bin/backend ./backend

EXPOSE 8000

CMD ["./backend"]
