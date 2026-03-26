// Package utils
package utils

import (
	modelsv1 "github.com/rijum8906/relay/packages/pb/user/models/v1"
	"github.com/rijum8906/relay/services/graphql-gateway/graph/model"
)

func MapUser(user *modelsv1.User) *model.User {
	emailVerfiedAt := user.EmialVerifiedAt.String()
	twoFactorEnabledAt := user.TwoFactorEnabledAt.String()

	return &model.User{
		ID:                 user.Id,
		Email:              user.Email,
		TwoFactorEnabled:   user.TwoFactorEnabled,
		IsEmailVerified:    user.IsEmailVerified,
		EmailVerifiedAt:    &emailVerfiedAt,
		TwoFactorEnabledAt: &twoFactorEnabledAt,
	}
}

func MapProfile(profile *modelsv1.Profile) *model.Profile {
	avatarURL := profile.AvatarUrl

	return &model.Profile{
		ID:        profile.Id,
		FirstName: profile.FirstName,
		LastName:  profile.LastName,
		AvatarURL: &avatarURL,
	}
}

func MapTokens(tokens *modelsv1.AuthToken) *model.AuthTokens {
	return &model.AuthTokens{
		AccessToken: &model.Token{
			Value:     tokens.AccessToken.Value,
			ExpiresIn: tokens.AccessToken.ExpiresIn.String(),
		},
		RefreshToken: &model.Token{
			Value:     tokens.RefreshToken.Value,
			ExpiresIn: tokens.RefreshToken.ExpiresIn.String(),
		},
	}
}

func MapAuthResponse(user *modelsv1.User, profile *modelsv1.Profile, tokens *modelsv1.AuthToken) *model.AuthResponse {
	return &model.AuthResponse{
		Tokens:  MapTokens(tokens),
		User:    MapUser(user),
		Profile: MapProfile(profile),
	}
}
