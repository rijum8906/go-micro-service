# Stage 1: Build
FROM golang:1.26-bookworm AS builder

WORKDIR /app

# 1. Copy workspace configuration and all module manifests
COPY go.work go.work.sum ./
COPY go.mod ./
COPY packages/pb/go.mod packages/pb/go.mod
COPY packages/common/go.mod packages/common/go.mod
COPY services/user-service/go.mod services/user-service/go.mod
COPY services/graphql-gateway/go.mod services/graphql-gateway/go.mod
COPY services/graphql-gateway/.env services/graphql-gateway/.env

# 2. Download dependencies for the whole workspace
RUN go mod download

# 3. Copy the actual source code for required modules
COPY packages/pb/ ./packages/pb/
COPY packages/common/ ./packages/common/
COPY services/graphql-gateway/ ./services/graphql-gateway/

# 4. Build the binary from the root context
RUN go build -o /graphql-gateway ./services/graphql-gateway/cmd/main.go

# Stage 2: Final Image
FROM debian:bookworm-slim
WORKDIR /app

# Copy binary from builder
COPY --from=builder /graphql-gateway .
COPY --from=builder /app/services/graphql-gateway/.env . 

EXPOSE 8080

CMD ["./graphql-gateway"]
