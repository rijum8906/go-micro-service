package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	modelsv1 "github.com/rijum8906/relay/packages/pb/user_service/models/v1"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

func (s *UserService) GetUser(ctx context.Context, req *corev1.EmptyRequest) (*modelsv1.User, error) {
	// Extract user information from authenticated context
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, apperror.ErrInternal.WithDetail("internal_message", "failed to retrieve user info from context")
	}

	userID, err := uuid.Parse(userInfo.UserID)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid user id").WithDetail("error", err.Error())
	}

	user, appErr := s.DBQ.GetUser(ctx, userID)
	if appErr != nil {
		return nil, appErr
	}

	return utils.MapUser(&user), nil
}
