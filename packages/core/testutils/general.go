// Package testutils
package testutils

import (
	"math/rand"
	"strings"

	"github.com/rijum8906/relay/packages/core/dto"
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

func GenerateClientInfo() *dto.ClientInfo {
	return &dto.ClientInfo{
		DeviceID:   GenerateRandomString(10),
		UserAgent:  GenerateRandomString(10),
		IPAddress:  GenerateRandomString(10),
		ClientType: "Web",
		APIVersion: "v1",
		SDKVersion: "0.0.1",
		RequestID:  GenerateRandomString(10),
	}
}
