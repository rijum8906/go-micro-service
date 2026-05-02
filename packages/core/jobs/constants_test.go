package jobs_test

import (
	"testing"

	"github.com/rijum8906/relay/packages/core/jobs"
)

func TestGetDomainWildcard(t *testing.T) {
	got := jobs.GetDomainWildcard(jobs.JobUserRequestedEmailVerification)
	if got != "user.*.*.v1" {
		t.Errorf("GetDomainWildcard() = %v, want %v", got, "user.*.*.v1")
	}
}

func TestGetDomainWildcardWithoutVersion(t *testing.T) {
	got := jobs.GetDomainWildcardWithoutVersion(jobs.JobUserRequestedPasswordReset)
	if got != "user.*.*.*" {
		t.Errorf("GetDomainWildcard() = %v, want %v", got, "user.*.*.*")
	}
}

func TestGetSubdomainWildcard(t *testing.T) {
	got := jobs.GetSubdomainWildcard(jobs.JobUserRequestedEmailVerification)
	if got != "user.auth.*.v1" {
		t.Errorf("GetSubdomainWildcard() = %v, want %v", got, "user.auth.*.v1")
	}
}

func TestGetSubdomainWildcardWithoutVersion(t *testing.T) {
	got := jobs.GetSubdomainWildcardWithoutVersion(jobs.JobUserRequestedPasswordReset)
	if got != "user.auth.*.*" {
		t.Errorf("GetSubdomainWildcard() = %v, want %v", got, "user.auth.*.*")
	}
}
