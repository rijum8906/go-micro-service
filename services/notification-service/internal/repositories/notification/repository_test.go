package notification_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/template"
	"github.com/rijum8906/relay/packages/core/testutils"
	"github.com/rijum8906/relay/services/notification-service/internal/db"
	"github.com/rijum8906/relay/services/notification-service/internal/repositories/notification"
)

func Test_service_CreateNotification(t *testing.T) {
	pool := testutils.MustConnectDB()
	querier := db.New(pool)
	repo := notification.New(querier)
	tests := []struct {
		name    string // description of this test case
		params  db.CreateNotificationParams
		wantErr bool
		errCode apperror.ErrorCode
	}{
		{
			name:    "empty params",
			params:  db.CreateNotificationParams{},
			wantErr: true,
			errCode: apperror.CodeValidation,
		},
		{
			name: "valid params",
			params: db.CreateNotificationParams{
				RecepientEmail:  "9jEw0@example.com",
				RecepientUserID: uuid.New(),
				MessageData:     []byte(`{"type":"email","data":{"to":"9jEw0@example.com","subject":"test subject","body":"test body"}}`),
				Status:          "pending",
				TemplateType:    string(template.TemplateTypeEmailVerification),
				RetryCount:      5,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, appErr := repo.CreateNotification(context.Background(), tt.params)
			if tt.wantErr {
				if appErr == nil {
					t.Errorf("CreateNotification() error = %v, wantErr %v", appErr, tt.wantErr)
				}
				if appErr.Code != tt.errCode {
					t.Errorf("CreateNotification() error = %v, wantErr %v", appErr, tt.wantErr)
				}
			} else {
				if appErr != nil {
					t.Errorf("CreateNotification() error = %v, wantErr %v", appErr.Details, tt.wantErr)
				}
			}
		})
	}
}
