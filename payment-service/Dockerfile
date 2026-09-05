FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o payment-service ./cmd/main.go

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/payment-service .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8083
CMD ["./payment-service"]
