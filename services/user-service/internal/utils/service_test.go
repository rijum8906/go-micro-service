package utils_test

import (
	"testing"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/user/internal/utils"
)

func TestNewTokenURL(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		token   string
		baseURL string
		path    string
		wantErr bool
		errCode apperror.ErrorCode
	}{
		{
			name:    "should return error if token is empty",
			token:   "",
			path:    "reset-token",
			baseURL: "relay.com",

			wantErr: true,
			errCode: apperror.CodeValidation,
		},
		{
			name:    "should return error if path is empty",
			token:   "token",
			path:    "",
			baseURL: "",
			wantErr: true,
			errCode: apperror.CodeValidation,
		},
		{
			name:    "should return error if baseURL is invalid",
			token:   "token",
			path:    "reset-token",
			baseURL: "path",
			wantErr: true,
			errCode: apperror.CodeValidation,
		},
		{
			name:    "should return success",
			token:   "token",
			path:    "reset-token",
			baseURL: "http://relay.com",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, appErr := utils.NewTokenURL(tt.token, tt.baseURL, tt.path)
			if (appErr != nil) != tt.wantErr {
				t.Errorf("NewTokenURL() error = %v, wantErr %v", appErr, tt.wantErr)
				return
			}
		})
	}
}
