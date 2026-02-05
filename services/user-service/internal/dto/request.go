// Package dto contains data transfer objects for the auth service.
package dto

import "github.com/jackc/pgx/v5/pgtype"

const (
	PassAuthzType = "password_authorization"
	MFAAuthzType  = "mfa_authorization"
)

type Metadata struct {
	DeviceID string `json:"deviceId"  binding:"required"`
}

type RequestMetadata struct {
	UserAgent string `json:"userAgent" binding:"required"`
	IPAddr    string `json:"ipAddr"    binding:"required,ip"`
	DeviceID  string `json:"deviceId"  binding:"required"`
}

type AuthzMetadata struct {
	UserID pgtype.UUID `json:"userId" binding:"required"`
}

type Authorization struct {
	Type  string `json:"type"  binding:"required"`
	Value string `json:"value" binding:"required"`
}

type SigninRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`

	Metadata Metadata `json:"metadata" binding:"required"`
}

type SignupRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`

	FirstName string `json:"firstName" binding:"required,min=1,max=20"`
	LastName  string `json:"lastName"  binding:"required,min=1,max=20"`

	Metadata Metadata `json:"metadata" binding:"required"`
}

type RequestPasswordResetRequest struct {
	Email    string   `json:"email"    binding:"required,email"`
	Metadata Metadata `json:"metadata" binding:"required"`
}

type ChangePasswordRequest struct {
	Token       string   `json:"token"         binding:"required,uuid4"`
	OldPassword string   `json:"oldPasswprd"   binding:"required,min=8,max=64"`
	NewPassword string   `json:"newPasswprd"   binding:"required,min=8,max=64"`
	Metadata    Metadata `json:"metadata"      binding:"required"`
}

type ResetPasswordRequest struct {
	Token       string   `json:"token"       binding:"required,uuid4"`
	NewPassword string   `json:"newPassword" binding:"required,min=8,max=64"`
	Metadata    Metadata `json:"metadata"    binding:"required"`
}

type RequestEmailVerificationRequest struct {
	Email    string   `json:"email"    binding:"required,email"`
	Metadata Metadata `json:"metadata" binding:"required"`
}

type VerifyEmailRequest struct {
	Token    string   `json:"token"    binding:"required,uuid4"`
	Metadata Metadata `json:"metadata" binding:"required"`
}

type GenerateScopedTokenRequest struct {
	Scope         string        `json:"scope"    binding:"required"`
	Authorization Authorization `json:"authorization" binding:"required"`
	Metadata      Metadata      `json:"metadata" binding:"required"`
}
