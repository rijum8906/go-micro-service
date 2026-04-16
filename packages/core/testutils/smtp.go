package testutils

import (
	"os"
	"strconv"

	"github.com/rijum8906/relay/packages/core/mailer"
)

func MustReturnSMTPConfig() mailer.Config {
	host := os.Getenv("TEST_SMTP_HOST")
	port := os.Getenv("TEST_SMTP_PORT")
	username := os.Getenv("TEST_SMTP_USERNAME")
	password := os.Getenv("TEST_SMTP_PASSWORD")

	portInt64, err := strconv.ParseInt(port, 10, 64)
	if err != nil {
		panic(err)
	}

	return mailer.Config{
		Host:        host,
		Port:        int(portInt64),
		Username:    username,
		Password:    password,
		UseStartTLS: true,
		Retries:     3,
	}
}
