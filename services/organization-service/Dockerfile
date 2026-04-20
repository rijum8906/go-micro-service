FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy only go.mod and go.sum files first
COPY go.mod go.sum ./
COPY packages/core/go.mod packages/core/go.sum ./packages/core/
COPY packages/pb/go.mod packages/pb/go.sum ./packages/pb/
COPY services/organization-service/go.mod services/organization-service/go.sum ./services/organization-service/

RUN go work init
RUN go work use ./services/*
RUN go work use ./packages/*

# Download dependencies - this layer will be cached unless go.mod changes
RUN go mod download

# Now copy source code - but only the necessary directories
COPY packages/core/ ./packages/core/
COPY packages/pb/ ./packages/pb/
COPY services/organization-service/ ./services/organization-service/

# Build
RUN go build -o service-organization services/organization-service/cmd/app/main.go

FROM alpine:3.10

WORKDIR /app

COPY --from=builder /app/service-organization .
COPY services/organization-service/.env .

EXPOSE 8081

CMD ["./service-organization"]
