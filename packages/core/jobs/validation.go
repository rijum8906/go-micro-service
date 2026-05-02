package jobs

import (
	"regexp"
	"slices"
	"strings"

	"github.com/rijum8906/relay/packages/core/apperror"
)

func ValidateJob(job string) error {
	parts := strings.Split(job, ".")

	// A job segment must start with a lowercase letter and may then include
	validationRegex := regexp.MustCompile(`^[a-z][a-z0-9_-]*$`)

	if len(parts) != 4 {
		return apperror.ErrInternal.WithMessage("invalid job format")
	}

	if !slices.Contains(Domains, Domain(parts[0])) {
		return apperror.ErrInternal.WithMessage("invalid job format")
	}

	for _, part := range parts[:3] {
		if !validationRegex.MatchString(part) {
			return apperror.ErrInternal.WithMessage("invalid job format")
		}
	}

	if !regexp.MustCompile(`^v[0-9]+$`).MatchString(parts[3]) {
		return apperror.ErrInternal.WithMessage("invalid version format")
	}

	return nil
}
