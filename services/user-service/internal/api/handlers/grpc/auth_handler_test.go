package handlers_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/common/errors"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	handlers "github.com/rijum8906/relay/services/user-service/internal/api/handlers/grpc"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
	"github.com/rijum8906/relay/services/user-service/internal/tests/mocks"
	"github.com/rijum8906/relay/services/user-service/internal/tests/testutils"
	"github.com/rijum8906/relay/services/user-service/internal/utils"
)

func TestAuthHandler_Signin(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		services *handlers.Services
		// Named input parameters for target function.
		req     *authv1.SigninRequest
		want    *authv1.SigninResponse
		wantErr bool
	}{
		{
			name: "",
			services: &handlers.Services{
				AccountService: testutils.MockAccountService(&mocks.MockAccountRepo{
					GetAccountByEmailFunc: func(ctx context.Context, email string) (*db.Account, *errors.AppError) {
						return &db.Account{
							ID:    pgtype.UUID{},
							Email: "",
						}, nil
					},
				}),
				AuthService:    testutils.MockAuthService(&mocks.MockAuthRepo{}, nil, nil),
				Profileservice: testutils.MockProfileService(&mocks.MockProfileRepo{}),
			},
			req: &authv1.SigninRequest{
				Email:    utils.NewEmail(""),
				Password: utils.NewPassword(""),
			},
			want:    &authv1.SigninResponse{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handlers.NewAuthHandler(tt.services)
			got, gotErr := h.Signin(context.Background(), tt.req)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("Signin() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("Signin() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if true {
				t.Errorf("Signin() = %v, want %v", got, tt.want)
			}
		})
	}
}
