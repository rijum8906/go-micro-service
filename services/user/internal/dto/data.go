// Package dto
package dto

type Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Register struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type UpdateProfileName struct {
	ProfileID string `json:"profile_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type UpdateProfileAvatarUrl struct {
	ProfileID string `json:"profile_id"`
	AvatarURL string `json:"avatar_url"`
}

type GenerateScopedToken struct {
	Scope string `json:"scope"`
}

type ChangePassword struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}
