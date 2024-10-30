# Используем официальный образ Go 1.23 как базовый образ
FROM golang:1.23-alpine AS builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Копируем файлы go.mod и go.sum для установки зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download

# Копируем исходный код приложения
COPY . .

# Собираем приложение
RUN go build -o sso ./cmd/sso/main.go
RUN go build -o migrator ./cmd/migrator/main.go

# Используем минимальный образ для запуска приложения
FROM alpine:latest

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем собранные приложения из предыдущего этапа
COPY --from=builder /app/sso .
COPY --from=builder /app/migrator .

# Копируем конфигурационные файлы и миграции
COPY config/local.yaml ./config/local.yaml
COPY migrations ./migrations
COPY storage ./storage

# Открываем порт, на котором будет работать gRPC-сервер
EXPOSE 50051

# Запускаем миграции и приложение
CMD ["sh", "-c", "./migrator --storage-path=./storage/shilkinskaya-sso.db --migrations-path=./migrations && ./sso --config=config/local.yaml"]