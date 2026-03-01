FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum* ./
RUN go mod download || true

COPY . .

# Сборка с оптимизациями для продакшена
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o notification .

# Production образ с минимальными правами
FROM alpine:3.18

# Создаём непривилегированного пользователя
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -D appuser

WORKDIR /app

COPY --from=builder /app/notification .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Переключаемся на непривилегированного пользователя
USER appuser

EXPOSE 8085

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:8085/health || exit 1

CMD ["./notification"]
