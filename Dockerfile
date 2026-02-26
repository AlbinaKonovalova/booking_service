# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git

# Копируем go.mod и go.sum
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем бинарник
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server cmd/server/main.go

# Final stage
FROM alpine:3.19

WORKDIR /app

# Копируем бинарник
COPY --from=builder /bin/server /app/server

# Копируем конфигурацию
COPY config/config.yaml /app/config/config.yaml

# Создаем непривилегированного пользователя
RUN adduser -D -g '' appuser
USER appuser

EXPOSE 8080

ENTRYPOINT ["/app/server"]
CMD ["-config", "/app/config/config.yaml"]
