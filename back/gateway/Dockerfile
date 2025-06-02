FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY gateway /app/gateway
COPY proto /app/proto
COPY pkg /app/pkg

# Сборка с уникальным именем
RUN CGO_ENABLED=0 GOOS=linux go build -o gateway_bin ./gateway/cmd/main.go

FROM alpine:latest
WORKDIR /app

# Копируем бинарник как gateway
COPY --from=builder /app/gateway_bin ./gateway

EXPOSE 8090
CMD ["./gateway"]
