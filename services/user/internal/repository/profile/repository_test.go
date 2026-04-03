package profile_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	coredb "github.com/rijum8906/relay/packages/core/db"
	migrations "github.com/rijum8906/relay/services/user/db/migrations"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/repository/profile"
)

var dbConf = coredb.Config{
	Host:     "localhost",
	Port:     5433,
	User:     "test_user",
	Password: "test_password",
	DBName:   "test_db",
	SSLMode:  "disable",
}

func TestProfileRepository_CreateProfile(t *testing.T) {
	repo, querier, ctx := newTestRepo(t)
	user := mustCreateUser(t, ctx, querier)

	created, appErr := repo.CreateProfile(ctx, db.CreateProfileParams{
		UserID:    user.ID,
		FirstName: "Jane",
		LastName:  "Doe",
		AvatarUrl: "https://example.com/avatar.png",
	})
	if appErr != nil {
		t.Fatalf("CreateProfile() unexpected error = %v", appErr)
	}

	if created.UserID != user.ID {
		t.Fatalf("CreateProfile() userID = %s, want %s", created.UserID, user.ID)
	}
	if created.FirstName != "Jane" || created.LastName != "Doe" {
		t.Fatalf("CreateProfile() names = %q %q", created.FirstName, created.LastName)
	}
}

func TestProfileRepository_GetProfile(t *testing.T) {
	repo, querier, ctx := newTestRepo(t)
	user := mustCreateUser(t, ctx, querier)
	created := mustCreateProfile(t, repo, ctx, user.ID)

	got, appErr := repo.GetProfile(ctx, user.ID)
	if appErr != nil {
		t.Fatalf("GetProfile() unexpected error = %v", appErr)
	}
	if got.ID != created.ID {
		t.Fatalf("GetProfile() id = %s, want %s", got.ID, created.ID)
	}

	_, appErr = repo.GetProfile(ctx, uuid.New())
	if appErr == nil {
		t.Fatalf("GetProfile() missing profile error = nil, want error")
	}
}

func TestProfileRepository_GetProfileByUserID(t *testing.T) {
	repo, querier, ctx := newTestRepo(t)
	user := mustCreateUser(t, ctx, querier)
	created := mustCreateProfile(t, repo, ctx, user.ID)

	got, appErr := repo.GetProfileByUserID(ctx, user.ID)
	if appErr != nil {
		t.Fatalf("GetProfileByUserID() unexpected error = %v", appErr)
	}
	if got.ID != created.ID {
		t.Fatalf("GetProfileByUserID() id = %s, want %s", got.ID, created.ID)
	}
}

func TestProfileRepository_UpdateProfileNames(t *testing.T) {
	repo, querier, ctx := newTestRepo(t)
	user := mustCreateUser(t, ctx, querier)
	created := mustCreateProfile(t, repo, ctx, user.ID)

	updated, appErr := repo.UpdateProfileNames(ctx, created.ID, "Updated", "Name")
	if appErr != nil {
		t.Fatalf("UpdateProfileNames() unexpected error = %v", appErr)
	}
	if updated.FirstName != "Updated" || updated.LastName != "Name" {
		t.Fatalf("UpdateProfileNames() names = %q %q", updated.FirstName, updated.LastName)
	}
}

func TestProfileRepository_UpdateProfileAvatar(t *testing.T) {
	repo, querier, ctx := newTestRepo(t)
	user := mustCreateUser(t, ctx, querier)
	created := mustCreateProfile(t, repo, ctx, user.ID)

	updated, appErr := repo.UpdateProfileAvatar(ctx, created.ID, "https://example.com/updated.png")
	if appErr != nil {
		t.Fatalf("UpdateProfileAvatar() unexpected error = %v", appErr)
	}
	if updated.AvatarUrl != "https://example.com/updated.png" {
		t.Fatalf("UpdateProfileAvatar() avatar = %q", updated.AvatarUrl)
	}
}

func TestProfileRepository_DeleteProfile(t *testing.T) {
	repo, querier, ctx := newTestRepo(t)
	user := mustCreateUser(t, ctx, querier)
	created := mustCreateProfile(t, repo, ctx, user.ID)

	appErr := repo.DeleteProfile(ctx, created.ID)
	if appErr != nil {
		t.Fatalf("DeleteProfile() unexpected error = %v", appErr)
	}

	_, appErr = repo.GetProfileByUserID(ctx, user.ID)
	if appErr == nil {
		t.Fatalf("GetProfileByUserID() after delete error = nil, want error")
	}
}

func newTestRepo(t *testing.T) (profile.ProfileRepository, *db.Queries, context.Context) {
	t.Helper()

	ctx := context.Background()
	pool, appErr := coredb.Connect(ctx, dbConf)
	if appErr != nil {
		t.Skipf("test database unavailable: %v", appErr)
	}
	t.Cleanup(pool.Close)

	querier := db.New(pool)
	applyMigrations(t, ctx, pool)
	resetTables(t, ctx, pool)

	return profile.NewProfileRepository(querier), querier, ctx
}

func applyMigrations(t *testing.T, ctx context.Context, pool db.DBTX) {
	t.Helper()

	allMigrations, err := migrations.All()
	if err != nil {
		t.Fatalf("migrations.All() error = %v", err)
	}

	for _, migration := range allMigrations {
		if _, err := pool.Exec(ctx, migration.Content); err != nil && !isIgnorableMigrationError(err) {
			t.Fatalf("apply migration %s error = %v", migration.Name, err)
		}
	}
}

func resetTables(t *testing.T, ctx context.Context, pool db.DBTX) {
	t.Helper()

	if _, err := pool.Exec(ctx, `TRUNCATE TABLE users RESTART IDENTITY CASCADE`); err != nil {
		t.Fatalf("resetTables() error = %v", err)
	}
}

func mustCreateUser(t *testing.T, ctx context.Context, querier *db.Queries) db.User {
	t.Helper()

	created, err := querier.CreateUser(ctx, db.CreateUserParams{
		Email: uniqueEmail(),
		PasswordHash: pgtype.Text{
			String: "password-" + uuid.NewString(),
			Valid:  true,
		},
	})
	if err != nil {
		t.Fatalf("CreateUser() setup error = %v", err)
	}
	return created
}

func mustCreateProfile(t *testing.T, repo profile.ProfileRepository, ctx context.Context, userID uuid.UUID) *db.Profile {
	t.Helper()

	created, appErr := repo.CreateProfile(ctx, db.CreateProfileParams{
		UserID:    userID,
		FirstName: "First",
		LastName:  "Last",
		AvatarUrl: "",
	})
	if appErr != nil {
		t.Fatalf("CreateProfile() setup error = %v", appErr)
	}
	return created
}

func uniqueEmail() string {
	return "profile-" + uuid.NewString() + "@example.com"
}

func isIgnorableMigrationError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	switch pgErr.Code {
	case "42P07", "42710":
		return true
	default:
		return false
	}
}

func TestMain(m *testing.M) {
	time.Local = time.UTC
	os.Exit(m.Run())
}
