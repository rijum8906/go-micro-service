package client

import (
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	user_servicev1 "github.com/rijum8906/relay/packages/pb/user_service/v1"
)

func NewUserClient() user_servicev1.AuthServiceClient {
	addr := os.Getenv("USER_SERVICE_GRPC_ADDR")
	if addr == "" {
		addr = "user-service:8906"
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}

	return user_servicev1.NewAuthServiceClient(conn)
}
