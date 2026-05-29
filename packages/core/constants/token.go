package coreconstants

type TokenType string

const (
	TokenTypeBearer    TokenType = "Bearer"
	TokenType2FA       TokenType = "2FA-Challenge"
	TokenTypeUndefined TokenType = "undefined"
)
