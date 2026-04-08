FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY packages/core/go.mod packages/core/go.sum ./
COPY services/user-service/go.mod services/user-service/go.sum ./
COPY packages/pb/go.mod packages/pb/go.sum ./

CMD [ "go", "mod", "download" ]

COPY . .

RUN go build -o service-user services/user-service/cmd/main.go

FROM alpine:3.10

WORKDIR /app

COPY --from=builder /app/service-user .

COPY services/user-service/.env .

EXPOSE 8906

CMD ["./service-user"]
