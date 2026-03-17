# Stage 1: Build
FROM golang:1.25-bookworm AS builder
WORKDIR /app
COPY ./go.mod /app/go.mod
RUN go mod download
COPY . .
RUN go build -o graphql-gateway ./cmd/main.go

# Stage 2: Final Image
FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/graphql-gateway .
COPY --from=builder /app/.env .
CMD ["./graphql-gateway"]
EXPOSE 8080
