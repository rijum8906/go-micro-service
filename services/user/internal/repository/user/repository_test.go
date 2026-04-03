package user_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	coredb "github.com/rijum8906/relay/packages/core/db"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/repository/user"
)

var dbConf = coredb.Config{
	Host:     "localhost",
	Port:     5433,
	User:     "test_user",
	Password: "test_password",
	DBName:   "test_db",
	SSLMode:  "disable",
}

func TestUserRepository_CreateUser(t *testing.T) {
	repo, ctx := newTestRepo(t)

	req := newRegisterRequest()
	created, appErr := repo.CreateUser(ctx, req)
	if appErr != nil {
		t.Fatalf("CreateUser() unexpected error = %v", appErr)
	}

	if created.Email != req.Email {
		t.Fatalf("CreateUser() email = %q, want %q", created.Email, req.Email)
	}
	if !created.PasswordHash.Valid || created.PasswordHash.String != req.Password {
		t.Fatalf("CreateUser() password hash = %#v, want valid hash %q", created.PasswordHash, req.Password)
	}

	_, duplicateErr := repo.CreateUser(ctx, req)
	if duplicateErr == nil {
		t.Fatalf("CreateUser() duplicate email error = nil, want error")
	}
}

func TestUserRepository_GetUser(t *testing.T) {
	repo, ctx := newTestRepo(t)

	created := mustCreateUser(t, repo, ctx)

	got, appErr := repo.GetUser(ctx, created.ID)
	if appErr != nil {
		t.Fatalf("GetUser() unexpected error = %v", appErr)
	}
	if got.ID != created.ID {
		t.Fatalf("GetUser() id = %s, want %s", got.ID, created.ID)
	}

	_, appErr = repo.GetUser(ctx, uuid.New())
	if appErr == nil {
		t.Fatalf("GetUser() missing user error = nil, want error")
	}
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	repo, ctx := newTestRepo(t)

	created := mustCreateUser(t, repo, ctx)

	got, appErr := repo.GetUserByEmail(ctx, created.Email)
	if appErr != nil {
		t.Fatalf("GetUserByEmail() unexpected error = %v", appErr)
	}
	if got.ID != created.ID {
		t.Fatalf("GetUserByEmail() id = %s, want %s", got.ID, created.ID)
	}

	_, appErr = repo.GetUserByEmail(ctx, uniqueEmail())
	if appErr == nil {
		t.Fatalf("GetUserByEmail() missing user error = nil, want error")
	}
}

func TestUserRepository_UpdateUserPassword(t *testing.T) {
	repo, ctx := newTestRepo(t)

	created := mustCreateUser(t, repo, ctx)
	newPassword := "updated-password"

	appErr := repo.UpdateUserPassword(ctx, created.ID, newPassword)
	if appErr != nil {
		t.Fatalf("UpdateUserPassword() unexpected error = %v", appErr)
	}

	got, appErr := repo.GetUser(ctx, created.ID)
	if appErr != nil {
		t.Fatalf("GetUser() after UpdateUserPassword unexpected error = %v", appErr)
	}
	if got.PasswordHash.String != newPassword {
		t.Fatalf("UpdateUserPassword() password = %q, want %q", got.PasswordHash.String, newPassword)
	}
}

func TestUserRepository_UpdateUserEmail(t *testing.T) {
	repo, ctx := newTestRepo(t)

	created := mustCreateUser(t, repo, ctx)
	newEmail := uniqueEmail()

	appErr := repo.UpdateUserEmail(ctx, created.ID, newEmail)
	if appErr != nil {
		t.Fatalf("UpdateUserEmail() unexpected error = %v", appErr)
	}

	got, appErr := repo.GetUser(ctx, created.ID)
	if appErr != nil {
		t.Fatalf("GetUser() after UpdateUserEmail unexpected error = %v", appErr)
	}
	if got.Email != newEmail {
		t.Fatalf("UpdateUserEmail() email = %q, want %q", got.Email, newEmail)
	}
}

func TestUserRepository_VerifyUserEmail(t *testing.T) {
	repo, ctx := newTestRepo(t)

	created := mustCreateUser(t, repo, ctx)

	appErr := repo.VerifyUserEmail(ctx, created.ID)
	if appErr != nil {
		t.Fatalf("VerifyUserEmail() unexpected error = %v", appErr)
	}

	got, appErr := repo.GetUser(ctx, created.ID)
	if appErr != nil {
		t.Fatalf("GetUser() after VerifyUserEmail unexpected error = %v", appErr)
	}
	if !got.IsEmailVerified {
		t.Fatalf("VerifyUserEmail() IsEmailVerified = false, want true")
	}
	if !got.EmailVerifiedAt.Valid {
		t.Fatalf("VerifyUserEmail() EmailVerifiedAt.Valid = false, want true")
	}

	appErr = repo.VerifyUserEmail(ctx, uuid.New())
	if appErr == nil {
		t.Fatalf("VerifyUserEmail() missing user error = nil, want error")
	}
}

func TestUserRepository_DeleteUser(t *testing.T) {
	repo, ctx := newTestRepo(t)

	created := mustCreateUser(t, repo, ctx)

	appErr := repo.DeleteUser(ctx, created.ID)
	if appErr != nil {
		t.Fatalf("DeleteUser() unexpected error = %v", appErr)
	}

	_, appErr = repo.GetUser(ctx, created.ID)
	if appErr == nil {
		t.Fatalf("GetUser() after DeleteUser error = nil, want error")
	}
}

func newTestRepo(t *testing.T) (user.UserRepository, context.Context) {
	t.Helper()

	ctx := context.Background()
	pool, appErr := coredb.Connect(ctx, dbConf)
	if appErr != nil {
		t.Skipf("test database unavailable: %v", appErr)
	}
	t.Cleanup(pool.Close)

	resetUsersTable(t, ctx, pool)

	return user.NewAuthRepository(db.New(pool)), ctx
}

func resetUsersTable(t *testing.T, ctx context.Context, pool db.DBTX) {
	t.Helper()

	if _, err := pool.Exec(ctx, `TRUNCATE TABLE users RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("resetUsersTable() error = %v", err)
	}
}

func mustCreateUser(t *testing.T, repo user.UserRepository, ctx context.Context) *db.User {
	t.Helper()

	created, appErr := repo.CreateUser(ctx, newRegisterRequest())
	if appErr != nil {
		t.Fatalf("CreateUser() setup error = %v", appErr)
	}
	return created
}

func newRegisterRequest() *authv1.RegisterRequest {
	return &authv1.RegisterRequest{
		FirstName: "Test",
		LastName:  "User",
		Email:     uniqueEmail(),
		Password:  "password-" + uuid.NewString(),
	}
}

func uniqueEmail() string {
	return "user-" + uuid.NewString() + "@example.com"
}

func TestMain(m *testing.M) {
	time.Local = time.UTC
	os.Exit(m.Run())
}
