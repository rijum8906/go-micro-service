FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY packages/core/go.mod packages/core/go.sum ./
COPY services/graphql-gateway/go.mod services/graphql-gateway/go.sum ./
COPY packages/pb/go.mod packages/pb/go.sum ./

CMD [ "go", "mod", "download" ]

COPY . .

RUN go build -o graphql-gateway services/graphql-gateway/cmd/main.go

FROM alpine:3.10

WORKDIR /app

COPY --from=builder /app/graphql-gateway .

COPY services/graphql-gateway/.env .

EXPOSE 8080

CMD ["./graphql-gateway"]
