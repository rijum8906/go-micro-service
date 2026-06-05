package metadata

import (
	"context"
	"testing"

	"github.com/rijum8906/relay/packages/core/dto"
	grpcmetadata "google.golang.org/grpc/metadata"
)

func TestSendUserInfoOverwritesExistingOutgoingValues(t *testing.T) {
	ctx := grpcmetadata.NewOutgoingContext(context.Background(), grpcmetadata.Pairs(
		dto.MetaUserIDKey, "old-user",
		"keep-me", "yes",
	))

	ctx = SendUserInfo(ctx, dto.UserInfo{
		UserID:    "new-user",
		TokenID:   "token-1",
		SessionID: "session-1",
	})

	md, ok := grpcmetadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("expected outgoing metadata")
	}

	if got := md.Get(dto.MetaUserIDKey); len(got) != 1 || got[0] != "new-user" {
		t.Fatalf("expected user-id to be overwritten, got %v", got)
	}

	if got := md.Get("keep-me"); len(got) != 1 || got[0] != "yes" {
		t.Fatalf("expected unrelated metadata to be preserved, got %v", got)
	}
}

