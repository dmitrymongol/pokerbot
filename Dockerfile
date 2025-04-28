# Этап сборки
FROM golang:1.24-alpine AS builder

# Установка зависимостей для сборки
RUN apk add --no-cache git ca-certificates

# Настройка модуля Go
ENV GO111MODULE=on
ENV GOPATH=/go
ENV PATH=$PATH:/go/bin

# Создание рабочей директории
WORKDIR /app

# Копируем файлы модулей
COPY go.mod go.sum ./

RUN go mod tidy

RUN go mod verify

# Скачиваем зависимости
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарник с правильным модулем
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bot ./cmd/bot/main.go

# Финальный этап
FROM alpine:latest

# Копируем SSL сертификаты
RUN apk --no-cache add ca-certificates

# Копируем бинарник из этапа сборки
COPY --from=builder /bot /bot

# Устанавливаем рабочую директорию
WORKDIR /

# Указываем точку входа
CMD ["/bot"]