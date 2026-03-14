// Package dto contains data transfer objects for the user service.
package dto

import (
	"mime/multipart"

	"github.com/jackc/pgx/v5/pgtype"
)

const (
	PassAuthzType = "password_authorization"
	MFAAuthzType  = "mfa_authorization"
)

//--------- BASE ---------

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

type Request struct {
	Metadata Metadata `json:"metadata" binding:"required"`
}

// --------- AUTH ----------

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

// --------- PASSWORDS ----------

type RequestPasswordResetRequest struct {
	Email    string   `json:"email"    binding:"required,email"`
	Metadata Metadata `json:"metadata" binding:"required"`
}

type ChangePasswordRequest struct {
	Token       string   `json:"token"         binding:"required"`
	NewPassword string   `json:"newPassword"   binding:"required,min=8,max=64"`
	Metadata    Metadata `json:"metadata"      binding:"required"`
}

type ResetPasswordRequest struct {
	Token       string   `json:"token"       binding:"required"`
	NewPassword string   `json:"newPassword" binding:"required,min=8,max=64"`
	Metadata    Metadata `json:"metadata"    binding:"required"`
}

// --------- EMAILS ----------

type RequestEmailVerificationRequest struct {
	Email    string   `json:"email"    binding:"required,email"`
	Metadata Metadata `json:"metadata" binding:"required"`
}

type VerifyEmailRequest struct {
	Token    string   `json:"token"    binding:"required,uuid4"`
	Metadata Metadata `json:"metadata" binding:"required"`
}

type ChangeEmailRequest struct {
	Token    string   `json:"token"         binding:"required"`
	NewEmail string   `json:"newEmail"      binding:"required,email"`
	Metadata Metadata `json:"metadata"      binding:"required"`
}

// -------- TOKENS --------

type GenerateScopedTokenRequest struct {
	Scope         string        `json:"scope"    binding:"required"`
	Authorization Authorization `json:"authorization" binding:"required"`
	Metadata      Metadata      `json:"metadata" binding:"required"`
}

// --------- PROFILES ----------

type UpdateProfileRequest struct {
	ProfileID   string   `form:"profileId" binding:"required,uuid4"`
	FirstName   *string  `form:"firstName" binding:"omitempty,min=1,max=20"`
	LastName    *string  `form:"lastName"  binding:"omitempty,min=1,max=20"`
	DisplayName *string  `form:"displayName"  binding:"omitempty,min=1,max=20"`
	AvatarURL   *string  `binding:"omitempty,max=255"`
	Metadata    Metadata `form:"metadata"  binding:"required"`

	Avatar *multipart.FileHeader `form:"avatar"`
}

type GetProfileRequest struct {
	ProfileID string   `form:"profileId" binding:"required,uuid4"`
	Metadata  Metadata `form:"metadata"  binding:"required"`
}

type DeleteProfileRequest struct {
	Metadata Metadata `form:"metadata"  binding:"required"`
}

type CreateProfileRequest struct {
	FirstName   string   `form:"firstName" binding:"required,min=1,max=20"`
	LastName    string   `form:"lastName"  binding:"required,min=1,max=20"`
	DisplayName *string  `form:"displayName"  binding:"omitempty,min=1,max=20"`
	AvatarURL   *string  `form:"avatarUrl"  binding:"omitempty,max=255"`
	Metadata    Metadata `form:"metadata"  binding:"required"`

	Avatar *multipart.FileHeader `form:"avatar"`
}

// --------- ACCOUNTS SECURITIES ----------
