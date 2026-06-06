package resolver

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	coreconstants "github.com/rijum8906/relay/packages/core/constants"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/metadata"
	"github.com/rijum8906/relay/packages/core/token"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	"github.com/rijum8906/relay/services/graphql-gateway/internal/dto/coredto"
	"github.com/rijum8906/relay/services/graphql-gateway/internal/utils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// attachClientInfo attach client info to context
func attachClientInfo(ctx context.Context, meta coredto.RequestMeta) (context.Context, *apperror.AppError) {
	browserInfo := utils.GetBrowserInfo(ctx)

	ctx = metadata.SetClientInfoToOutgoingContext(ctx, dto.ClientInfo{
		TraceID:   uuid.NewString(),
		UserAgent: browserInfo.UserAgent,
		IPAddress: browserInfo.IPAddr,
		DeviceID:  meta.DeviceId,
	})

	return ctx, nil
}

// validateAndAttachUserInfo validate the bearer token and on validation success attach user info to context
func validateAndAttachUserInfo(ctx context.Context, tokenManager token.TokenManager) (context.Context, *apperror.AppError) {
	accessToken, appErr := utils.GetAccessTokenFromHeader(ctx)
	if appErr != nil {
		return nil, appErr
	}

	claims, appErr := tokenManager.ValidateAuthToken(ctx, accessToken)
	if appErr != nil {
		return nil, appErr
	}

	ctx = metadata.SetUserInfoToOutgoingContext(ctx, dto.UserInfo{
		UserID:    claims.Subject,
		SessionID: claims.ID,
		TokenID:   claims.ID,
	})

	if ctx == nil {
		return nil, apperror.ErrInternal.WithMessage("Failed to set user info to outgoing context")
	}

	return ctx, nil
}

func parseScopedToken(method coreconstants.AuthMethod, scope string) (corev1.AuthMethod, corev1.TokenScope, *apperror.AppError) {
	authMethod, appErr := utils.ParseAuthMethod(method)
	if appErr != nil {
		return corev1.AuthMethod_AUTH_METHOD_UNSPECIFIED, corev1.TokenScope_TOKEN_SCOPE_UNSPECIFIED, appErr
	}
	tokenScope, appErr := utils.ParseScope(scope)
	if appErr != nil {
		return corev1.AuthMethod_AUTH_METHOD_UNSPECIFIED, corev1.TokenScope_TOKEN_SCOPE_UNSPECIFIED, appErr
	}

	return authMethod, tokenScope, nil
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}

	return strings.TrimSpace(*value)
}

func parseOptionalDateTime(value *string, field string) (*timestamppb.Timestamp, *apperror.AppError) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339, strings.TrimSpace(*value))
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid datetime").WithDetail("field", field).WithDetail("error", err.Error())
	}

	return timestamppb.New(parsed), nil
}

func getTokensFromCtx(ctx context.Context) (*dto.AuthTokens, *apperror.AppError) {
	accessToken, ok := ctx.Value(dto.MetaAccessTokenKey).(string)
	if !ok {
		return nil, apperror.ErrValidation.WithMessage("access token not found in req header")
	}
	refreshToken, ok := ctx.Value(dto.MetaRefreshTokenKey).(string)
	if !ok {
		return nil, apperror.ErrValidation.WithMessage("refresh token not found in req header")
	}
	return &dto.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
