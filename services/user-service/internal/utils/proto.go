package utils

import (
	commonv1 "github.com/rijum8906/relay/packages/pb/common/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func NewID(id string) *commonv1.Id {
	return &commonv1.Id{
		Value: id,
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
