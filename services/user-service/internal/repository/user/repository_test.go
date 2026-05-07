package user_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/testutils"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/repository/user"
)

func TestMain(m *testing.M) {
	time.Local = time.UTC
	os.Exit(m.Run())
}

func TestUserRepository_CreateUser(t *testing.T) {
	repos := createRepo()

	req := newRegisterRequest()
	created, appErr := repos.User.CreateUser(context.Background(), req)
	if appErr != nil {
		t.Fatalf("CreateUser() unexpected error = %v", appErr)
	}

	if created.Email != req.Email {
		t.Fatalf("CreateUser() email = %q, want %q", created.Email, req.Email)
	}
	if !created.PasswordHash.Valid || created.PasswordHash.String != req.Password {
		t.Fatalf("CreateUser() password hash = %#v, want valid hash %q", created.PasswordHash, req.Password)
	}

	_, duplicateErr := repos.User.CreateUser(context.Background(), req)
	if duplicateErr == nil {
		t.Fatalf("CreateUser() duplicate email error = nil, want error")
	}
}

func TestUserRepository_GetUser(t *testing.T) {
	repos := createRepo()
	created := mustCreateUser(repos)

	got, appErr := repos.User.GetUser(context.Background(), created.ID)
	if appErr != nil {
		t.Fatalf("GetUser() unexpected error = %v", appErr)
	}
	if got.ID != created.ID {
		t.Fatalf("GetUser() id = %s, want %s", got.ID, created.ID)
	}

	_, appErr = repos.User.GetUser(context.Background(), uuid.New())
	if appErr == nil {
		t.Fatalf("GetUser() missing user error = nil, want error")
	}
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	repos := createRepo()
	created := mustCreateUser(repos)

	got, appErr := repos.User.GetUserByEmail(context.Background(), created.Email)
	if appErr != nil {
		t.Fatalf("GetUserByEmail() unexpected error = %v", appErr)
	}
	if got.ID != created.ID {
		t.Fatalf("GetUserByEmail() id = %s, want %s", got.ID, created.ID)
	}

	_, appErr = repos.User.GetUserByEmail(context.Background(), uniqueEmail())
	if appErr == nil {
		t.Fatalf("GetUserByEmail() missing user error = nil, want error")
	}
}

func TestUserRepository_UpdateUserPassword(t *testing.T) {
	repos := createRepo()
	created := mustCreateUser(repos)
	newPassword := "updated-password"

	appErr := repos.User.UpdateUserPassword(context.Background(), created.ID, newPassword)
	if appErr != nil {
		t.Fatalf("UpdateUserPassword() unexpected error = %v", appErr)
	}

	got, appErr := repos.User.GetUser(context.Background(), created.ID)
	if appErr != nil {
		t.Fatalf("GetUser() after UpdateUserPassword unexpected error = %v", appErr)
	}
	if got.PasswordHash.String != newPassword {
		t.Fatalf("UpdateUserPassword() password = %q, want %q", got.PasswordHash.String, newPassword)
	}
}

func TestUserRepository_UpdateUserEmail(t *testing.T) {
	repos := createRepo()
	created := mustCreateUser(repos)
	newEmail := uniqueEmail()

	appErr := repos.User.UpdateUserEmail(context.Background(), created.ID, newEmail)
	if appErr != nil {
		t.Fatalf("UpdateUserEmail() unexpected error = %v", appErr)
	}

	got, appErr := repos.User.GetUser(context.Background(), created.ID)
	if appErr != nil {
		t.Fatalf("GetUser() after UpdateUserEmail unexpected error = %v", appErr)
	}
	if got.Email != newEmail {
		t.Fatalf("UpdateUserEmail() email = %q, want %q", got.Email, newEmail)
	}
}

