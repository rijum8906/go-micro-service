package token_test

import (
	"testing"

	"github.com/rijum8906/relay/packages/core/token"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
)

func TestTokenScopeFromProto_ChangePassword(t *testing.T) {
	scope, err := token.TokenScopeFromProto(corev1.TokenScope_TOKEN_SCOPE_CHANGE_PASSWORD)
	if err != nil {
		t.Fatalf("TokenScopeFromProto: %v", err)
	}
	if scope != token.TokenScopeChangePassword {
		t.Fatalf("got %q want %q", scope, token.TokenScopeChangePassword)
	}
}

func TestTokenScopeFromProto_Unspecified(t *testing.T) {
	_, err := token.TokenScopeFromProto(corev1.TokenScope_TOKEN_SCOPE_UNSPECIFIED)
	if err == nil {
		t.Fatal("expected error for unspecified scope")
	}
}
