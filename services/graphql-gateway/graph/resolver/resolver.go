package resolver

import (
	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/common/env"
	commonv1 "github.com/rijum8906/relay/packages/pb/common/v1"
	user_servicev1 "github.com/rijum8906/relay/packages/pb/user_service/v1"
	"github.com/rijum8906/relay/services/graphql-gateway/graph/model"
	"github.com/rijum8906/relay/services/graphql-gateway/internal/client"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	Env        *env.Env
	UserClient user_servicev1.AuthServiceClient
}

func NewResolver(env *env.Env) *Resolver {
	userClient := client.NewUserClient()
	return &Resolver{
		Env:        env,
		UserClient: userClient,
	}
}

func mapAuthPayload(resp *user_servicev1.AuthResponse) *model.AuthPayload {
	if resp == nil {
		return &model.AuthPayload{
			Account:  &model.AuthAccount{},
			Tokens:   &model.AuthToken{},
			Profiles: []*model.AuthProfile{},
		}
	}

	profiles := make([]*model.AuthProfile, 0, len(resp.GetProfiles()))
	for _, profile := range resp.GetProfiles() {
		profiles = append(profiles, mapAuthProfile(profile))
	}

	return &model.AuthPayload{
		Account:  mapAuthAccount(resp.GetAccount()),
		Tokens:   mapAuthToken(resp.GetTokens()),
		Profiles: profiles,
	}
}

func mapAuthAccount(account *user_servicev1.AccountResponse) *model.AuthAccount {
	if account == nil {
		return &model.AuthAccount{}
	}

	return &model.AuthAccount{
		ID:    parseUUID(account.GetId()),
		Email: valueOfEmail(account.GetEmail()),
	}
}

func mapAuthToken(token *user_servicev1.AuthTokenResponse) *model.AuthToken {
	if token == nil {
		return &model.AuthToken{}
	}

	return &model.AuthToken{
		AccessToken:  valueOfToken(token.GetAccessToken()),
		RefreshToken: valueOfToken(token.GetRefreshToken()),
	}
}

func mapAuthProfile(profile *user_servicev1.ProfileResponse) *model.AuthProfile {
	if profile == nil {
		return &model.AuthProfile{}
	}

	return &model.AuthProfile{
		ID:          valueOfUUID(profile.GetId()),
		FirstName:   valueOfName(profile.GetFirstName()),
		LastName:    valueOfName(profile.GetLastName()),
		DisplayName: optionalName(profile.GetDisplayName()),
		AvatarURL:   optionalURL(profile.GetAvatarUrl()),
	}
}

func mapResponse(success bool) *model.Response {
	return &model.Response{Success: success}
}

func parseUUID(id *commonv1.UUID) uuid.UUID {
	parsed, err := uuid.Parse(valueOfUUID(id))
	if err != nil {
		return uuid.Nil
	}

	return parsed
}

func valueOfUUID(id *commonv1.UUID) string {
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

func valueOfName(name *commonv1.Name) string {
	if name == nil {
		return ""
	}
	return name.GetValue()
}

func valueOfToken(token *commonv1.Token) string {
	if token == nil {
		return ""
	}
	return token.GetValue()
}

func optionalName(name *commonv1.Name) *string {
	if name == nil {
		return nil
	}
	value := name.GetValue()
	return &value
}

func optionalURL(url *commonv1.Url) *string {
	if url == nil {
		return nil
	}
	value := url.GetValue()
	return &value
}
