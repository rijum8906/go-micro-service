package template_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/template"
)

func mustCreateTemplateManager(t *testing.T) template.TemplateManager {
	t.Helper()

	manager, err := template.NewTemplateManager(filepath.Join("..", "..", "templates"))
	if err != nil {
		t.Fatalf("failed to create template manager: %v", err)
	}

	return manager
}

func TestNewTemplateManager_Failure(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		templatesDir string
	}{
		{
			name:         "empty directory",
			templatesDir: "",
		},
		{
			name:         "missing directory",
			templatesDir: filepath.Join("..", "..", "missing-templates"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := template.NewTemplateManager(tc.templatesDir)
			if err == nil {
				t.Fatal("expected constructor error, got nil")
			}
		})
	}
}

func TestTemplateManager_ReloadTemplates(t *testing.T) {
	t.Parallel()

	tm := mustCreateTemplateManager(t)
	if err := tm.ReloadTemplates(); err != nil {
		t.Fatalf("reload returned error: %v", err)
	}
}

func TestTemplateManager_ValidateData(t *testing.T) {
	t.Parallel()

	tm := mustCreateTemplateManager(t)

	testCases := []struct {
		name     string
		data     any
		wantErr  bool
		wantCode apperror.ErrorCode
	}{
		{
			name: "valid password reset dto",
			data: template.PasswordResetDTO{
				ClientName:  "John Doe",
				ClientEmail: "john@example.com",
				ResetToken:  "token123",
				Validity:    "15 minutes",
			},
		},
		{
			name:     "nil data",
			data:     nil,
			wantErr:  true,
			wantCode: apperror.CodeValidation,
		},
		{
			name: "invalid email verification dto",
			data: template.EmailVerificationDTO{
				ClientName:        "John Doe",
				ClientEmail:       "invalid-email",
				VerificationToken: "",
				Validity:          "",
			},
			wantErr:  true,
			wantCode: apperror.CodeValidation,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tm.ValidateData(tc.data)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected validation error, got nil")
				}
				if err.Code != tc.wantCode {
					t.Fatalf("unexpected error code: got %s want %s", err.Code, tc.wantCode)
				}
				return
			}

			if err != nil {
				t.Fatalf("expected no validation error, got %v", err)
			}
		})
	}
}

func TestTemplateManager_RenderToString(t *testing.T) {
	t.Parallel()

	tm := mustCreateTemplateManager(t)

	testCases := []struct {
		name         string
		templateType template.TemplateType
		data         any
		wantContains []string
	}{
		{
			name:         "password reset",
			templateType: template.TemplateTypeEmailPasswordReset,
			data: template.PasswordResetDTO{
				ClientName:  "John Doe",
				ClientEmail: "john@example.com",
				ResetToken:  "token123",
				Validity:    "15 minutes",
			},
			wantContains: []string{"John Doe", "john@example.com", "token123", "15 minutes"},
		},
		{
			name:         "welcome email",
			templateType: template.TemplateTypeEmailWelcome,
			data: template.WelcomeTemplateDTO{
				ClientName:  "John Doe",
				ClientEmail: "john@example.com",
			},
			wantContains: []string{"John Doe", "john@example.com"},
		},
		{
			name:         "email verification",
			templateType: template.TemplateTypeEmailVerification,
			data: template.EmailVerificationDTO{
				ClientName:        "John Doe",
				ClientEmail:       "john@example.com",
				VerificationToken: "verify-123",
				Validity:          "15 minutes",
			},
			wantContains: []string{"John Doe", "john@example.com", "verify-123", "15 minutes"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			rendered, err := tm.RenderToString(tc.templateType, tc.data)
			if err != nil {
				t.Fatalf("render returned error: %v", err)
			}
			if strings.TrimSpace(rendered) == "" {
				t.Fatal("rendered template is empty")
			}
			for _, expected := range tc.wantContains {
				if !strings.Contains(rendered, expected) {
					t.Fatalf("rendered template does not contain %q: %s", expected, rendered)
				}
			}
		})
	}
}

func TestTemplateManager_RenderToBytes(t *testing.T) {
	t.Parallel()

	tm := mustCreateTemplateManager(t)

	data := template.PasswordResetDTO{
		ClientName:  "John Doe",
		ClientEmail: "john@example.com",
		ResetToken:  "token123",
		Validity:    "15 minutes",
	}

	rendered, err := tm.RenderToBytes(template.TemplateTypeEmailPasswordReset, data)
	if err != nil {
		t.Fatalf("render returned error: %v", err)
	}

	renderedString := string(rendered)
	if strings.TrimSpace(renderedString) == "" {
		t.Fatal("rendered template is empty")
	}

	for _, expected := range []string{"John Doe", "john@example.com", "token123", "15 minutes"} {
		if !strings.Contains(renderedString, expected) {
			t.Fatalf("rendered template does not contain %q: %s", expected, renderedString)
		}
	}
}

func TestTemplateManager_RenderFailures(t *testing.T) {
	t.Parallel()

	tm := mustCreateTemplateManager(t)

	testCases := []struct {
		name         string
		templateType template.TemplateType
		data         any
		wantCode     apperror.ErrorCode
	}{
		{
			name:         "empty template type",
			templateType: "",
			data: template.WelcomeTemplateDTO{
				ClientName:  "John Doe",
				ClientEmail: "john@example.com",
			},
			wantCode: apperror.CodeValidation,
		},
		{
			name:         "unknown template",
			templateType: template.TemplateType("missing-template"),
			data: template.WelcomeTemplateDTO{
				ClientName:  "John Doe",
				ClientEmail: "john@example.com",
			},
			wantCode: apperror.CodeInternal,
		},
		{
			name:         "invalid payload",
			templateType: template.TemplateTypeEmailVerification,
			data: template.EmailVerificationDTO{
				ClientName:  "John Doe",
				ClientEmail: "not-an-email",
			},
			wantCode: apperror.CodeValidation,
		},
	}

	for _, tc := range testCases 
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := tm.RenderToString(tc.templateType, tc.data)
			if err == nil {
				t.Fatal("expected render error, got nil")
			}
			if err.Code != tc.wantCode {
				t.Fatalf("unexpected error code: got %s want %s", err.Code, tc.wantCode)
			}
		})
	}
}