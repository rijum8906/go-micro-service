package response

import "github.com/jackc/pgx/v5/pgtype"

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type AccountResponse struct {
	ID    pgtype.UUID `json:"id"`
	Email string      `json:"email"`
}

type MyAccountRespose struct {
	ID               pgtype.UUID        `json:"id"`
	Email            string             `json:"email"`
	IsEmailVerified  bool               `json:"isEmailVerified"`
	TwoFactorEnabled bool               `json:"twoFactorEnabled"`
	CreatedAt        pgtype.Timestamptz `json:"createdAt"`
	UpdatedAt        pgtype.Timestamptz `json:"updatedAt"`
}

type ProfileResponse struct {
	ID          pgtype.UUID `json:"id"`
	FirstName   string      `json:"firstName"`
	LastName    string      `json:"lastName"`
	DisplayName pgtype.Text `json:"displayName"`
	AvatarUrl   pgtype.Text `json:"avatarUrl"`
}

type PublicProfileResponse struct {
	ID          pgtype.UUID `json:"id"`
	DisplayName pgtype.Text `json:"displayName"`
	AvatarUrl   pgtype.Text `json:"avatarUrl"`
}

type AuthResponse struct {
	Account  *AccountResponse   `json:"account"`
	Profiles []*ProfileResponse `json:"profiles"`
	Token    *TokenResponse     `json:"tokens"`
}
