package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	corev1 "github.com/rijum8906/relay/packages/pb/core/v1"
	userv1 "github.com/rijum8906/relay/packages/pb/user_service/user/v1"
)

func (s *UserService) CheckEmailExists(ctx context.Context, req *corev1.EmailRequest) (*userv1.CheckExistsResponse, error) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("check email exists request is required")
	}

	exists, err := s.DBQ.CheckUserEmailExists(ctx, req.Email)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("Failed to check user exists").WithDetail("error", err.Error())
	}

	return &userv1.CheckExistsResponse{
		Exists: exists,
	}, nil
}

func (s *UserService) CheckExists(ctx context.Context, req *corev1.IDRequest) (*userv1.CheckExistsResponse, error) {
	if req == nil {
		return nil, apperror.ErrValidation.WithMessage("check email exists request is required")
	}
	userID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, apperror.ErrValidation.WithMessage("invalid user id").WithDetail("error", err.Error())
	}

	exists, err := s.DBQ.CheckUserExists(ctx, userID)
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("Failed to check user exists").WithDetail("error", err.Error())
	}

	return &userv1.CheckExistsResponse{
		Exists: exists,
	}, nil
}
