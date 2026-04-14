package utils

import (
	"net/mail"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/mailer"
)

func ParseMailEnvelop(config *mailer.Config, to string) (mailer.Envelope, *apperror.AppError) {
	fromEmail, err := mail.ParseAddress(config.FromEmail)
	if err != nil {
		return mailer.Envelope{}, apperror.ErrInternal.WithMessage("failed to parse from email").WithDetail("error", err.Error())
	}

	tos := make([]*mail.Address, 1)
	toAddr, err := mail.ParseAddress(to)
	if err != nil {
		return mailer.Envelope{}, apperror.ErrInternal.WithMessage("failed to parse to email").WithDetail("error", err.Error())
	}
	tos = append(tos, toAddr)

	return mailer.Envelope{
		From: fromEmail,
		To:   tos,
		CC:   []*mail.Address{},
		BCC:  []*mail.Address{},
	}, nil
}
