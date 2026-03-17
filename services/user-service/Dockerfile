# Stage 1: Build
FROM golang:1.26-bookworm AS builder
WORKDIR /app
COPY ./go.mod /app/go.mod
RUN go mod download
COPY . .
RUN go build -o user-service ./cmd/main.go

# Stage 2: Final Image
FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/user-service .
COPY --from=builder /app/.env .
CMD ["./user-service"]
EXPOSE 8906
