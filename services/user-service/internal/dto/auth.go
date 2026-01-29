// Package dto
package dto

type UserMetadata struct {
	DeviceID  string
	IPAddr    string
	UserAgent string
}

type SignInDTO struct {
	Email    string
	Password string

	Metadata UserMetadata
}

type SignUpDTO struct {
	Email    string
	Password string

	FirstName string
	LastName  string

	Metadata UserMetadata
}
type RequestPasswordResetDTO struct {
	Email    string
	Metadata UserMetadata
}

type ResetpasswordDTO struct {
	Token       string
	NewPassword string
	Metadata    UserMetadata
}

type RequestEmailVerificationDTO struct {
	Email    string
	Metadata UserMetadata
}

type VerifyEmailDTO struct {
	Token    string
	Metadata UserMetadata
}
