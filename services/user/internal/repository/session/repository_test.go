package session_test

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
	"github.com/rijum8906/relay/services/user/internal/repository/session"
)

var dbConf = coredb.Config{
	Host:     "localhost",
	Port:     5433,
	User:     "test_user",
	Password: "test_password",
	DBName:   "test_db",
	SSLMode:  "disable",
}

func TestSessionRepository_CreateAndGetSession(t *testing.T) {
	repo, querier, ctx := newTestRepo(t)
	user := mustCreateUser(t, ctx, querier)

	created := mustCreateSession(t, repo, ctx, user.ID, "refresh-token-1", time.Now().Add(time.Hour))

	got, appErr := repo.GetSession(ctx, created.ID)
	if appErr != nil {
		t.Fatalf("GetSession() unexpected error = %v", appErr)
	}
	if got.ID != created.ID {
		t.Fatalf("GetSession() id = %s, want %s", got.ID, created.ID)
	}

	gotByToken, appErr := repo.GetSessionByRefreshToken(ctx, created.RefreshTokenHash)
	if appErr != nil {
		t.Fatalf("GetSessionByRefreshToken() unexpected error = %v", appErr)
	}
	if gotByToken.ID != created.ID {
		t.Fatalf("GetSessionByRefreshToken() id = %s, want %s", gotByToken.ID, created.ID)
	}
}

func TestSessionRepository_GetActiveSessions(t *testing.T) {
	repo, querier, ctx := newTestRepo(t)
	user := mustCreateUser(t, ctx, querier)

	first := mustCreateSession(t, repo, ctx, user.ID, "refresh-token-1", time.Now().Add(2*time.Hour))
	time.Sleep(10 * time.Millisecond)
	second := mustCreateSession(t, repo, ctx, user.ID, "refresh-token-2", time.Now().Add(3*time.Hour))

	sessions, appErr := repo.GetActiveSessions(ctx, user.ID, 10, 0)
	if appErr != nil {
		t.Fatalf("GetActiveSessions() unexpected error = %v", appErr)
	}
	if len(*sessions) != 2 {
		t.Fatalf("GetActiveSessions() len = %d, want 2", len(*sessions))
	}
	if (*sessions)[0].ID != second.ID || (*sessions)[1].ID != first.ID {
		t.Fatalf("GetActiveSessions() order = [%s %s], want [%s %s]", (*sessions)[0].ID, (*sessions)[1].ID, second.ID, first.ID)
	}

	paged, appErr := repo.GetActiveSessions(ctx, user.ID, 1, 1)
	if appErr != nil {
		t.Fatalf("GetActiveSessions() paged unexpected error = %v", appErr)
	}
	if len(*paged) != 1 || (*paged)[0].ID != first.ID {
		t.Fatalf("GetActiveSessions() paged result invalid")
	}
}

func TestSessionRepository_RevokeSession(t *testing.T) {
	repo, querier, ctx := newTestRepo(t)
	user := mustCreateUser(t, ctx, querier)
	created := mustCreateSession(t, repo, ctx, user.ID, "refresh-token-1", time.Now().Add(time.Hour))

	appErr := repo.RevokeSession(ctx, created.ID)
	if appErr != nil {
		t.Fatalf("RevokeSession() unexpected error = %v", appErr)
	}

	got, appErr := repo.GetSession(ctx, created.ID)
	if appErr != nil {
		t.Fatalf("GetSession() after revoke unexpected error = %v", appErr)
	}
	if !got.IsRevoked {
		t.Fatalf("RevokeSession() IsRevoked = false, want true")
	}
}

func TestSessionRepository_RevokeAllSessions(t *testing.T) {
	repo, querier, ctx := newTestRepo(t)
	user := mustCreateUser(t, ctx, querier)
	first := mustCreateSession(t, repo, ctx, user.ID, "refresh-token-1", time.Now().Add(time.Hour))
	second := mustCreateSession(t, repo, ctx, user.ID, "refresh-token-2", time.Now().Add(time.Hour))

	appErr := repo.RevokeAllSessions(ctx, user.ID)
	if appErr != nil {
		t.Fatalf("RevokeAllSessions() unexpected error = %v", appErr)
	}

	assertSessionRevoked(t, repo, ctx, first.ID, true)
	assertSessionRevoked(t, repo, ctx, second.ID, true)
}

func TestSessionRepository_RevokeOtherSessions(t *testing.T) {
	repo, querier, ctx := newTestRepo(t)
	user := mustCreateUser(t, ctx, querier)
	current := mustCreateSession(t, repo, ctx, user.ID, "refresh-token-1", time.Now().Add(time.Hour))
	other := mustCreateSession(t, repo, ctx, user.ID, "refresh-token-2", time.Now().Add(time.Hour))

	appErr := repo.RevokeOtherSessions(ctx, user.ID, current.ID)
	if appErr != nil {
		t.Fatalf("RevokeOtherSessions() unexpected error = %v", appErr)
	}

	assertSessionRevoked(t, repo, ctx, current.ID, false)
	assertSessionRevoked(t, repo, ctx, other.ID, true)
}

func TestSessionRepository_TerminateExpiredSessions(t *testing.T) {
	repo, querier, ctx := newTestRepo(t)
	user := mustCreateUser(t, ctx, querier)
	expired := mustCreateSession(t, repo, ctx, user.ID, "refresh-token-expired", time.Now().Add(-time.Hour))
	active := mustCreateSession(t, repo, ctx, user.ID, "refresh-token-active", time.Now().Add(time.Hour))

	appErr := repo.TerminateExpiredSessions(ctx)
	if appErr != nil {
		t.Fatalf("TerminateExpiredSessions() unexpected error = %v", appErr)
	}

	_, appErr = repo.GetSession(ctx, expired.ID)
	if appErr == nil {
		t.Fatalf("GetSession() expired session error = nil, want error")
	}

	_, appErr = repo.GetSession(ctx, active.ID)
	if appErr != nil {
		t.Fatalf("GetSession() active session unexpected error = %v", appErr)
	}
}

func newTestRepo(t *testing.T) (session.SessionRepository, *db.Queries, context.Context) {
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

	return session.NewSessionRepository(querier), querier, ctx
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

func mustCreateSession(t *testing.T, repo session.SessionRepository, ctx context.Context, userID uuid.UUID, refreshToken string, expiresAt time.Time) *db.Session {
	t.Helper()

	created, appErr := repo.CreateSession(ctx, db.CreateSessionParams{
		UserID:           userID,
		RefreshTokenHash: refreshToken,
		UserAgent:        "test-agent",
		IpAddr:           "127.0.0.1",
		DeviceID:         "device-" + uuid.NewString(),
		ExpiresAt: pgtype.Timestamptz{
			Time:  expiresAt,
			Valid: true,
		},
	})
	if appErr != nil {
		t.Fatalf("CreateSession() setup error = %v", appErr)
	}
	return created
}

func assertSessionRevoked(t *testing.T, repo session.SessionRepository, ctx context.Context, id uuid.UUID, want bool) {
	t.Helper()

	got, appErr := repo.GetSession(ctx, id)
	if appErr != nil {
		t.Fatalf("GetSession() unexpected error = %v", appErr)
	}
	if got.IsRevoked != want {
		t.Fatalf("GetSession().IsRevoked = %v, want %v", got.IsRevoked, want)
	}
}

func uniqueEmail() string {
	return "session-" + uuid.NewString() + "@example.com"
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
