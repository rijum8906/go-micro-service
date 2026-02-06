package dto

import db "github.com/rijum8906/go-micro-service/services/user-service/internal/db/generated"

// BaseSuccessResponse represents a consistent API response envelope.
type BaseSuccessResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
}

// BaseErrorResponse represents a consistent API response envelope.
type BaseErrorResponse struct {
	Success bool                `json:"success"`
	Message string              `json:"message,omitempty"`
	Errors  []BaseResponseError `json:"errors,omitempty"`
}

// BaseResponseError represents a single validation or domain error.
type BaseResponseError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
}

type Token struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type AuthResponse struct {
	Account  *db.Account   `json:"account"`
	Profiles *[]db.Profile `json:"profiles"`
	Token    *Token        `json:"token"`
}
