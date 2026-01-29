// Package dto contains request/response data transfer objects.
package dto

// UserMetadata holds request context information.
type UserMetadata struct {
	DeviceID  string `json:"deviceId"  binding:"required"`
	IPAddr    string `json:"ipAddr"    binding:"required,ip"`
	UserAgent string `json:"userAgent" binding:"required"`
}

// --------------------
// Auth DTOs
// --------------------

type SignInDTO struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`

	Metadata UserMetadata `json:"metadata" binding:"required"`
}

type SignUpDTO struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=64"`

	FirstName string `json:"firstName" binding:"required,min=1,max=50"`
	LastName  string `json:"lastName"  binding:"required,min=1,max=50"`

	Metadata UserMetadata `json:"metadata" binding:"required"`
}

// --------------------
// Password Reset
// --------------------

type RequestPasswordResetDTO struct {
	Email    string       `json:"email"    binding:"required,email"`
	Metadata UserMetadata `json:"metadata" binding:"required"`
}

type ResetPasswordDTO struct {
	Token       string       `json:"token"       binding:"required,uuid4"`
	NewPassword string       `json:"newPassword" binding:"required,min=8,max=64"`
	Metadata    UserMetadata `json:"metadata"    binding:"required"`
}

// --------------------
// Email Verification
// --------------------

type RequestEmailVerificationDTO struct {
	Email    string       `json:"email"    binding:"required,email"`
	Metadata UserMetadata `json:"metadata" binding:"required"`
}

type VerifyEmailDTO struct {
	Token    string       `json:"token"    binding:"required,uuid4"`
	Metadata UserMetadata `json:"metadata" binding:"required"`
}
