FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum* ./
RUN go mod download || true

COPY handlers/ ./handlers/
COPY config/ ./config/
COPY middleware/ ./middleware/
COPY web/ ./web/
COPY main.go ./main.go

# Сборка с оптимизациями для продакшена
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o gateway .

# Production образ с минимальными правами
FROM alpine:3.18

# Создаём непривилегированного пользователя
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -D appuser

WORKDIR /app

COPY --from=builder /app/gateway .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Переключаемся на непривилегированного пользователя
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:8080/health || exit 1

CMD ["./gateway"]
