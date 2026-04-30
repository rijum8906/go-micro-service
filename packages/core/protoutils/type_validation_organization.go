package protoutils

import (
	"net/mail"

	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	org_membershipv1 "github.com/rijum8906/relay/packages/pb/organization_service/org_membership/v1"
)

func ValidateSendInvitationReq(req *org_membershipv1.SendInvitationRequest) *apperror.AppError {
	if req == nil {
		return apperror.ErrValidation.WithMessage("request body cannot be nil")
	}

	if err := uuid.Validate(req.OrganizationId); err != nil {
		return apperror.ErrValidation.WithMessage("invalid organization id")
	}

	if req.Email == "" {
		return apperror.ErrValidation.WithMessage("email cannot be empty")
	}

	// Use Go's built-in email parser
	_, err := mail.ParseAddress(req.Email)
	if err != nil {
		return apperror.ErrValidation.WithMessage("invalid email format: " + err.Error())
	}

	return nil
}
