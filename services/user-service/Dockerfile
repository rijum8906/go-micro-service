# Stage 1: Build
FROM golang:1.25-bookworm AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o user-service ./cmd/main.go

# Stage 2: Final Image
FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/user-service .
COPY --from=builder /app/.env .
CMD ["./user-service"]
EXPOSE 8906
