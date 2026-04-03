// Package utils
package utils

import (
	"time"

	modelsv1 "github.com/rijum8906/relay/packages/pb/user_service/models/v1"
	"github.com/rijum8906/relay/services/graphql-gateway/graph/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MapUser(user *modelsv1.User) *model.User {
	return &model.User{
		ID:                 user.Id,
		Email:              user.Email,
		TwoFactorEnabled:   user.TwoFactorEnabled,
		IsEmailVerified:    user.IsEmailVerified,
		EmailVerifiedAt:    optionalTimestamp(user.EmailVerifiedAt),
		TwoFactorEnabledAt: optionalTimestamp(user.TwoFactorEnabledAt),
		CreatedAt:          requiredTimestamp(user.CreatedAt),
		UpdatedAt:          requiredTimestamp(user.UpdatedAt),
	}
}

func MapProfile(profile *modelsv1.Profile) *model.Profile {
	var avatarURL *string
	if profile.AvatarUrl != "" {
		avatarURL = &profile.AvatarUrl
	}

	return &model.Profile{
		ID:        profile.Id,
		UserID:    profile.UserId,
		FirstName: profile.FirstName,
		LastName:  profile.LastName,
		CreatedAt: requiredTimestamp(profile.CreatedAt),
		UpdatedAt: requiredTimestamp(profile.UpdatedAt),
		AvatarURL: avatarURL,
	}
}

func MapSession(session *modelsv1.Session) *model.Session {
	return &model.Session{
		ID:           session.Id,
		UserID:       session.UserId,
		RefreshToken: "",
		DeviceID:     session.DeviceId,
		IPAddr:       session.IpAddress,
		CreatedAt:    requiredTimestamp(session.CreatedAt),
		UpdatedAt:    requiredTimestamp(session.UpdatedAt),
	}
}

func MapSessions(sessions []*modelsv1.Session) []*model.Session {
	mapped := make([]*model.Session, 0, len(sessions))
	for _, session := range sessions {
		if session == nil {
			continue
		}
		mapped = append(mapped, MapSession(session))
	}
	return mapped
}

func MapTokens(tokens *modelsv1.AuthToken) *model.AuthTokens {
	return &model.AuthTokens{
		AccessToken: &model.Token{
			Value:     tokens.AccessToken.Value,
			ExpiresAt: tokens.AccessToken.ExpiresAt.String(),
		},
		RefreshToken: &model.Token{
			Value:     tokens.RefreshToken.Value,
			ExpiresAt: tokens.RefreshToken.ExpiresAt.String(),
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

func requiredTimestamp(ts *timestamppb.Timestamp) string {
	if ts == nil || !ts.IsValid() {
		return ""
	}
	return ts.AsTime().Format(time.RFC3339)
}

func optionalTimestamp(ts *timestamppb.Timestamp) *string {
	if ts == nil || !ts.IsValid() {
		return nil
	}
	formatted := ts.AsTime().Format(time.RFC3339)
	return &formatted
}
