package utils

import (
	authv1 "github.com/rijum8906/relay/packages/pb/user/auth/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/user/models/v1"
	"github.com/rijum8906/relay/services/user/internal/db"
)

func ParseTokens(accessToken string, refreshToken string) *modelsv1.AuthToken {
	return &modelsv1.AuthToken{
		AccessToken: &modelsv1.Token{
			Value: accessToken,
		},
		RefreshToken: &modelsv1.Token{
			Value: refreshToken,
		},
	}
}

func MapUser(user *db.User) *modelsv1.User {
	return &modelsv1.User{
		Id:    user.ID.String(),
		Email: user.Email,
	}
}

func MapProfile(profile *db.Profile) *modelsv1.Profile {
	return &modelsv1.Profile{
		Id:        profile.ID.String(),
		FirstName: profile.FirstName,
		LastName:  profile.LastName,
	}
}

func MapAuthResponse(user *db.User, profile *db.Profile, accessToken, refreshToken string) *authv1.AuthResponse {
	return &authv1.AuthResponse{
		Tokens:  ParseTokens(accessToken, refreshToken),
		User:    MapUser(user),
		Profile: MapProfile(profile),
	}
}
