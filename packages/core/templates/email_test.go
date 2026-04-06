package templates

import (
	"io/fs"
	"strings"
	"testing"
)

func TestEmailTemplateFilesExist(t *testing.T) {
	for _, name := range Names() {
		for _, file := range Files(name) {
			if _, err := fs.Stat(FS, file); err != nil {
				t.Fatalf("expected embedded template file %q to exist: %v", file, err)
			}
		}
	}
}

func TestEmailTemplateValidate(t *testing.T) {
	for _, name := range Names() {
		if err := name.Validate(); err != nil {
			t.Fatalf("expected template %q to validate: %v", name, err)
		}
	}

	if err := EmailTemplate("").Validate(); err == nil {
		t.Fatal("expected empty template to fail validation")
	}

	if err := EmailTemplate("missing-template").Validate(); err == nil {
		t.Fatal("expected unknown template to fail validation")
	}
}

func TestRenderEmail(t *testing.T) {
	rendered, err := RenderEmail(EmailTemplateVerifyEmail, map[string]string{
		"app_name":         "Relay",
		"email":            "user@example.com",
		"verification_url": "https://example.com/verify?token=abc",
	})
	if err != nil {
		t.Fatalf("expected template render to succeed: %v", err)
	}

	if !strings.Contains(rendered.Subject, "Relay") {
		t.Fatalf("expected rendered subject to include app name, got %q", rendered.Subject)
	}

	if !strings.Contains(rendered.TextBody, "https://example.com/verify?token=abc") {
		t.Fatalf("expected rendered text body to include verification url, got %q", rendered.TextBody)
	}

	if !strings.Contains(rendered.HTMLBody, "user@example.com") {
		t.Fatalf("expected rendered html body to include email, got %q", rendered.HTMLBody)
	}
}

func TestRenderEmailMissingData(t *testing.T) {
	_, err := RenderEmail(EmailTemplateVerifyEmail, map[string]string{
		"app_name": "Relay",
	})
	if err == nil {
		t.Fatal("expected render to fail when required template data is missing")
	}
}
