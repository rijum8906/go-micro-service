package handler

import (
	"context"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/metadata"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	sessionv1 "github.com/rijum8906/relay/packages/pb/user_service/session/v1"
	sessionservice "github.com/rijum8906/relay/services/user/internal/services/session"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

type SessionHandler struct {
	sessionv1.UnimplementedSessionServiceServer
	service sessionservice.SessionService
}

func NewSessionHandler(service sessionservice.SessionService) *SessionHandler {
	return &SessionHandler{
		service: service,
	}
}

func (h *SessionHandler) GetSessions(ctx context.Context, req *sessionv1.GetSessionsRequest) (*sessionv1.GetSessionsResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("get sessions request is required"))
	}
	
	userInfo, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		return nil, utils.MapAppError(apperror.ErrInternal.WithMessage("get session user metadata is required"))
	}
	
	result, appErr := h.service.GetSessions(ctx, req, &userInfo)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}

func (h *SessionHandler) GetActiveSessions(ctx context.Context, req *corev1.EmptyRequest) (*sessionv1.GetActiveSessionsResponse, error) {
	if req == nil {
		return nil, utils.MapAppError(apperror.ErrValidation.WithMessage("get active sessions request is required"))
	}

	metadata, ok := metadata.ReceiveUserInfo(ctx)
	if !ok {
		// NOTE: if not ok then it's the gateway's bug
		return nil, utils.MapAppError(apperror.ErrInternal.WithMessage("logout user metadata is required"))
	}

	result, appErr := h.service.GetActiveSessions(ctx, &metadata)
	if appErr != nil {
		return nil, utils.MapAppError(appErr)
	}

	return result, nil
}
