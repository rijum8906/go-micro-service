// Package gorm
package gorm

import (
	"context"

	"github.com/rijum8906/go-micro-service/packages/common/hash"
	"github.com/rijum8906/go-micro-service/services/user-service/internal/repository/model"
	"gorm.io/gorm"
)

type Repository interface{}

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
	hashedPassword, err := r.hashService.HashPassword(dto.PasswordHash)
	if err != nil {
		return nil, err
	}

	var account model.Account
	account.Email = dto.Email
	account.PasswordHash = hashedPassword

	return &account, r.db.WithContext(ctx).Create(&account).Error
}

func (r *repository) UpdateAccount(ctx context.Context, accountID string, dto AccountDTO) (*model.Account, error) {
	var account model.Account
	return &account, r.db.WithContext(ctx).Where("id = ?", accountID).Updates(&account).Error
}

func (r *repository) GetAccountByEmail(ctx context.Context, email string) (*model.Account, error) {
	var account model.Account
	return &account, r.db.WithContext(ctx).Where("email = ?", email).First(&account).Error
}

func (r *repository) GetAccount(ctx context.Context, accountID string) (*model.Account, error) {
	var account model.Account
	return &account, r.db.WithContext(ctx).Where("id = ?", accountID).First(&account).Error
}

// Profile
func (r *repository) CreateProfile(ctx context.Context, dto ProfileDTO) (*model.Profile, error) {
	var profile model.Profile
	profile.AccountID = dto.AccountID
	profile.FirstName = dto.FirstName
	profile.LastName = dto.LastName
	profile.DisplayName = dto.DisplayName
	return &profile, r.db.WithContext(ctx).Create(&profile).Error
}

func (r *repository) UpdateProfile(ctx context.Context, accountID string, dto ProfileDTO) (*model.Profile, error) {
	var profile model.Profile
	return &profile, r.db.WithContext(ctx).Where("account_id = ?", accountID).Updates(&profile).Error
}

func (r *repository) GetProfile(ctx context.Context, accountID string) (*model.Profile, error) {
	var profile model.Profile
	return &profile, r.db.WithContext(ctx).Where("account_id = ?", accountID).First(&profile).Error
}

func (r *repository) GetProfileByAccountID(ctx context.Context, accountID string) (*model.Profile, error) {
	var profile model.Profile
	return &profile, r.db.WithContext(ctx).Where("account_id = ?", accountID).First(&profile).Error
}

// OAuth
func (r *repository) CreateOAuth(ctx context.Context, dto OAuthDTO) (*model.OAuth, error) {
	var oauth model.OAuth
	oauth.AccountID = dto.AccountID
	oauth.Provider = dto.Provider
	oauth.Subject = dto.Subject
	oauth.Token = dto.Token
	return &oauth, r.db.WithContext(ctx).Create(&oauth).Error
}

func (r *repository) GetOAuth(ctx context.Context, accountID string) (*model.OAuth, error) {
	var oauth model.OAuth
	return &oauth, r.db.WithContext(ctx).Where("account_id = ?", accountID).First(&oauth).Error
}

func (r *repository) GetOAuthByAccountID(ctx context.Context, accountID string) (*model.OAuth, error) {
	var oauth model.OAuth
	return &oauth, r.db.WithContext(ctx).Where("account_id = ?", accountID).First(&oauth).Error
}

// AccountSecurity
func (r *repository) CreateAccountSecurity(ctx context.Context, dto AccountSecurityDTO) (*model.AccountSecurity, error) {
	var accountSecurity model.AccountSecurity
	accountSecurity.AccountID = dto.AccountID
	return &accountSecurity, r.db.WithContext(ctx).Create(&accountSecurity).Error
}

func (r *repository) UpdateAccountSecurity(ctx context.Context, accountID string, dto AccountSecurityDTO) (*model.AccountSecurity, error) {
	var accountSecurity model.AccountSecurity
	return &accountSecurity, r.db.WithContext(ctx).Where("account_id = ?", accountID).Updates(&accountSecurity).Error
}

func (r *repository) GetAccountSecurity(ctx context.Context, accountID string) (*model.AccountSecurity, error) {
	var accountSecurity model.AccountSecurity
	return &accountSecurity, r.db.WithContext(ctx).Where("account_id = ?", accountID).First(&accountSecurity).Error
}

func (r *repository) GetAccountSecurityByAccountID(ctx context.Context, accountID string) (*model.AccountSecurity, error) {
	var accountSecurity model.AccountSecurity
	return &accountSecurity, r.db.WithContext(ctx).Where("account_id = ?", accountID).First(&accountSecurity).Error
}
