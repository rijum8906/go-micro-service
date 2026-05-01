package jobs_test

import (
	"testing"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/jobs"
)

func TestValidateJob(t *testing.T) {
	tests := []struct {
		name        string
		job         string
		wantMessage string
	}{
		{
			name: "valid user job",
			job:  "user.auth2.created-user.v1",
		},
		{
			name: "valid organization job",
			job:  "organization.membership.invited_user.v12",
		},
		{
			name:        "requires exactly four parts",
			job:         "user.auth.created_user",
			wantMessage: "invalid job format",
		},
		{
			name:        "rejects unknown domain",
			job:         "invalid.auth.created_user.v1",
			wantMessage: "invalid job format",
		},
		{
			name:        "rejects names starting with a number",
			job:         "user.2auth.created_user.v1",
			wantMessage: "invalid job format",
		},
		{
			name:        "rejects names starting with a hyphen",
			job:         "user.auth.-created_user.v1",
			wantMessage: "invalid job format",
		},
		{
			name:        "rejects names starting with a special character",
			job:         "user.@auth.created_user.v1",
			wantMessage: "invalid job format",
		},
		{
			name:        "rejects uppercase letters",
			job:         "user.auth.Created_user.v1",
			wantMessage: "invalid job format",
		},
		{
			name:        "rejects invalid version prefix",
			job:         "user.auth.created_user.version1",
			wantMessage: "invalid version format",
		},
		{
			name:        "rejects invalid version suffix",
			job:         "user.auth.created_user.v1.1",
			wantMessage: "invalid job format",
		},
		{
			name:        "rejects blank job",
			job:         "",
			wantMessage: "invalid job format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := jobs.ValidateJob(tt.job)

			if tt.wantMessage == "" {
				if err != nil {
					t.Fatalf("ValidateJob(%q) returned error: %v", tt.job, err)
				}
				return
			}

			if err == nil {
				t.Fatalf("ValidateJob(%q) returned nil error", tt.job)
			}

			appErr, ok := err.(*apperror.AppError)
			if !ok {
				t.Fatalf("ValidateJob(%q) error type = %T, want *apperror.AppError", tt.job, err)
			}

			if appErr.Code != apperror.CodeInternal {
				t.Fatalf("ValidateJob(%q) code = %q, want %q", tt.job, appErr.Code, apperror.CodeInternal)
			}

			if appErr.Message != tt.wantMessage {
				t.Fatalf("ValidateJob(%q) message = %q, want %q", tt.job, appErr.Message, tt.wantMessage)
			}
		})
	}
}
