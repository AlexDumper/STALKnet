FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum* ./
RUN go mod download || true

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o chat .

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/chat .

EXPOSE 8083

CMD ["./chat"]
