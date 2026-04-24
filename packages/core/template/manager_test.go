package template_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/template"
)

func sampleCompanyInfo() *dto.CompanyInfo {
	return &dto.CompanyInfo{
		Name:       "Relay Labs",
		Emails:     []string{"support@relay.dev", "hello@relay.dev"},
		Addresses:  []string{"123 Relay Street, Dhaka 1212", "42 Signal Avenue, Tokyo 150-0001"},
		WebsiteURL: "https://relay.dev",
		SocialLinks: []dto.SocialLink{
			{Label: "LinkedIn", URL: "https://linkedin.com/company/relay"},
			{Label: "X", URL: "https://x.com/relay"},
		},
	}
}

func mustCreateTemplateManager(t *testing.T) template.TemplateManager {
	t.Helper()

	manager, err := template.NewTemplateManager(filepath.Join("..", "..", "templates"))
	if err != nil {
		t.Fatalf("failed to create template manager: %v", err)
	}

	return manager
}

func mustCreateTemplateManagerWithCompanyInfo(t *testing.T) template.TemplateManager {
	t.Helper()

	manager, err := template.NewTemplateManagerWithCompanyInfo(filepath.Join("..", "..", "templates"), sampleCompanyInfo())
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

func TestNewTemplateManager_Success(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		_, err := template.NewTemplateManager(filepath.Join("..", "..", "templates"))
		if err != nil {
			t.Fatalf("failed to create template manager: %v", err)
		}
	})
}

func TestTemplateManager_ReloadTemplates(t *testing.T) {
	t.Parallel()

	tm := mustCreateTemplateManagerWithCompanyInfo(t)
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
			data: dto.PasswordResetDTO{
				ClientName:  "John Doe",
				ClientEmail: "john@example.com",
				ResetURL:    "https://relay.dev/reset-password?token=token123",
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
			data: dto.EmailVerificationDTO{
				ClientName:      "John Doe",
				ClientEmail:     "invalid-email",
				VerificationURL: "",
				Validity:        "",
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

	tm := mustCreateTemplateManagerWithCompanyInfo(t)

	testCases := []struct {
		name         string
		templateType template.TemplateType
		data         any
		wantContains []string
	}{
		{
			name:         "password reset",
			templateType: template.TemplateTypeEmailPasswordReset,
			data: dto.PasswordResetDTO{
				ClientName:  "John Doe",
				ClientEmail: "john@example.com",
				ResetURL:    "https://relay.dev/reset-password?token=token123",
				Validity:    "15 minutes",
			},
			wantContains: []string{
				"John Doe",
				"john@example.com",
				"token123",
				"15 minutes",
				"support@relay.dev",
				"123 Relay Street, Dhaka 1212",
				"https://linkedin.com/company/relay",
			},
		},
		{
			name:         "welcome email",
			templateType: template.TemplateTypeEmailWelcome,
			data: dto.WelcomeTemplateDTO{
				ClientName:  "John Doe",
				ClientEmail: "john@example.com",
			},
			wantContains: []string{
				"John Doe",
				"john@example.com",
				"https://relay.dev",
				"hello@relay.dev",
			},
		},
		{
			name:         "email verification",
			templateType: template.TemplateTypeEmailVerification,
			data: dto.EmailVerificationDTO{
				ClientName:      "John Doe",
				ClientEmail:     "john@example.com",
				VerificationURL: "verify-123",
				Validity:        "15 minutes",
			},
			wantContains: []string{
				"John Doe",
				"john@example.com",
				"verify-123",
				"15 minutes",
				"https://x.com/relay",
				"Relay Labs",
			},
		},
	}

	for _, tc := range testCases {
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

	tm := mustCreateTemplateManagerWithCompanyInfo(t)

	data := dto.PasswordResetDTO{
		ClientName:  "John Doe",
		ClientEmail: "john@example.com",
		ResetURL:    "https://relay.dev/reset-password?token=token123",
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
		if !strings.Contains(renderedString, "support@relay.dev") {
			t.Fatalf("rendered template does not contain %q: %s", "support@relay.dev", renderedString)
		}
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
			data: dto.WelcomeTemplateDTO{
				ClientName:  "John Doe",
				ClientEmail: "john@example.com",
			},
			wantCode: apperror.CodeValidation,
		},
		{
			name:         "unknown template",
			templateType: template.TemplateType("missing-template"),
			data: dto.WelcomeTemplateDTO{
				ClientName:  "John Doe",
				ClientEmail: "john@example.com",
			},
			wantCode: apperror.CodeInternal,
		},
		{
			name:         "invalid payload",
			templateType: template.TemplateTypeEmailVerification,
			data: dto.EmailVerificationDTO{
				ClientName:  "John Doe",
				ClientEmail: "not-an-email",
			},
			wantCode: apperror.CodeValidation,
		},
	}

	for _, tc := range testCases {
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

func TestTemplateManager_RenderToString_GenericWrapperData(t *testing.T) {
	t.Parallel()

	tm := mustCreateTemplateManagerWithCompanyInfo(t)

	type wrappedWelcomeDTO struct {
		dto.WelcomeTemplateDTO
		Extra string
	}

	rendered, err := tm.RenderToString(template.TemplateTypeEmailWelcome, wrappedWelcomeDTO{
		WelcomeTemplateDTO: dto.WelcomeTemplateDTO{
			ClientName:  "Jane Doe",
			ClientEmail: "jane@example.com",
		},
		Extra: "ignored",
	})
	if err != nil {
		t.Fatalf("render returned error: %v", err)
	}

	for _, expected := range []string{"Jane Doe", "jane@example.com", "support@relay.dev", "Relay Labs"} {
		if !strings.Contains(rendered, expected) {
			t.Fatalf("rendered template does not contain %q: %s", expected, rendered)
		}
	}
}
