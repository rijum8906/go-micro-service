package session_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rijum8906/relay/packages/core/testutils"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/repository/session"
	"github.com/rijum8906/relay/services/user/internal/repository/user"
)

func TestMain(m *testing.M) {
	time.Local = time.UTC
	os.Exit(m.Run())
}

func TestSessionRepository_CreateAndGetSession(t *testing.T) {
	repos := createRepo()
	user := mustCreateUser(repos)
	session1 := mustCreateSession(repos, user)

	// Get Session By ID
	fetchedSess, appErr := repos.Session.GetSession(context.Background(), session1.ID)
	if appErr != nil {
		t.Fatalf("failed to get session: %v", appErr)
	}
	if !verifySession(session1, fetchedSess) {
		t.Fatalf("GetSession() session = %v, want %v", fetchedSess, session1)
	}

	// Get Session By RefreshTokenHash
	fetchedSess2, appErr := repos.Session.GetSessionByRefreshToken(context.Background(), session1.RefreshTokenHash)
	if appErr != nil {
		t.Fatalf("failed to get session: %v", appErr)
	}
	if !verifySession(session1, fetchedSess2) {
		t.Fatalf("GetSessionByRefreshToken() session = %v, want %v", fetchedSess2, session1)
	}
}

// NOTE: Not testing Revoke Other Sessions

func TestSessionRepository_CreateAndRevokeSession(t *testing.T) {
	repos := createRepo()
	user := mustCreateUser(repos)
	session1 := mustCreateSession(repos, user)
	session2 := mustCreateSession(repos, user)
	session3 := mustCreateSession(repos, user)

	// Revoke session1
	appErr := repos.Session.RevokeSession(context.Background(), session1.ID)
	if appErr != nil {
		t.Fatalf("failed to revoke session: %v", appErr)
	}
	fetchedSession1, appErr := repos.Session.GetSession(context.Background(), session1.ID)
	if appErr != nil {
		t.Fatalf("failed to revoke session: %v", appErr)
	}
	if !verifySession(session1, fetchedSession1) || !fetchedSession1.IsRevoked {
		t.Fatalf("GetSession() after revoke error = %v, want %v", fetchedSession1, session1)
	}

	// Revoke Multiple sessions
	appErr = repos.Session.RevokeAllSessions(context.Background(), user.ID)
	if appErr != nil {
		t.Fatalf("failed to revoke session: %v", appErr)
	}

	// Try to retrive session 2 and 3
	fetchedSession2, appErr := repos.Session.GetSession(context.Background(), session2.ID)
	if appErr != nil {
		t.Fatalf("failed to revoke session: %v", appErr)
	}
	if !verifySession(session2, fetchedSession2) || !fetchedSession2.IsRevoked {
		t.Fatalf("GetSession() after revoke error = %v, want %v", fetchedSession1, session1)
	}

	fetchedSession3, appErr := repos.Session.GetSession(context.Background(), session3.ID)
	if appErr != nil {
		t.Fatalf("failed to revoke session: %v", appErr)
	}
	if !verifySession(session3, fetchedSession3) || !fetchedSession3.IsRevoked {
		t.Fatalf("GetSession() after revoke error = %v, want %v", fetchedSession1, session1)
	}
}

func TestSessionRepository_GetActiveSessions(t *testing.T) {
	repos := createRepo()
	user := mustCreateUser(repos)
	session1 := mustCreateSession(repos, user)
	session2 := mustCreateSession(repos, user)
	session3 := mustCreateSession(repos, user)

	// Revoke session 1
	appErr := repos.Session.RevokeSession(context.Background(), session1.ID)
	if appErr != nil {
		t.Fatalf("failed to revoke session: %v", appErr)
	}

	// Get Active sessions
	sessions, appErr := repos.Session.GetActiveSessions(context.Background(), user.ID, 10, 0)
	if appErr != nil {
		t.Fatalf("failed to get active sessions: %v", appErr)
	}

	if len(*sessions) != 2 {
		t.Fatalf("GetActiveSessions() sessions = %v, want %v", len(*sessions), 2)
	}

	if verifySession(&(*sessions)[0], session2) || verifySession(&(*sessions)[1], session3) {
		t.Fatalf("GetActiveSessions() sessions = %v, want %v", sessions, []*db.Session{session2, session3})
	}
}

func TestSessionRepository_RevokeOtherSessions(t *testing.T) {
	repos := createRepo()
	user := mustCreateUser(repos)
	session1 := mustCreateSession(repos, user)
	session2 := mustCreateSession(repos, user)
	session3 := mustCreateSession(repos, user)

	appErr := repos.Session.RevokeOtherSessions(context.Background(), user.ID, session1.ID)
	if appErr != nil {
		t.Fatalf("failed to revoke other sessions: %v", appErr)
	}

	// NOTE: this session is not revoked
	fetchedSession1, appErr := repos.Session.GetSession(context.Background(), session1.ID)
	if appErr != nil {
		t.Fatalf("failed to revoke session: %v", appErr)
	}

	if !verifySession(session1, fetchedSession1) {
		t.Fatalf("GetSession() after revoke error = %v, want %v", fetchedSession1, session1)
	}

	fetchedSession2, appErr := repos.Session.GetSession(context.Background(), session2.ID)
	if appErr != nil {
		t.Fatalf("failed to revoke session: %v", appErr)
	}
	if !verifySession(session2, fetchedSession2) || !fetchedSession2.IsRevoked {
		t.Fatalf("GetSession() after revoke error = %v, want %v", fetchedSession1, session1)
	}

	fetchedSession3, appErr := repos.Session.GetSession(context.Background(), session3.ID)
	if appErr != nil {
		t.Fatalf("failed to revoke session: %v", appErr)
	}
	if !verifySession(session3, fetchedSession3) || !fetchedSession3.IsRevoked {
		t.Fatalf("GetSession() after revoke error = %v, want %v", fetchedSession1, session1)
	}
}

// TODO: add edge cases

// Helper functions

func verifySession(sess1 *db.Session, sess2 *db.Session) bool {
	return sess1.ID == sess2.ID && sess1.UserID == sess2.UserID
}

func mustCreateUser(repos *Repos) *db.User {
	u, appErr := repos.User.CreateUser(context.Background(), &authv1.RegisterRequest{
		FirstName: testutils.GenerateRandomString(10),
		LastName:  testutils.GenerateRandomString(10),
		Password:  testutils.GenerateRandomString(10), // NOTE: not hashed password
		Email:     testutils.GenerateRandomString(10),
	})
	if appErr != nil {
		panic(appErr)
	}

	return u
}

func mustCreateSession(repos *Repos, u *db.User) *db.Session {
	sess, appErr := repos.Session.CreateSession(context.Background(), db.CreateSessionParams{
		UserID:           u.ID,
		RefreshTokenHash: testutils.GenerateRandomString(32),
		UserAgent:        testutils.GenerateRandomString(10),
		IpAddr:           testutils.GenerateRandomString(10),
		DeviceID:         testutils.GenerateRandomString(10),
		ExpiresAt: pgtype.Timestamptz{
			Valid: true,
			Time:  time.Now().Add(time.Hour),
		},
	})
	if appErr != nil {
		panic(appErr)
	}

	return sess
}

type Repos struct {
	Session session.SessionRepository
	User    user.UserRepository
}

func createRepo() *Repos {
	pool := testutils.MustConnectDB()
	querier := db.New(pool)

	return &Repos{
		Session: session.NewSessionRepository(querier),
		User:    user.NewAuthRepository(querier),
	}
}
