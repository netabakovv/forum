FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY forum_service /app/forum_service
COPY proto /app/proto
COPY pkg /app/pkg

# Сборка бинарника с уникальным именем
RUN CGO_ENABLED=0 GOOS=linux go build -o forum_service_bin ./forum_service/cmd/main/main.go

FROM alpine:latest
WORKDIR /app

# Копируем бинарник
COPY --from=builder /app/forum_service_bin ./forum_service

EXPOSE 50051 8080
CMD ["./forum_service"]
