// Package utils
package utils

import (
	"fmt"
	"net/url"
	pathPkg "path"
	"strconv"
	"strings"
	"time"

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

type TOTPProtocol string

const (
	TOTPProtocolGoogle TOTPProtocol = "otpauth"
)

type TOTPType string

const (
	TOTPTypeTOTP TOTPType = "totp"
)

type TOTPAuthURIConfig struct {
	Protocol     TOTPProtocol  // e.g., "otpauth"
	Type         TOTPType      // e.g., "totp"
	Issuer       string        // e.g., "YourCompany"
	Email        string        // e.g., "user@example.com"
	Secret       string        // The Base32 encoded secret key
	Algorithm    string        // e.g., "SHA1", "SHA256"
	CodeLength   int           // e.g., 6 or 8
	CodeValidity time.Duration // e.g., 30 * time.Second
}

// GenerateTOTPAuthURI builds a fully valid provisioning URI for authenticator QR codes
func GenerateTOTPAuthURI(config TOTPAuthURIConfig) string {
	// 1. Build the path: Issuer:Email (e.g., "YourCompany:user@example.com")
	// Both the issuer and the email should be clearly identified in the path
	label := fmt.Sprintf("%s:%s", config.Issuer, config.Email)

	// 2. Set up the base URL structure
	u := &url.URL{
		Scheme: "otpauth",
		Host:   string(config.Type), // typically "totp"
		Path:   label,
	}

	// 3. Construct the query parameters safely
	query := url.Values{}
	query.Set("secret", config.Secret)
	query.Set("issuer", config.Issuer)

	// Default to SHA1 if not specified, as many apps only support SHA1
	algo := strings.ToUpper(config.Algorithm)
	if algo == "" {
		algo = "SHA1"
	}
	query.Set("algorithm", algo)

	// Default to 6 digits if not specified
	digits := config.CodeLength
	if digits == 0 {
		digits = 6
	}
	query.Set("digits", strconv.Itoa(digits))

	// Convert duration (seconds) to integer string (default to 30s)
	period := int(config.CodeValidity.Seconds())
	if period == 0 {
		period = 30
	}
	query.Set("period", strconv.Itoa(period))

	// 4. Assign the query parameters back to the URL
	u.RawQuery = query.Encode()

	// 5. Return the string representation
	// Note: Go's net/url package keeps path slashes safe, but some old apps
	// expect raw escaping. u.String() works perfectly across modern standards.
	return u.String()
}
