package utils

import (
	commonv1 "github.com/rijum8906/relay/packages/pb/common/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewID(value string) *commonv1.UUID {
	if value == "" {
		return nil
	}
	return &commonv1.UUID{
		Value: value,
	}
}

func NewURL(value string) *commonv1.Url {
	if value == "" {
		return nil
	}
	return &commonv1.Url{
		Value: value,
	}
}

func NewIPAddr(value string) *commonv1.IPAddr {
	if value == "" {
		return nil
	}
	return &commonv1.IPAddr{
		Value: value,
	}
}

func NewIPV6Addr(value string) *commonv1.IPV6Addr {
	if value == "" {
		return nil
	}
	return &commonv1.IPV6Addr{
		Value: value,
	}
}

func NewEmail(value string) *commonv1.Email {
	if value == "" {
		return nil
	}
	return &commonv1.Email{
		Value: value,
	}
}

func NewPassword(value string) *commonv1.Password {
	if value == "" {
		return nil
	}
	return &commonv1.Password{
		Value: value,
	}
}

func NewName(value string) *commonv1.Name {
	if value == "" {
		return nil
	}
	return &commonv1.Name{
		Value: value,
	}
}

func NewToken(value string, expiresInSec int64) *commonv1.Token {
	if value == "" {
		return nil
	}
	return &commonv1.Token{
		Value: value,
		ExpiresAt: &timestamppb.Timestamp{
			Seconds: expiresInSec,
		},
	}
}
