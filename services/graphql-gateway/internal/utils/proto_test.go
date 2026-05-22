package utils

import (
	"testing"

	"github.com/rijum8906/relay/packages/core/token"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
)

func TestParseAuthMethod_GraphQLShortNames(t *testing.T) {
	m, err := ParseAuthMethod(token.AuthMethodPassword)
	if err != nil {
		t.Fatalf("ParseAuthMethod(PASSWORD): %v", err)
	}
	if m != corev1.AuthMethod_AUTH_METHOD_PASSWORD {
		t.Fatalf("got %v want AUTH_METHOD_PASSWORD", m)
	}
}

func TestParseAuthMethod_FullProtoName(t *testing.T) {
	m, err := ParseAuthMethod(token.AuthMethod("AUTH_METHOD_PASSWORD"))
	if err != nil {
		t.Fatalf("ParseAuthMethod(AUTH_METHOD_PASSWORD): %v", err)
	}
	if m != corev1.AuthMethod_AUTH_METHOD_PASSWORD {
		t.Fatalf("got %v want AUTH_METHOD_PASSWORD", m)
	}
}

func TestParseScope_GraphQLShortNames(t *testing.T) {
	s, err := ParseScope(token.TokenScopeChangePassword)
	if err != nil {
		t.Fatalf("ParseScope(CHANGE_PASSWORD): %v", err)
	}
	if s != corev1.TokenScope_TOKEN_SCOPE_CHANGE_PASSWORD {
		t.Fatalf("got %v want TOKEN_SCOPE_CHANGE_PASSWORD", s)
	}
}

func TestParseScope_FullProtoName(t *testing.T) {
	s, err := ParseScope(token.TokenScope("TOKEN_SCOPE_CHANGE_PASSWORD"))
	if err != nil {
		t.Fatalf("ParseScope(TOKEN_SCOPE_CHANGE_PASSWORD): %v", err)
	}
	if s != corev1.TokenScope_TOKEN_SCOPE_CHANGE_PASSWORD {
		t.Fatalf("got %v want TOKEN_SCOPE_CHANGE_PASSWORD", s)
	}
}
