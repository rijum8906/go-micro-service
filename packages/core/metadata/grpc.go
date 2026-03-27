// Package metadata provides gRPC metadata helpers
package metadata

import (
	"context"

	"google.golang.org/grpc/metadata"
)

type ClientInfo struct {
	DeviceID   string
	UserAgent  string
	IPAddress  string
	ClientType string
}

// Send adds client info to outgoing gRPC metadata
func Send(ctx context.Context, info ClientInfo) context.Context {
	md := metadata.Pairs(
		"device-id", info.DeviceID,
		"user-agent", info.UserAgent,
		"client-ip", info.IPAddress,
		"client-type", info.ClientType,
	)

	if existing, ok := metadata.FromOutgoingContext(ctx); ok {
		md = metadata.Join(existing, md)
	}

	return metadata.NewOutgoingContext(ctx, md)
}

// Receive extracts client info from incoming gRPC metadata
func Receive(ctx context.Context) (ClientInfo, bool) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ClientInfo{}, false
	}

	return ClientInfo{
		DeviceID:   get(md, "device-id"),
		UserAgent:  get(md, "user-agent"),
		IPAddress:  get(md, "client-ip"),
		ClientType: get(md, "client-type"),
	}, true
}

func get(md metadata.MD, key string) string {
	if vals := md.Get(key); len(vals) > 0 {
		return vals[0]
	}
	return ""
}
