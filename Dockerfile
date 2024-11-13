# Этап сборки
FROM golang:1.22-alpine AS builder
ENV CGO_ENABLED=0 TZ=Europe/Moscow

RUN apk add --no-cache git tzdata

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o /app/main cmd/auth/main.go

# Финальный образ
FROM alpine:3.20
ENV TZ=Europe/Moscow

RUN apk add --no-cache tzdata

WORKDIR /app
COPY --from=builder /app/main .
COPY .env .
COPY config config

CMD ["./main"]
