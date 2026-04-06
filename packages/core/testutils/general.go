// Package testutils
package testutils

import (
	"math/rand"
	"strings"
)

func GenerateRandomString(length int) string {
	set := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var result strings.Builder
	for range length {
		result.WriteString(string(set[rand.Intn(len(set))]))
	}
	return result.String()
}

func GenerateRandomEmail() string {
	return GenerateRandomString(10) + "@example.com"
}
