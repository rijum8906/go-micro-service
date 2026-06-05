package userauth

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/template"
)

func (s *UserAuthEmailService) processEmailVerification(msg *nats.Msg) *apperror.AppError {
	var data dto.EmailVerificationDTO

	if err := json.Unmarshal(msg.Data, &data); err != nil {
		return apperror.ErrInternal.WithMessage("error unmarshalling email verification").WithDetail("error", err.Error())
	}

	emailTemplate, err := s.TemplateManager.RenderToString(template.TemplateTypeEmailVerification, data)
	if err != nil {
		return apperror.ErrThirdParty.WithMessage("error rendering email verification template").WithDetail("error", err.Error())
	}

	if appErr := s.sendEmail(msg, emailTemplate, "Email Verification", data.BaseEmailDTO); appErr != nil {
		return appErr
	}

	return nil
}
