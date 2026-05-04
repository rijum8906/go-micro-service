package utils

import (
	"github.com/rijum8906/relay/packages/core/mailer"
)

func ParseMailEnvelop(config *mailer.Config, to string) mailer.Envelope {
	return mailer.Envelope{
		From: config.FromEmail,
		To:   []string{to},
		CC:   []string{},
		BCC:  []string{},
	}
}
