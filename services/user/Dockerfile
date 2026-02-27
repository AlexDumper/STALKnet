FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum* ./
RUN go mod download || true

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o user .

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/user .

EXPOSE 8082

CMD ["./user"]
