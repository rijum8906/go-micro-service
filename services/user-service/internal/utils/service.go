// Package utils
package utils

import (
	"net/url"
	pathPkg "path"

	"github.com/rijum8906/relay/packages/core/apperror"
)

func NewTokenURL(token, baseURL, path string) (string, *apperror.AppError) {
	// Basic presence checks
	if token == "" || baseURL == "" || path == "" {
		return "", apperror.ErrValidation.WithMessage("token, base url, and path are required")
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return "", apperror.ErrInternal.WithMessage("failed to parse url").WithDetail("error", err.Error())
	}

	if base.Scheme == "" || base.Host == "" {
		return "", apperror.ErrValidation.
			WithMessage("base url must be an absolute URL (e.g., https://example.com)").
			WithDetail("provided", baseURL)
	}

	// Safely join path
	base.Path = pathPkg.Join(base.Path, path)

	// Encode query
	q := base.Query()
	q.Set("token", token)
	base.RawQuery = q.Encode()

	return base.String(), nil
}
