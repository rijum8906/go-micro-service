FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY packages/core/go.mod packages/core/go.sum ./
COPY services/notification-service/go.mod services/notification-service/go.sum ./
COPY packages/pb/go.mod packages/pb/go.sum ./

CMD [ "go", "mod", "download" ]

COPY . .

RUN go build -o service-notification services/notification-service/cmd/app/main.go

FROM alpine:3.10

WORKDIR /app

COPY --from=builder /app/service-notification .

COPY services/notification-service/.env .

EXPOSE 8082

CMD ["./service-notification"]
