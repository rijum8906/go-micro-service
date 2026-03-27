// Package server
package server

import (
	"log"

	authv1 "github.com/rijum8906/relay/packages/pb/user/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcUserClient struct {
	client authv1.AuthServiceClient
}

func NewGrpcUserClient(serviceAddr string) *GrpcUserClient {
	conn, err := grpc.NewClient(serviceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to user service: %v", err)
	}
	defer conn.Close()
	// Initialize gRPC clients
	authClient := authv1.NewAuthServiceClient(conn)
	return &GrpcUserClient{
		client: authClient,
	}
}
