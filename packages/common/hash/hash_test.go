package hash_test

import (
	"testing"

	"github.com/rijum8906/go-micro-service/packages/common/hash"
)

func TestHashPassword(t *testing.T) {
	hashS := hash.NewService(10)
	const password = "password"
	hashedPassword, err := hashS.HashPassword(password)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if hashedPassword == password {
		t.Errorf("password must be hashed")
	}

	isValid := hashS.VerifyPassword(hashedPassword, password)
	if isValid != nil {
		t.Errorf("password must be valid")
	}
}
