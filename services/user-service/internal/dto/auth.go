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
