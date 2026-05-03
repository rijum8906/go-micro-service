package testutils

import (
	"testing"

	"github.com/rijum8906/relay/packages/core/apperror"
)

type TestError struct {
	Code apperror.ErrorCode
}

func AssertError(t *testing.T, err error, expectedMessage string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
	if err.Error() != expectedMessage {
		t.Fatalf("expected error message to be %s but got %s", expectedMessage, err.Error())
	}
}
