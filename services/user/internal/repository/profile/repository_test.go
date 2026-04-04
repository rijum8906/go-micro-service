package profile_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/rijum8906/relay/packages/core/testutils"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	"github.com/rijum8906/relay/services/user/internal/db"
	"github.com/rijum8906/relay/services/user/internal/repository/profile"
	"github.com/rijum8906/relay/services/user/internal/repository/session"
	"github.com/rijum8906/relay/services/user/internal/repository/user"
)

func TestMain(m *testing.M) {
	time.Local = time.UTC
	os.Exit(m.Run())
}

func TestProfileRepository_CreateAndGerProfile(t *testing.T) {
	repos := createRepo()
	user, profile := mustCreateUserAndProfile(repos)

	// Get profile
	fetchedProfile, appErr := repos.Profile.GetProfile(context.Background(), profile.ID)
	if appErr != nil {
		t.Fatalf("failed to get profile: %v", appErr)
	}
	if profile.ID != fetchedProfile.ID || user.ID != fetchedProfile.UserID {
		t.Fatalf("GetProfile() profile = %v, want %v", fetchedProfile, profile)
	}

	// Get Profile By UserID
	fetchedProfile2, appErr := repos.Profile.GetProfileByUserID(context.Background(), user.ID)
	if appErr != nil {
		t.Fatalf("failed to get profile: %v", appErr)
	}
	if profile.ID != fetchedProfile2.ID || user.ID != fetchedProfile2.UserID {
		t.Fatalf("GetProfileByUserID() profile = %v, want %v", fetchedProfile2, profile)
	}
}

func TestProfileRepository_CreateAndUpdateProfile(t *testing.T) {
	repos := createRepo()
	user, profile := mustCreateUserAndProfile(repos)

	updatedProfile, appErr := repos.Profile.UpdateProfileNames(context.Background(), profile.ID, "Anything", "LastName")
	if appErr != nil {
		t.Fatalf("failed to update profile: %v", appErr)
	}
	if profile.ID != updatedProfile.ID || user.ID != updatedProfile.UserID {
		t.Fatalf("UpdateProfileNames() profile = %v, want %v", updatedProfile, profile)
	}
	if profile.AvatarUrl != updatedProfile.AvatarUrl {
		t.Fatalf("UpdateProfileNames() profile = %v, want %v", updatedProfile, profile)
	}
	if updatedProfile.FirstName != "Anything" || updatedProfile.LastName != "LastName" {
		t.Fatalf("UpdateProfileNames() profile = %v, want %v", updatedProfile, profile)
	}

	updatedProfile2, appErr := repos.Profile.UpdateProfileAvatar(context.Background(), profile.ID, "Anything")
	if appErr != nil {
		t.Fatalf("failed to update profile: %v", appErr)
	}
	if profile.ID != updatedProfile2.ID || user.ID != updatedProfile2.UserID {
		t.Fatalf("UpdateProfileAvatarUrl() profile = %v, want %v", updatedProfile2, profile)
	}
	if updatedProfile2.AvatarUrl != "Anything" {
		t.Fatalf("UpdateProfileAvatarUrl() profile = %v, want %v", updatedProfile2, profile)
	}
}

func TestProfileRepository_DeleteProfile(t *testing.T) {
	repos := createRepo()
	user, profile := mustCreateUserAndProfile(repos)

	err := repos.Profile.DeleteProfile(context.Background(), profile.ID)
	if err != nil {
		t.Fatalf("failed to delete profile: %v", err)
	}

	fetchedProfile, appErr := repos.Profile.GetProfileByUserID(context.Background(), user.ID)
	if appErr == nil {
		t.Fatalf("failed to delete profile: %v", appErr)
	}
	if fetchedProfile != nil {
		t.Fatalf("GetProfile() profile = %v, want %v", fetchedProfile, nil)
	}
}

type Repos struct {
	Session session.SessionRepository
	User    user.UserRepository
	Profile profile.ProfileRepository
}

func createRepo() *Repos {
	pool := testutils.MustConnectDB()
	querier := db.New(pool)

	return &Repos{
		Session: session.NewSessionRepository(querier),
		User:    user.NewAuthRepository(querier),
		Profile: profile.NewProfileRepository(querier),
	}
}

func mustCreateUserAndProfile(repos *Repos) (*db.User, *db.Profile) {
	u, appErr := repos.User.CreateUser(context.Background(), &authv1.RegisterRequest{
		FirstName: testutils.GenerateRandomString(10),
		LastName:  testutils.GenerateRandomString(10),
		Password:  testutils.GenerateRandomString(10), // NOTE: not hashed password
		Email:     testutils.GenerateRandomString(10),
	})
	if appErr != nil {
		panic(appErr)
	}

	p, appErr := repos.Profile.CreateProfile(context.Background(), db.CreateProfileParams{
		UserID:    u.ID,
		FirstName: testutils.GenerateRandomString(10),
		LastName:  testutils.GenerateRandomString(10),
		AvatarUrl: testutils.GenerateRandomString(10),
	})
	if appErr != nil {
		panic(appErr)
	}

	return u, p
}
