package utils

import "regexp"

func ValidateSlug(slug string) bool {
	re, _ := regexp.Compile(`^[a-z0-9-]+$`)
	return re.MatchString(slug)
}
