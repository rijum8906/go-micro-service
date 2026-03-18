package utils

import (
	commonv1 "github.com/rijum8906/relay/packages/pb/common/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewID(id string) *commonv1.UUID {
	return &commonv1.UUID{
		Value: id,
	}
}

func NewURL(url string) *commonv1.Url {
	return &commonv1.Url{
		Value: url,
	}
}

func NewIPAddr(ipAddr string) *commonv1.IPAddr {
	return &commonv1.IPAddr{
		Value: ipAddr,
	}
}

func NewIPV6Addr(ipv6Addr string) *commonv1.IPV6Addr {
	return &commonv1.IPV6Addr{
		Value: ipv6Addr,
	}
}

func NewEmail(email string) *commonv1.Email {
	return &commonv1.Email{
		Value: email,
	}
}

func NewPassword(password string) *commonv1.Password {
	return &commonv1.Password{
		Value: password,
	}
}

func NewName(name string) *commonv1.Name {
	return &commonv1.Name{
		Value: name,
	}
}

func NewToken(value string, expiresInSec int64) *commonv1.Token {
	return &commonv1.Token{
		Value: value,
		ExpiresAt: &timestamppb.Timestamp{
			Seconds: expiresInSec,
		},
	}
}
