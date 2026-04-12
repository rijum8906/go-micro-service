package dto

// EmailVerificationJob carries the data needed to send a verification email.
type EmailVerificationJob struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	ScopedToken string `json:"scoped_token"`
	ExpiresIn   string `json:"expires_in"`
}

// PasswordResetJob carries the data needed to send a password reset email.
type PasswordResetJob struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	ScopedToken string `json:"scoped_token"`
	ExpiresIn   string `json:"expires_in"`
}
