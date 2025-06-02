FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY auth_service /app/auth_service
COPY proto /app/proto
COPY pkg /app/pkg

# Сборка с другим именем
RUN CGO_ENABLED=0 GOOS=linux go build -o auth_service_bin ./auth_service/cmd/main/main.go

FROM alpine:latest
WORKDIR /app

# Копируем бинарник
COPY --from=builder /app/auth_service_bin ./auth_service

EXPOSE 50053
CMD ["./auth_service"]
