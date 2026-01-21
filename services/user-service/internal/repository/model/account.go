// Package model
package model

import (
	"time"

	"github.com/google/uuid"
)

type Account struct {
	ID uuid.UUID `gorm:"primary_key;type:uuid;default:gen_random_uuid();"`

	Email        string `gorm:"uniqueIndex; not null"`
	PasswordHash string `gorm:"not null; type:varchar(255)"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type Profile struct {
	ID        uuid.UUID `gorm:"primary_key;type:uuid;default:gen_random_uuid();"`
	AccountID uuid.UUID `gorm:"uniqueIndex; not null"`

	FirstName   string `gorm:"not null"`
	LastName    string `gorm:"not null"`
	DisplayName string `gorm:"not null"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type AccountSecurity struct {
	ID        uuid.UUID `gorm:"primary_key;type:uuid;default:gen_random_uuid();"`
	AccountID uuid.UUID `gorm:"uniqueIndex; not null"`

	IsEmailVerified    bool      `gorm:"not null; default: false"`
	EmailVerifiedAt    time.Time `gorm:"default: null"`
	TwoFactorEnabled   bool      `gorm:"not null; default: false"`
	TwoFactorEnabledAt time.Time `gorm:"default: null"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

type OAuth struct {
	ID        uuid.UUID `gorm:"primary_key;type:uuid;default:gen_random_uuid();"`
	AccountID uuid.UUID `gorm:"uniqueIndex; not null"`

	Provider string `gorm:"not null"`
	Subject  string `gorm:"not null"`
	Token    string `gorm:"not null"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
