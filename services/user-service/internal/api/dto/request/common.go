// Package request contains request data transfer objects for the auth service.
package request

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
	UserID       pgtype.UUID `json:"userId" binding:"required"`
	RefreshToken string      `json:"refreshToken" binding:"required"`
}

type Authorization struct {
	Type  string `json:"type"  binding:"required"`
	Value string `json:"value" binding:"required"`
}

type Request struct {
	Metadata Metadata `json:"metadata" binding:"required"`
}
