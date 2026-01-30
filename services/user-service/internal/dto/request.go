package dto

type Metadata struct {
	DeviceID string `json:"deviceId"  binding:"required"`
}

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
