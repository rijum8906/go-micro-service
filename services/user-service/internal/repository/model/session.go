package model

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        uuid.UUID `gorm:"primary_key;type:uuid;default:gen_random_uuid();"`
	AccountID uuid.UUID `gorm:"uniqueIndex; not null"`

	UserAgent        string `gorm:"not null"`
	IPAddress        string `gorm:"not null"`
	Location         *string
	DeviceID         string `gorm:"not null"`
	RefreshTokenHash string `gorm:"not null; uniqueIndex"`

	ExpiresAt time.Time `gorm:"not null"`
	IsRevoked bool      `gorm:"not null; default: false"`
	LastUsed  time.Time `gorm:"not null; default: current_timestamp"`

	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
