package request

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

type SignoutRequest struct {
	Metadata Metadata `json:"metadata" binding:"required"`
}

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
