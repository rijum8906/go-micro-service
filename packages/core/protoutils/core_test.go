package protoutils_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/protoutils"
	"github.com/rijum8906/relay/packages/core/token"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
)

func TestValidateIDAndScopedToken(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		req     *corev1.IDAndScopedTokenRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &corev1.IDAndScopedTokenRequest{
				Id:         uuid.NewString(),
				TokenScope: string(token.TokenScopeAdmin),
			},
			wantErr: false,
		},
		{
			name: "invalid id",
			req: &corev1.IDAndScopedTokenRequest{
				Id:         "invalid",
				TokenScope: string(token.TokenScopeAdmin),
			},
			wantErr: true,
		},
		{
			name: "invalid token scope",
			req: &corev1.IDAndScopedTokenRequest{
				Id:         uuid.NewString(),
				TokenScope: "invalid",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := protoutils.ValidateIDAndScopedToken(tt.req)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("ValidateIDAndScopedToken() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("ValidateIDAndScopedToken() succeeded unexpectedly")
			}
		})
	}
}
