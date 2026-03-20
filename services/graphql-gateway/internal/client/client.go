package client

import (
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
)

func NewUserClient() authv1.AuthServiceClient {
	addr := os.Getenv("USER_SERVICE_GRPC_ADDR")
	if addr == "" {
		addr = "user-service:8906"
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	return authv1.NewAuthServiceClient(conn)
}
