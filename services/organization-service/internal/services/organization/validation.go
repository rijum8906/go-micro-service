package organization

import (
	"github.com/google/uuid"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/token"
	organizationv1 "github.com/rijum8906/relay/packages/pb/organization_service/organization/v1"
	"github.com/rijum8906/relay/services/organization-service/internal/utils"
)

func validateCreateOrganizationRequest(req *organizationv1.CreateOrganizationRequest) error {
	if err := uuid.Validate(req.CreatedBy); err != nil {
		return apperror.ErrValidation.WithMessage("invalid user id")
	}
	if err := uuid.Validate(req.CreatedBy); err != nil {
		return apperror.ErrValidation.WithMessage("invalid create_by id")
	}
	if !utils.ValidateSlug(req.Slug) {
		return apperror.ErrValidation.WithMessage("invalid slug")
	}

	return nil
}

func validateChangeOwnershipRequst(req *organizationv1.ChangeOrganizationOwnershipRequest) error {
	if req == nil {
		return apperror.ErrValidation.WithMessage("request cannot be nil")
	}
	if err := uuid.Validate(req.NewOwnerId); err != nil {
		return apperror.ErrValidation.WithMessage("invalid new owner id")
	}
	if err := uuid.Validate(req.OrganizationId); err != nil {
		return apperror.ErrValidation.WithMessage("invalid organization id")
	}
	if !token.ValidateTokenScope(req.TokenScope) {
		return apperror.ErrValidation.WithMessage("invalid token scope")
	}

	return nil
}
