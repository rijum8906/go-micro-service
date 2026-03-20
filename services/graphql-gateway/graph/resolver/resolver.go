package resolver

import (
	"log"
	"os"

	"github.com/rijum8906/relay/packages/common/env"
	accountv1 "github.com/rijum8906/relay/packages/pb/user_service/account/v1"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	Env           *env.Env
	AuthClient    authv1.AuthServiceClient
	AccountClient accountv1.AccountServiceClient
	conn          *grpc.ClientConn
}

func NewResolver(env *env.Env) *Resolver {
	addr := os.Getenv("USER_SERVICE_GRPC_ADDR")
	if addr == "" {
		addr = "user-service:8906"
	}

	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*10)),
	)
	if err != nil {
		log.Fatalf("Failed to connect to user service: %v", err)
	}

	return &Resolver{
		Env:           env,
		AuthClient:    authv1.NewAuthServiceClient(conn),
		AccountClient: accountv1.NewAccountServiceClient(conn),
		conn:          conn,
	}
}

func (r *Resolver) Close() error {
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
