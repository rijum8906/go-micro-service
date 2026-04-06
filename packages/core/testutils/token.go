package testutils

import (
	"fmt"

	modelsv1 "github.com/rijum8906/relay/packages/pb/user_service/models/v1"
)

func VerifyToken(t1, t2 *modelsv1.Token) error {
	if t1.GetValue() != t2.GetValue() {
		return fmt.Errorf("token verification failed: invalid token string: expected %s got %s", t2.Value, t1.Value)
	}

	if t1.ExpiresAt.AsTime() != t2.ExpiresAt.AsTime() {
		return fmt.Errorf("token verification failed: invalid ExpiresAt : expected %v got %v", t2.ExpiresAt.AsTime(), t1.ExpiresAt.AsTime())
	}

	return nil
}

func VerifyTokenExpiry(t1, t2 *modelsv1.Token) error {
	if t1.ExpiresAt.AsTime().Unix() < t2.ExpiresAt.AsTime().Unix() {
		return fmt.Errorf("token verification failed: invalid ExpiresAt : expected %v got %v", t2.ExpiresAt.AsTime(), t1.ExpiresAt.AsTime())
	}
	return nil
}