func TestUserRepository_VerifyUserEmail(t *testing.T) {
	repos := createRepo()
	created := mustCreateUser(repos)

	appErr := repos.User.VerifyUserEmail(context.Background(), created.ID)
	if appErr != nil {
		t.Fatalf("VerifyUserEmail() unexpected error = %v", appErr)
	}

	got, appErr := repos.User.GetUser(context.Background(), created.ID)
	if appErr != nil {
		t.Fatalf("GetUser() after VerifyUserEmail unexpected error = %v", appErr)
	}
	if !got.IsEmailVerified {
		t.Fatalf("VerifyUserEmail() IsEmailVerified = false, want true")
	}
	if !got.EmailVerifiedAt.Valid {
		t.Fatalf("VerifyUserEmail() EmailVerifiedAt.Valid = false, want true")
	}

	appErr = repos.User.VerifyUserEmail(context.Background(), uuid.New())
	if appErr == nil {
		t.Fatalf("VerifyUserEmail() missing user error = nil, want error")
	}
}

func TestUserRepository_DeleteUser(t *testing.T) {
	repos := createRepo()
	created := mustCreateUser(repos)

	appErr := repos.User.DeleteUser(context.Background(), created.ID)
	if appErr != nil {
		t.Fatalf("DeleteUser() unexpected error = %v", appErr)
	}

	_, appErr = repos.User.GetUser(context.Background(), created.ID)
	if appErr == nil {
		t.Fatalf("GetUser() after DeleteUser error = nil, want error")
	}
}

func TestUserRepository_CheckExists(t *testing.T) {
	repos := createRepo()
	user := mustCreateUser(repos)

	exists, appErr := repos.User.CheckExists(context.Background(), uuid.New())
	if appErr != nil {
		t.Fatalf("ExistsUser() unexpected error = %v", appErr)
	}
	if exists {
		t.Fatalf("ExistsUser() exists = true, want false")
	}

	exists, appErr = repos.User.CheckExists(context.Background(), user.ID)
	if appErr != nil {
		t.Fatalf("ExistsUser() unexpected error = %v", appErr)
	}
	if !exists {
		t.Fatalf("ExistsUser() exists = false, want true")
	}

	t.Cleanup(func() {
		if err := repos.User.DeleteUser(context.Background(), user.ID); err != nil {
			t.Fatalf("DeleteUser() unexpected error = %v", err)
		}
	})
}

func TestUserRepository_CheckEmailExists(t *testing.T) {
	repos := createRepo()
	user := mustCreateUser(repos)

	exists, appErr := repos.User.CheckEmailExists(context.Background(), uniqueEmail())
	if appErr != nil {
		t.Fatalf("CheckEmailExists() unexpected error = %v", appErr)
	}
	if exists {
		t.Fatalf("CheckEmailExists() exists = true, want false")
	}

	exists, appErr = repos.User.CheckEmailExists(context.Background(), user.Email)
	if appErr != nil {
		t.Fatalf("CheckEmailExists() unexpected error = %v", appErr)
	}
	if !exists {
		t.Fatalf("CheckEmailExists() exists = false, want true")
	}

	t.Cleanup(func() {
		if err := repos.User.DeleteUser(context.Background(), user.ID); err != nil {
			t.Fatalf("DeleteUser() unexpected error = %v", err)
		}
	})
}

type Repos struct {
	User user.UserRepository
}

func createRepo() *Repos {
	pool := testutils.MustConnectDB(testutils.WithDBName(testutils.GetTestDBName("user-service")))
	querier := db.New(pool)

	return &Repos{
		User: user.NewAuthRepository(querier),
	}
}

func mustCreateUser(repos *Repos) *db.User {
	u, appErr := repos.User.CreateUser(context.Background(), newRegisterRequest())
	if appErr != nil {
		panic(appErr)
	}

	return u
}

func newRegisterRequest() *authv1.RegisterRequest {
	return &authv1.RegisterRequest{
		FirstName: testutils.GenerateRandomString(10),
		LastName:  testutils.GenerateRandomString(10),
		Email:     uniqueEmail(),
		Password:  testutils.GenerateRandomString(10),
	}
}

func uniqueEmail() string {
	return testutils.GenerateRandomString(10) + "@example.com"
}
