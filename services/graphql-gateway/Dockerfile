FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy only go.mod and go.sum files first
COPY go.mod go.sum ./
COPY packages/core/go.mod packages/core/go.sum ./packages/core/
COPY packages/pb/go.mod packages/pb/go.sum ./packages/pb/
COPY services/graphql-gateway/go.mod services/graphql-gateway/go.sum ./services/graphql-gateway/

RUN go work init
RUN go work use ./services/*
RUN go work use ./packages/*

# Download dependencies - this layer will be cached unless go.mod changes
RUN go mod download

# Now copy source code - but only the necessary directories
COPY packages/core/ ./packages/core/
COPY packages/pb/ ./packages/pb/
COPY services/graphql-gateway/ ./services/graphql-gateway/

# Build
RUN go build -o graphql-gateway services/graphql-gateway/cmd/main.go

FROM alpine:3.10

WORKDIR /app

COPY --from=builder /app/graphql-gateway .
COPY services/graphql-gateway/.env .

EXPOSE 8081

CMD ["./graphql-gateway"]
