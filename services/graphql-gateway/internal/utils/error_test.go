package utils

import (
	"context"
	"errors"
	"testing"

	"github.com/vektah/gqlparser/v2/gqlerror"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestWrapErrorMapsGRPCCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "invalid argument",
			err:  status.Error(codes.InvalidArgument, "invalid payload"),
			want: CodeBadUserInput,
		},
		{
			name: "unauthenticated",
			err:  status.Error(codes.Unauthenticated, "missing token"),
			want: CodeUnauthorized,
		},
		{
			name: "unavailable",
			err:  status.Error(codes.Unavailable, "service down"),
			want: CodeServiceUnavailable,
		},
		{
			name: "non grpc error falls back",
			err:  errors.New("boom"),
			want: CodeInternalServerError,
		},
		{
			name: "wrapped grpc error string",
			err:  errors.New("input: signin rpc error: code = Unauthenticated desc = invalid email or password"),
			want: CodeUnauthorized,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			appErr := WrapError(tt.err)
			if appErr == nil {
				t.Fatal("WrapError() returned nil")
			}

			if appErr.Code != tt.want {
				t.Fatalf("WrapError() code = %q, want %q", appErr.Code, tt.want)
			}
		})
	}
}

func TestPresentErrorHidesInternalMessage(t *testing.T) {
	t.Parallel()

	gqlErr := PresentError(context.Background(), errors.New("sensitive"), false)

	if gqlErr.Message != "Internal server error" {
		t.Fatalf("PresentError() message = %q, want internal fallback", gqlErr.Message)
	}

	if gqlErr.Extensions["code"] != CodeInternalServerError {
		t.Fatalf("PresentError() code = %v, want %q", gqlErr.Extensions["code"], CodeInternalServerError)
	}
}

func TestPresentErrorPreservesKnownClientMessage(t *testing.T) {
	t.Parallel()

	err := status.Error(codes.InvalidArgument, "email is invalid")
	gqlErr := PresentError(context.Background(), err, false)

	if gqlErr.Message != "Email is invalid" {
		t.Fatalf("PresentError() message = %q, want grpc message", gqlErr.Message)
	}

	if gqlErr.Extensions["code"] != CodeBadUserInput {
		t.Fatalf("PresentError() code = %v, want %q", gqlErr.Extensions["code"], CodeBadUserInput)
	}
}

func TestPresentErrorParsesWrappedGRPCMessage(t *testing.T) {
	t.Parallel()

	err := errors.New("input: signin rpc error: code = Unauthenticated desc = invalid email or password")
	gqlErr := PresentError(context.Background(), err, false)

	if gqlErr.Message != "Invalid email or password" {
		t.Fatalf("PresentError() message = %q, want parsed description", gqlErr.Message)
	}

	if gqlErr.Extensions["code"] != CodeUnauthorized {
		t.Fatalf("PresentError() code = %v, want %q", gqlErr.Extensions["code"], CodeUnauthorized)
	}
}

func TestPresentErrorParsesGraphQLErrorMessage(t *testing.T) {
	t.Parallel()

	err := &gqlerror.Error{Message: "rpc error: code = Unauthenticated desc = invalid email or password"}
	gqlErr := PresentError(context.Background(), err, false)

	if gqlErr.Message != "Invalid email or password" {
		t.Fatalf("PresentError() message = %q, want parsed description", gqlErr.Message)
	}

	if gqlErr.Extensions["code"] != CodeUnauthorized {
		t.Fatalf("PresentError() code = %v, want %q", gqlErr.Extensions["code"], CodeUnauthorized)
	}
}

func TestSanitizeClientMessageRemovesInputPrefix(t *testing.T) {
	t.Parallel()

	got := sanitizeClientMessage("input: signin failed")
	if got != "Failed" {
		t.Fatalf("sanitizeClientMessage() = %q, want %q", got, "Failed")
	}
}

func TestSanitizeClientMessageParsesRawRPCMessage(t *testing.T) {
	t.Parallel()

	got := sanitizeClientMessage("rpc error: code = Unauthenticated desc = invalid email or password")
	if got != "Invalid email or password" {
		t.Fatalf("sanitizeClientMessage() = %q, want %q", got, "Invalid email or password")
	}
}
