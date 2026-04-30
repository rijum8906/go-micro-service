package userauth

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/template"
	"go.uber.org/zap"
)

func (s *UserAuthEmailService) processPasswordReset(msg *nats.Msg) *apperror.AppError {
	var data dto.PasswordResetDTO

	if err := json.Unmarshal(msg.Data, &data); err != nil {
		s.Logger.Error("error unmarshalling password reset", zap.Error(err))
		return apperror.ErrInternal.WithMessage("error unmarshalling password reset").WithDetail("error", err.Error())
	}

	emailTemplate, err := s.TemplateManager.RenderToString(template.TemplateTypeEmailPasswordReset, data)
	if err != nil {
		s.Logger.Error("error rendering password reset template", zap.Error(err))
		return apperror.ErrThirdParty.WithMessage("error rendering password reset template").WithDetail("error", err.Error())
	}

	if appErr := s.sendEmail(msg, emailTemplate, "Password Reset", data.BaseEmailDTO); appErr != nil {
		s.Logger.Error("error sending password reset email", zap.Error(appErr))
		return appErr
	}

	return nil
}
