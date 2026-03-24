package utils

import "github.com/rijum8906/relay/services/user/internal/dto"

func ParseTokens(accessToken string, refreshToken string) *dto.Tokens {
	return &dto.Tokens{
		AccessToken:  dto.TokenResult{Value: accessToken},
		RefreshToken: dto.TokenResult{Value: refreshToken},
	}
}
