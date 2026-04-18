// Package constants
package constants

type DurableName string

const (
	DurableEmailVerification DurableName = "email_verification"
	DurablePasswordReset     DurableName = "password_reset"
)

type StreamName string

const (
	StreamEmailVerification StreamName = "email_verification"
	StreamPasswordReset     StreamName = "password_reset"
)
