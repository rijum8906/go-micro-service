package account_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/common/env"
	"github.com/rijum8906/relay/packages/common/errors"
	accountv1 "github.com/rijum8906/relay/packages/pb/user_service/account/v1"
	"github.com/rijum8906/relay/services/user-service/internal/api/dto/request"
	db "github.com/rijum8906/relay/services/user-service/internal/db/generated"
	"github.com/rijum8906/relay/services/user-service/internal/services/account"
	"github.com/rijum8906/relay/services/user-service/internal/tests/mocks"
	"github.com/rijum8906/relay/services/user-service/internal/utils"
)

func TestAccountService_MyAccount_Success(t *testing.T) {
	mockRepo := &mocks.MockAccountRepo{
		GetAccountFunc: func(ctx context.Context, id pgtype.UUID) (*db.Account, *errors.AppError) {
			return &db.Account{
				Email: "test@gmail.com",
			}, nil
		},
	}

	service := account.NewAccountService(
		mockRepo,
		nil,
		&utils.UtilsConfig{},
		&env.Env{},
	)

	meta := &request.AuthzMetadata{
		UserID: pgtype.UUID{}, // fake
	}

	acc, err := service.MyAccount(context.Background(), meta)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if acc.Email != "test@gmail.com" {
		t.Errorf("got %v, want %v", acc.Email, "test@gmail.com")
	}
}

func Test_accountService_UpdateEmail_Failure(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		req           *accountv1.UpdateEmailRequest
		authzMetadata *request.AuthzMetadata
		email         string
		wantNil       bool
	}{
		{
			name: "UpdateEmail success",
			req: &accountv1.UpdateEmailRequest{
				NewEmail: utils.NewEmail("rijum8906@gamil.com"),
			},
			authzMetadata: &request.AuthzMetadata{
				UserID: pgtype.UUID{},
			},
			email:   "newemail@gmail.com",
			wantNil: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: construct the receiver type.
			s := account.NewAccountService(&mocks.MockAccountRepo{
				UpdateEmailFunc: func(ctx context.Context, newEmail string, authzMetadata *request.AuthzMetadata) *errors.AppError {
					return nil
				},
			}, nil, &utils.UtilsConfig{}, &env.Env{})
			got := s.UpdateEmail(context.Background(), tt.req, tt.authzMetadata, tt.email)
			// TODO: update the condition below to compare got with tt.want.
			if got != nil {
				t.Errorf("UpdateEmail() = %v, want %v", got, tt.wantNil)
			}
		})
	}
}
