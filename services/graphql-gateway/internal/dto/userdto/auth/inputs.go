// Package userdto
package userdto

import (
	coreconstants "github.com/rijum8906/relay/packages/core/constants"
	"github.com/rijum8906/relay/services/graphql-gateway/internal/dto/coredto"
)

type LoginInput struct {
	Email    string              `json:"email" validate:"required,email"`
	Password string              `json:"password" validate:"required,min=8,max=50"`
	Meta     coredto.RequestMeta `json:"meta" validate:"required"`
}

type RegisterInput struct {
	Email     string              `json:"email" validate:"required,email"`
	Password  string              `json:"password" validate:"required,min=8,max=50"`
	FirstName string              `json:"firstName" validate:"required"`
	LastName  string              `json:"lastName" validate:"required"`
	Meta      coredto.RequestMeta `json:"meta" validate:"required"`
}

type LogoutInput struct {
	Meta coredto.RequestMeta `json:"meta" validate:"required"`
}

type GenerateScopedTokenInput struct {
	Scope      string                   `json:"scope" validate:"required"`
	AuthMethod coreconstants.AuthMethod `json:"authMethod" validate:"required"`
	AuthValue  string                   `json:"authValue" validate:"required"`
	Meta       coredto.RequestMeta      `json:"meta" validate:"required"`
}

type UpdateProfileAvatarUrlInput struct {
	ProfileID string `json:"profileId" validate:"required"`
	AvatarURL string `json:"avatarUrl" validate:"required,url"`
}

type UpdateProfileNameInput struct {
	ProfileID string `json:"profileId" validate:"required"`
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
}

type ChangePasswordInput struct {
	Token       string              `json:"token" validate:"required"`
	NewPassword string              `json:"newPassword" validate:"required,min=8,max=50"`
	Meta        coredto.RequestMeta `json:"meta" validate:"required"`
}

type RequestPasswordResetInput struct {
	Email string              `json:"email" validate:"required,email"`
	Meta  coredto.RequestMeta `json:"meta" validate:"required"`
}

type ResetPasswordInput struct {
	Token       string              `json:"token" validate:"required"`
	NewPassword string              `json:"newPassword" validate:"required,min=8,max=50"`
	Meta        coredto.RequestMeta `json:"meta" validate:"required"`
}

type RequestEmailVerificationInput struct {
	Email string              `json:"email" validate:"required,email"`
	Meta  coredto.RequestMeta `json:"meta" validate:"required"`
}

type VerifyEmailInput struct {
	Token string              `json:"token" validate:"required"`
	Meta  coredto.RequestMeta `json:"meta" validate:"required"`
}
