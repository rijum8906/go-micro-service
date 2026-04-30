FROM golang:1.26-alpine AS builder

WORKDIR /app

# Copy only go.mod and go.sum files first
COPY go.mod go.sum ./
COPY packages/core/go.mod packages/core/go.sum ./packages/core/
COPY packages/pb/go.mod packages/pb/go.sum ./packages/pb/
COPY services/task-service/go.mod services/task-service/go.sum ./services/task-service/

RUN go work init
RUN go work use ./services/*
RUN go work use ./packages/*

# Download dependencies - this layer will be cached unless go.mod changes
RUN go mod download

# Now copy source code - but only the necessary directories
COPY packages/core/ ./packages/core/
COPY packages/pb/ ./packages/pb/

COPY services/task-service/go.mod ./services/task-service/go.mod
COPY services/task-service/internal ./services/task-service/internal
COPY services/task-service/app ./services/task-service/app
COPY services/task-service/cmd ./services/task-service/cmd

# Build
RUN go build -o service-task services/task-service/cmd/app/main.go

FROM alpine:3.10

WORKDIR /app

COPY --from=builder /app/service-task .
COPY services/task-service/.env .

EXPOSE 8083

CMD ["./service-task"]
