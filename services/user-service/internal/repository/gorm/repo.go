// Package gorm
package gorm

import (
	"context"

	"github.com/rijum8906/go-micro-service/packages/common/hash"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/repository/model"
	"gorm.io/gorm"
)

type Repository interface {
	CreateAccount(ctx context.Context, dto AccountDTO) (*model.Account, error)
	UpdateAccount(ctx context.Context, accountID string, dto AccountDTO) error
	GetAccount(ctx context.Context, accountID string) (*model.Account, error)
	GetAccountByEmail(ctx context.Context, email string) (*model.Account, error)

	CreateProfile(ctx context.Context, dto ProfileDTO) (*model.Profile, error)
	UpdateProfile(ctx context.Context, accountID string, dto ProfileDTO) error
	GetProfileByAccountID(ctx context.Context, accountID string) (*model.Profile, error)

	CreateOAuth(ctx context.Context, dto OAuthDTO) (*model.OAuth, error)
	GetOAuthByAccountID(ctx context.Context, accountID string) (*model.OAuth, error)

	CreateAccountSecurity(ctx context.Context, dto AccountSecurityDTO) (*model.AccountSecurity, error)
	UpdateAccountSecurity(ctx context.Context, accountID string, dto AccountSecurityDTO) error
	GetAccountSecurityByAccountID(ctx context.Context, accountID string) (*model.AccountSecurity, error)

	CreateSession(ctx context.Context, dto SessionDTO) (*model.Session, error)
}

type repository struct {
	db          *gorm.DB
	hashService hash.Service
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{
		db:          db,
		hashService: hash.NewService(10),
	}
}

// Account

func (r *repository) CreateAccount(ctx context.Context, dto AccountDTO) (*model.Account, error) {
	hashedPassword, err := r.hashService.HashPassword(dto.PlainPassword)
	if err != nil {
		return nil, err
	}

	account := model.Account{
		Email:        dto.Email,
		PasswordHash: hashedPassword,
	}

	return &account, r.db.WithContext(ctx).Create(&account).Error
}

func (r *repository) UpdateAccount(ctx context.Context, accountID string, dto AccountDTO) error {
	updates := map[string]any{}

	if dto.Email != "" {
		updates["email"] = dto.Email
	}

	if dto.PlainPassword != "" {
		hash, err := r.hashService.HashPassword(dto.PlainPassword)
		if err != nil {
			return err
		}
		updates["password_hash"] = hash
	}

	return r.db.WithContext(ctx).
		Model(&model.Account{}).
		Where("id = ?", accountID).
		Updates(updates).Error
}

func (r *repository) GetAccount(ctx context.Context, accountID string) (*model.Account, error) {
	var account model.Account
	return &account, r.db.WithContext(ctx).
		Where("id = ?", accountID).
		First(&account).Error
}

func (r *repository) GetAccountByEmail(ctx context.Context, email string) (*model.Account, error) {
	var account model.Account
	return &account, r.db.WithContext(ctx).
		Where("email = ?", email).
		First(&account).Error
}

// Profile

func (r *repository) CreateProfile(ctx context.Context, dto ProfileDTO) (*model.Profile, error) {
	profile := model.Profile{
		AccountID:   dto.AccountID,
		FirstName:   dto.FirstName,
		LastName:    dto.LastName,
		DisplayName: dto.DisplayName,
	}
	return &profile, r.db.WithContext(ctx).Create(&profile).Error
}

func (r *repository) UpdateProfile(ctx context.Context, accountID string, dto ProfileDTO) error {
	return r.db.WithContext(ctx).
		Model(&model.Profile{}).
		Where("account_id = ?", accountID).
		Updates(map[string]any{
			"first_name":   dto.FirstName,
			"last_name":    dto.LastName,
			"display_name": dto.DisplayName,
		}).Error
}

func (r *repository) GetProfileByAccountID(ctx context.Context, accountID string) (*model.Profile, error) {
	var profile model.Profile
	return &profile, r.db.WithContext(ctx).
		Where("account_id = ?", accountID).
		First(&profile).Error
}

// OAuth

func (r *repository) CreateOAuth(ctx context.Context, dto OAuthDTO) (*model.OAuth, error) {
	oauth := model.OAuth{
		AccountID: dto.AccountID,
		Provider:  dto.Provider,
		Subject:   dto.Subject,
		Token:     dto.Token,
	}
	return &oauth, r.db.WithContext(ctx).Create(&oauth).Error
}

func (r *repository) GetOAuthByAccountID(ctx context.Context, accountID string) (*model.OAuth, error) {
	var oauth model.OAuth
	return &oauth, r.db.WithContext(ctx).
		Where("account_id = ?", accountID).
		First(&oauth).Error
}

// AccountSecurity

func (r *repository) CreateAccountSecurity(ctx context.Context, dto AccountSecurityDTO) (*model.AccountSecurity, error) {
	sec := model.AccountSecurity{
		AccountID:          dto.AccountID,
		IsEmailVerified:    dto.IsEmailVerified,
		EmailVerifiedAt:    dto.EmailVerifiedAt,
		TwoFactorEnabled:   dto.TwoFactorEnabled,
		TwoFactorEnabledAt: dto.TwoFactorEnabledAt,
	}
	return &sec, r.db.WithContext(ctx).Create(&sec).Error
}

func (r *repository) UpdateAccountSecurity(ctx context.Context, accountID string, dto AccountSecurityDTO) error {
	return r.db.WithContext(ctx).
		Model(&model.AccountSecurity{}).
		Where("account_id = ?", accountID).
		Updates(dto).Error
}

func (r *repository) GetAccountSecurityByAccountID(ctx context.Context, accountID string) (*model.AccountSecurity, error) {
	var sec model.AccountSecurity
	return &sec, r.db.WithContext(ctx).
		Where("account_id = ?", accountID).
		First(&sec).Error
}

// Session

func (r *repository) CreateSession(ctx context.Context, dto SessionDTO) (*model.Session, error) {
	session := model.Session{
		AccountID:        dto.AccountID,
		RefreshTokenHash: dto.RefreshTokenHash,
		UserAgent:        dto.UserAgent,
		IPAddress:        dto.IPAddress,
		Location:         dto.Location,
		DeviceID:         dto.DeviceID,
		ExpiresAt:        dto.ExpiresAt,
		IsRevoked:        false,
		LastUsed:         dto.LastUsed,
	}
	return &session, r.db.WithContext(ctx).Create(&session).Error
}
