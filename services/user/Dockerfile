FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY packages/core/go.mod packages/core/go.sum ./
COPY services/user/go.mod services/user/go.sum ./
COPY packages/pb/go.mod packages/pb/go.sum ./

CMD [ "go", "mod", "download" ]

COPY . .

RUN go build -o service-user services/user/cmd/main.go

FROM alpine:3.10

WORKDIR /app

COPY --from=builder /app/service-user .

COPY services/user/.env .

EXPOSE 8906

CMD ["./service-user"]
