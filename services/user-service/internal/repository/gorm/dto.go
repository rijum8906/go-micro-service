package gorm

import (
	"time"

	"github.com/google/uuid"
)

type AccountDTO struct {
	Email         string `json:"email" binding:"required,email"`
	PlainPassword string `json:"password" binding:"required"`
}

type ProfileDTO struct {
	AccountID   uuid.UUID `json:"account_id" binding:"required"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DisplayName string    `json:"display_name"`
}

type OAuthDTO struct {
	AccountID uuid.UUID `json:"account_id" binding:"required"`
	Provider  string    `json:"provider" binding:"required"`
	Subject   string    `json:"subject" binding:"required"`
	Token     string    `json:"token" binding:"required"`
}

type AccountSecurityDTO struct {
	AccountID          uuid.UUID `json:"account_id" binding:"required"`
	IsEmailVerified    bool      `json:"is_email_verified"`
	EmailVerifiedAt    time.Time `json:"email_verified_at"`
	TwoFactorEnabled   bool      `json:"two_factor_enabled"`
	TwoFactorEnabledAt time.Time `json:"two_factor_enabled_at"`
}

type SessionDTO struct {
	AccountID        uuid.UUID `json:"account_id" binding:"required"`
	UserAgent        string    `json:"user_agent"`
	IPAddress        string    `json:"ip_address"`
	Location         *string   `json:"location"`
	DeviceID         string    `json:"device_id"`
	RefreshTokenHash string    `json:"refresh_token_hash"`
	ExpiresAt        time.Time `json:"expires_at"`
	LastUsed         time.Time `json:"last_used"`
}
