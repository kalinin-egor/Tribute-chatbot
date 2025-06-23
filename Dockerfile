# Многоэтапная сборка для оптимизации размера образа
FROM golang:1.21-bullseye AS builder

# Устанавливаем необходимые пакеты для сборки
RUN apt-get update && apt-get install -y --no-install-recommends git ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение с оптимизациями
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -a -installsuffix cgo \
    -o tribute-chatbot .

# Финальный образ
FROM debian:bullseye-slim

# Устанавливаем необходимые пакеты и создаем пользователя
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata && \
    rm -rf /var/lib/apt/lists/* && \
    addgroup --system --gid 1001 appgroup && \
    adduser --system --no-create-home --uid 1001 --ingroup appgroup appuser

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем бинарный файл из builder
COPY --from=builder /app/tribute-chatbot .

# Копируем конфигурационный файл (опционально)
COPY --from=builder /app/config.env ./config.env.example

# Меняем владельца файлов
RUN chown -R appuser:appgroup /app

# Переключаемся на непривилегированного пользователя
USER appuser

# Метаданные образа
LABEL maintainer="Tribute Chatbot Team"
LABEL description="Telegram bot built with Go"
LABEL version="1.0.0"

# Проверка здоровья (healthcheck)
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ps aux | grep tribute-chatbot || exit 1

# Открываем порт (если понадобится в будущем)
EXPOSE 8080

# Переменные окружения по умолчанию
ENV LOG_LEVEL=info
ENV PORT=8080

# Запускаем приложение
ENTRYPOINT ["./tribute-chatbot"] 