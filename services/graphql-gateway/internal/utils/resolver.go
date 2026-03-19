package utils

import (
	"time"

	"github.com/google/uuid"
	commonv1 "github.com/rijum8906/relay/packages/pb/common/v1"
	user_servicev1 "github.com/rijum8906/relay/packages/pb/user_service/v1"
	"github.com/rijum8906/relay/services/graphql-gateway/graph/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MapAuthPayload(resp *user_servicev1.AuthResponse) *model.AuthPayload {
	if resp == nil {
		return &model.AuthPayload{
			Account:  &model.AuthAccount{},
			Tokens:   &model.AuthToken{},
			Profiles: []*model.AuthProfile{},
		}
	}

	profiles := make([]*model.AuthProfile, 0, len(resp.GetProfiles()))
	for _, profile := range resp.GetProfiles() {
		profiles = append(profiles, MapAuthProfile(profile))
	}

	return &model.AuthPayload{
		Account:  MapAuthAccount(resp.GetAccount()),
		Tokens:   MapAuthToken(resp.GetTokens()),
		Profiles: profiles,
	}
}

func MapAuthAccount(account *user_servicev1.AccountResponse) *model.AuthAccount {
	if account == nil {
		return &model.AuthAccount{}
	}

	return &model.AuthAccount{
		ID:    ParseUUID(account.GetId()),
		Email: valueOfEmail(account.GetEmail()),
	}
}

func MapAuthToken(token *user_servicev1.AuthTokenResponse) *model.AuthToken {
	if token == nil {
		return &model.AuthToken{}
	}

	return &model.AuthToken{
		AccessToken:  ValueOfToken(token.GetAccessToken()),
		RefreshToken: ValueOfToken(token.GetRefreshToken()),
	}
}

func MapAuthProfile(profile *user_servicev1.ProfileResponse) *model.AuthProfile {
	if profile == nil {
		return &model.AuthProfile{}
	}

	return &model.AuthProfile{
		ID:          ValueOfUUID(profile.GetId()),
		FirstName:   ValueOfName(profile.GetFirstName()),
		LastName:    ValueOfName(profile.GetLastName()),
		DisplayName: OptionalName(profile.GetDisplayName()),
		AvatarURL:   OptionalURL(profile.GetAvatarUrl()),
	}
}

func MapResponse(success bool) *model.Response {
	return &model.Response{Success: success}
}

func MapToken(token *commonv1.Token) *model.Token {
	if token == nil {
		return &model.Token{}
	}

	return &model.Token{
		Value:     token.GetValue(),
		ExpiresAt: OptionalTimestamp(token.GetExpiresAt()),
	}
}

func MapMyAccount(resp *user_servicev1.GetMyAccountResponse) *model.MyAccount {
	if resp == nil {
		return &model.MyAccount{}
	}

	return &model.MyAccount{
		ID:        ParseUUID(resp.GetId()),
		Email:     valueOfEmail(resp.GetEmail()),
		Security:  MapAccountSecurity(resp.GetSecurity()),
		CreatedAt: TimestampString(resp.GetCreatedAt()),
		UpdatedAt: TimestampString(resp.GetUpdatedAt()),
	}
}

func MapAccountSecurity(resp *user_servicev1.GetMyAccountSecurityResponse) *model.AccountSecurity {
	if resp == nil {
		return nil
	}

	return &model.AccountSecurity{
		IsEmailVerified:    resp.GetIsEmailVerified(),
		EmailVerifiedAt:    OptionalTimestamp(resp.GetEmailVerifiedAt()),
		TwoFactorEnabled:   resp.GetTwoFactorEnabled(),
		TwoFactorEnabledAt: OptionalTimestamp(resp.GetTwoFactorEnabledAt()),
	}
}

func ParseUUID(id *commonv1.UUID) uuid.UUID {
	parsed, err := uuid.Parse(ValueOfUUID(id))
	if err != nil {
		return uuid.Nil
	}

	return parsed
}

func ValueOfUUID(id *commonv1.UUID) string {
	if id == nil {
		return ""
	}
	return id.GetValue()
}

func valueOfEmail(email *commonv1.Email) string {
	if email == nil {
		return ""
	}
	return email.GetValue()
}

func ValueOfName(name *commonv1.Name) string {
	if name == nil {
		return ""
	}
	return name.GetValue()
}

func ValueOfToken(token *commonv1.Token) string {
	if token == nil {
		return ""
	}
	return token.GetValue()
}

func TimestampString(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}

	return ts.AsTime().UTC().Format(time.RFC3339Nano)
}

func OptionalTimestamp(ts *timestamppb.Timestamp) *string {
	if ts == nil {
		return nil
	}

	value := TimestampString(ts)
	return &value
}

func OptionalName(name *commonv1.Name) *string {
	if name == nil {
		return nil
	}
	value := name.GetValue()
	return &value
}

func OptionalURL(url *commonv1.Url) *string {
	if url == nil {
		return nil
	}
	value := url.GetValue()
	return &value
}
