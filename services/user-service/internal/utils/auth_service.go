package utils

import (
	"github.com/jackc/pgx/v5/pgtype"
	authv1 "github.com/rijum8906/relay/packages/pb/user_service/auth/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/user_service/models/v1"
	"github.com/rijum8906/relay/services/user/internal/db"
	"google.golang.org/protobuf/types/known/timestamppb"
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
		Id:              user.ID.String(),
		Email:           user.Email,
		IsEmailVerified: user.IsEmailVerified,
		EmailVerifiedAt: mapTimestamp(user.EmailVerifiedAt),
		CreatedAt:       mapTimestamp(user.CreatedAt),
		UpdatedAt:       mapTimestamp(user.UpdatedAt),
	}
}

func MapProfile(profile *db.Profile) *modelsv1.Profile {
	return &modelsv1.Profile{
		Id:        profile.ID.String(),
		UserId:    profile.UserID.String(),
		FirstName: profile.FirstName,
		LastName:  profile.LastName,
		AvatarUrl: profile.AvatarUrl,
		CreatedAt: mapTimestamp(profile.CreatedAt),
		UpdatedAt: mapTimestamp(profile.UpdatedAt),
	}
}

func MapSession(session *db.Session) *modelsv1.Session {
	return &modelsv1.Session{
		Id:          session.ID.String(),
		UserId:      session.UserID.String(),
		UserAgent:   session.UserAgent,
		IpAddress:   session.IpAddr,
		DeviceId:    session.DeviceID,
		LastLoginAt: timestamppb.New(session.LastLoginAt.Time),
		CreatedAt:   timestamppb.New(session.CreatedAt.Time),
		UpdatedAt:   timestamppb.New(session.UpdatedAt.Time),
	}
}

func MapAuthResponse(user *db.User, profile *db.Profile, accessToken, refreshToken string) *authv1.AuthResponse {
	return &authv1.AuthResponse{
		Status:  authv1.AuthStatus_AUTH_STATUS_SUCCESS,
		Tokens:  ParseTokens(accessToken, refreshToken),
		User:    MapUser(user),
		Profile: MapProfile(profile),
	}
}

func mapTimestamp(ts pgtype.Timestamptz) *timestamppb.Timestamp {
	if !ts.Valid {
		return nil
	}

	return timestamppb.New(ts.Time)
}
