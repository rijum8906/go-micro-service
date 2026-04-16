package utils

import (
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/mailer"
)

func ParseMailEnvelop(config *mailer.Config, to string) (mailer.Envelope, *apperror.AppError) {
	return mailer.Envelope{
		From: config.FromEmail,
		To:   []string{to},
		CC:   []string{},
		BCC:  []string{},
	}, nil
}
