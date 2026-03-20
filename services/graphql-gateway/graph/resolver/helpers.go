package resolver

import (
	"context"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/google/uuid"
	commonv1 "github.com/rijum8906/relay/packages/pb/common/v1"
	accountv1 "github.com/rijum8906/relay/packages/pb/user_service/account/v1"
	userservicev1 "github.com/rijum8906/relay/packages/pb/user_service/v1"
	"github.com/rijum8906/relay/services/graphql-gateway/graph/model"
	"github.com/rijum8906/relay/services/graphql-gateway/internal/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func authorizationToken(ctx context.Context) string {
	opCtx := graphql.GetOperationContext(ctx)
	if opCtx == nil {
		return ""
	}

	header := strings.TrimSpace(opCtx.Headers.Get("Authorization"))
	token, ok := strings.CutPrefix(header, "Bearer ")
	if !ok {
		return ""
	}

	return strings.TrimSpace(token)
}

func requireAuthorizationToken(ctx context.Context) (string, error) {
	token := authorizationToken(ctx)
	if token == "" {
		return "", utils.NewAppError("authentication required", utils.CodeUnauthorized)
	}

	return token, nil
}

func authenticatedUserID(ctx context.Context) (string, error) {
	opCtx := graphql.GetOperationContext(ctx)
	if opCtx == nil {
		return "", utils.NewAppError("authentication required", utils.CodeUnauthorized)
	}

	userID := strings.TrimSpace(opCtx.Headers.Get("X-User-ID"))
	if userID == "" {
		return "", utils.NewAppError("authentication required", utils.CodeUnauthorized)
	}

	return userID, nil
}

func gqlAuthPayload(result *userservicev1.AuthenticationResult) (*model.AuthPayload, error) {
	if result == nil || result.Account == nil || result.Tokens == nil {
		return nil, utils.NewAppError("authentication result is incomplete", utils.CodeInternalServerError)
	}

	accountID, err := uuid.Parse(result.Account.GetId().GetValue())
	if err != nil {
		return nil, err
	}

	payload := &model.AuthPayload{
		Account: &model.AuthAccount{
			ID:    accountID,
			Email: result.Account.GetEmail().GetValue(),
		},
		Tokens: &model.AuthToken{
			AccessToken:  result.Tokens.GetAccessToken().GetValue(),
			RefreshToken: result.Tokens.GetRefreshToken().GetValue(),
		},
		Profiles: make([]*model.AuthProfile, 0, len(result.Profiles)),
	}

	for _, profile := range result.Profiles {
		if profile == nil {
			continue
		}

		payload.Profiles = append(payload.Profiles, &model.AuthProfile{
			ID:          profile.GetId().GetValue(),
			FirstName:   profile.GetFirstName().GetValue(),
			LastName:    profile.GetLastName().GetValue(),
			DisplayName: optionalString(profile.GetDisplayName().GetValue()),
			AvatarURL:   optionalString(profile.GetAvatarUrl().GetValue()),
		})
	}

	return payload, nil
}

func gqlResponse(success bool, message string) *model.Response {
	resp := &model.Response{Success: success}
	if strings.TrimSpace(message) != "" {
		resp.Message = &message
	}
	return resp
}

func gqlToken(token *commonv1.Token) *model.Token {
	if token == nil {
		return &model.Token{}
	}

	return &model.Token{
		Value:     token.GetValue(),
		ExpiresAt: timestampString(token.GetExpiresAt()),
	}
}

func gqlMyAccount(resp *accountv1.MyAccountResponse) (*model.MyAccount, error) {
	if resp == nil || resp.Account == nil {
		return nil, utils.NewAppError("account not found", utils.CodeNotFound)
	}

	accountID, err := uuid.Parse(resp.Account.GetId().GetValue())
	if err != nil {
		return nil, err
	}

	return &model.MyAccount{
		ID:    accountID,
		Email: resp.Account.GetEmail().GetValue(),
		Security: &model.AccountSecurity{
			IsEmailVerified:    resp.Account.GetIsEmailVerified(),
			EmailVerifiedAt:    timestampString(resp.Account.GetEmailVerifiedAt()),
			TwoFactorEnabled:   resp.Account.GetTwoFactorEnabled(),
			TwoFactorEnabledAt: timestampString(resp.Account.GetTwoFactorEnabledAt()),
		},
		CreatedAt: timestampStringValue(resp.Account.GetCreatedAt()),
		UpdatedAt: timestampStringValue(resp.Account.GetUpdatedAt()),
	}, nil
}

func scopedActionToProto(action model.ScopedAction) userservicev1.ScopedTokenScope {
	switch action {
	case model.ScopedActionScopedActionChangePassword:
		return userservicev1.ScopedTokenScope_SCOPED_TOKEN_SCOPE_CHANGE_PASSWORD
	case model.ScopedActionScopedActionChangeEmail:
		return userservicev1.ScopedTokenScope_SCOPED_TOKEN_SCOPE_CHANGE_EMAIL
	case model.ScopedActionScopedActionDeleteAccount:
		return userservicev1.ScopedTokenScope_SCOPED_TOKEN_SCOPE_DELETE_ACCOUNT
	case model.ScopedActionScopedActionDeleteProfile:
		return userservicev1.ScopedTokenScope_SCOPED_TOKEN_SCOPE_DELETE_PROFILE
	default:
		return userservicev1.ScopedTokenScope_SCOPED_TOKEN_SCOPE_UNSPECIFIED
	}
}

func optionalString(value string) *string {
	if value == "" {
		return nil
	}

	v := value
	return &v
}

func timestampString(ts *timestamppb.Timestamp) *string {
	if ts == nil || !ts.IsValid() {
		return nil
	}

	v := ts.AsTime().UTC().Format("2006-01-02T15:04:05Z")
	return &v
}

func timestampStringValue(ts *timestamppb.Timestamp) string {
	if value := timestampString(ts); value != nil {
		return *value
	}
	return ""
}

func authRequestMetadata(ctx context.Context, input *model.MetadataInput) (*commonv1.RequestMetadata, error) {
	metadata := requestMetadataFromContext(ctx, input)
	if metadata == nil {
		return nil, utils.NewAppError("metadata is required", utils.CodeBadUserInput)
	}

	return metadata, nil
}
