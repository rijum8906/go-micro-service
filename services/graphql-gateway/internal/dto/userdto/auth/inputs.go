// Package userdto
package userdto

import "github.com/rijum8906/relay/services/graphql-gateway/internal/dto/coredto"

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
