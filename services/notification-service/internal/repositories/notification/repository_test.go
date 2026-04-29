package notification_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/template"
	"github.com/rijum8906/relay/packages/core/testutils"
	"github.com/rijum8906/relay/services/notification-service/internal/db"
	"github.com/rijum8906/relay/services/notification-service/internal/repositories/notification"
)

func Test_service_CreateNotification(t *testing.T) {
	pool := testutils.MustConnectDB(testutils.WithDBName(testutils.GetTestDBName("notification-service")))
	querier := db.New(pool)
	repo := notification.New(querier)

	// Add cleanup
	t.Cleanup(func() {
		pool.Exec(context.Background(), "TRUNCATE TABLE notifications CASCADE")
	})

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
			name: "missing email",
			params: db.CreateNotificationParams{
				RecepientEmail:  "",
				RecepientUserID: uuid.New(),
				MessageData:     []byte(`{"type":"email","data":{"to":"9jEw0@example.com","subject":"test subject","body":"test body"}}`),
				Status:          "pending",
				TemplateType:    string(template.TemplateTypeEmailVerification),
				RetryCount:      5,
			},
			wantErr: true,
			errCode: apperror.CodeValidation,
		},
		{
			name:    "valid params",
			params:  createNotificationParams(),
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

func Test_service_GetNotification(t *testing.T) {
	pool := testutils.MustConnectDB(testutils.WithDBName(testutils.GetTestDBName("notification-service")))
	querier := db.New(pool)
	repo := notification.New(querier)

	// Add cleanup
	t.Cleanup(func() {
		pool.Exec(context.Background(), "TRUNCATE TABLE notifications CASCADE")
	})

	// create notification
	params1 := createNotificationParams()
	notif1, appErr := repo.CreateNotification(context.Background(), params1)
	if appErr != nil {
		t.Fatal(appErr)
	}

	// Get One notfication
	fetchedNotif1, appErr := repo.GetNotification(context.Background(), notif1.ID.String())
	if appErr != nil {
		t.Fatal(appErr)
	}
	if fetchedNotif1.ID != notif1.ID {
		t.Errorf("GetNotification() got = %v, want %v", fetchedNotif1.ID, notif1.ID)
	}
}

func Test_service_GetNotificationsByUserID(t *testing.T) {
	pool := testutils.MustConnectDB(testutils.WithDBName(testutils.GetTestDBName("notification-service")))
	querier := db.New(pool)
	repo := notification.New(querier)

	// Add cleanup
	t.Cleanup(func() {
		pool.Exec(context.Background(), "TRUNCATE TABLE notifications CASCADE")
	})

	userID := uuid.New()

	// create 3 notifications
	params1 := createNotificationParams()
	params1.RecepientUserID = userID
	notif1, appErr := repo.CreateNotification(context.Background(), params1)
	if appErr != nil {
		t.Fatal(appErr)
	}
	time.Sleep(500 * time.Millisecond)
	params2 := createNotificationParams()
	params2.RecepientUserID = userID
	notif2, appErr := repo.CreateNotification(context.Background(), params2)
	if appErr != nil {
		t.Fatal(appErr)
	}
	time.Sleep(500 * time.Millisecond)
	params3 := createNotificationParams()
	params3.RecepientUserID = userID
	notif3, appErr := repo.CreateNotification(context.Background(), params3)
	if appErr != nil {
		t.Fatal(appErr)
	}
	time.Sleep(500 * time.Millisecond)

	// Retrive the notifications
	notifs, appErr := repo.GetNotificationsByUserID(context.Background(), userID.String(), 10, 1)
	if appErr != nil {
		t.Fatal(appErr)
	}

	if len(*notifs) != 3 {
		t.Errorf("GetNotificationsByUserID() got = %v, want %v", len(*notifs), 3)
	}

	if (*notifs)[2].ID != notif1.ID {
		t.Errorf("GetNotificationsByUserID() got = %v, want %v", (*notifs)[0].ID, notif1.ID)
	}
	if (*notifs)[1].ID != notif2.ID {
		t.Errorf("GetNotificationsByUserID() got = %v, want %v", (*notifs)[1].ID, notif2.ID)
	}
	if (*notifs)[0].ID != notif3.ID {
		t.Errorf("GetNotificationsByUserID() got = %v, want %v", (*notifs)[2].ID, notif3.ID)
	}
}

func Test_service_UpdateNotificationStatus(t *testing.T) {
	pool := testutils.MustConnectDB(testutils.WithDBName(testutils.GetTestDBName("notification-service")))
	querier := db.New(pool)
	repo := notification.New(querier)

	// Add cleanup
	t.Cleanup(func() {
		pool.Exec(context.Background(), "TRUNCATE TABLE notifications CASCADE")
	})

	// create
	notif, appErr := repo.CreateNotification(context.Background(), createNotificationParams())
	if appErr != nil {
		t.Fatal(appErr)
	}

	// update
	appErr = repo.UpdateNotificationStatus(context.Background(), db.UpdateNotificationStatusParams{
		ID:     notif.ID,
		Status: "delivered",
	})
	if appErr != nil {
		t.Fatal(appErr)
	}

	// get updated
	updatedNotif, appErr := repo.GetNotification(context.Background(), notif.ID.String())
	if appErr != nil {
		t.Fatal(appErr)
	}

	if updatedNotif.Status != "delivered" {
		t.Errorf("UpdateNotificationStatus() got = %v, want %v", updatedNotif.Status, "delivered")
	}
}

func Test_service_DeleteNotification(t *testing.T) {
	pool := testutils.MustConnectDB(testutils.WithDBName(testutils.GetTestDBName("notification-service")))
	querier := db.New(pool)
	repo := notification.New(querier)

	// Add cleanup
	t.Cleanup(func() {
		pool.Exec(context.Background(), "TRUNCATE TABLE notifications CASCADE")
	})

	// create
	notif, appErr := repo.CreateNotification(context.Background(), createNotificationParams())
	if appErr != nil {
		t.Fatal(appErr)
	}

	// delete
	appErr = repo.DeleteNotification(context.Background(), notif.ID.String())
	if appErr != nil {
		t.Fatal(appErr)
	}

	// get deleted
	_, appErr = repo.GetNotification(context.Background(), notif.ID.String())
	if appErr == nil {
		t.Errorf("DeleteNotification() got = %v, want %v", appErr, "error")
	}

	notifs, appErr := repo.GetNotificationsByUserID(context.Background(), notif.RecepientUserID.String(), 1, 1)
	if appErr != nil {
		t.Fatal(appErr)
	}

	if len(*notifs) != 0 {
		t.Errorf("DeleteNotification() got = %v, want %v", len(*notifs), 0)
	}
}

func createNotificationParams() db.CreateNotificationParams {
	return db.CreateNotificationParams{
		RecepientEmail:  testutils.GenerateRandomEmail(),
		RecepientUserID: uuid.New(),
		MessageData:     []byte(`{"type":"email","data":{"to":"9jEw0@example.com","subject":"test subject","body":"test body"}}`),
		Status:          "pending",
		TemplateType:    string(template.TemplateTypeEmailVerification),
		RetryCount:      5,
	}
}
