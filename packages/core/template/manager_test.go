package template_test

import (
	"strings"
	"testing"

	"github.com/rijum8906/relay/packages/core/template"
)

func mustCreateTemplateManager() template.TemplateManager {
	m, err := template.NewTemplateManager("../../templates/")
	if err != nil {
		panic(err)
	}

	return m
}

func Test_templateManager_LoadTemplates_Failure(t *testing.T) {
	_, err := template.NewTemplateManager("../templates/")
	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func Test_templateManager_ReloadTemplates(t *testing.T) {
	tm := mustCreateTemplateManager()
	err := tm.ReloadTemplates()
	if err != nil {
		t.Errorf("failed to reload templates: %v", err)
	}
}

func Test_templateManager_RenderToString(t *testing.T) {
	tm := mustCreateTemplateManager()

	// Test Reset Password Email
	resetPassowrdData := template.PasswordResetDTO{
		ClientName:  "John Doe",
		ClientEmail: "example@example.com",
		ResetToken:  "token123",
		Validity:    "15 minutes",
	}

	templateStr, err := tm.RenderToString(template.TemplateTypeEmailPasswordReset, resetPassowrdData)
	if err != nil {
		t.Errorf("failed to render template: %v", err)
	}

	if templateStr == "" {
		t.Errorf("template is empty")
	}

	if !strings.Contains(templateStr, "token123") || !strings.Contains(templateStr, "John Doe") || !strings.Contains(templateStr, "example@example.com") || !strings.Contains(templateStr, "15 minutes") {
		t.Errorf("password reset token is not masked, got %s", templateStr)
	}

	// Test Welcome Email
	welcomeData := template.WelcomeTemplateDTO{
		ClientName:  "John Doe",
		ClientEmail: "example@example.com",
	}

	templateStr, err = tm.RenderToString(template.TemplateTypeEmailWelcome, welcomeData)
	if err != nil {
		t.Errorf("failed to render template: %v", err)
	}

	if templateStr == "" {
		t.Errorf("template is empty")
	}

	if !strings.Contains(templateStr, "John Doe") || !strings.Contains(templateStr, "example@example.com") {
		t.Errorf("password reset token is not masked, got %s", templateStr)
	}

	// Test Email Verification
	verificationData := template.EmailVerificationDTO{
		ClientName:  "John Doe",
		ClientEmail: "example@example.com",
	}

	templateStr, err = tm.RenderToString(template.TemplateTypeEmailVerification, verificationData)
	if err != nil {
		t.Errorf("failed to render template: %v", err)
	}

	if templateStr == "" {
		t.Errorf("template is empty")
	}

	if !strings.Contains(templateStr, "John Doe") || !strings.Contains(templateStr, "example@example.com") {
		t.Errorf("password reset token is not masked, got %s", templateStr)
	}
}

func Test_templateManager_RenderToBytes(t *testing.T) {
	tm := mustCreateTemplateManager()

	// Test Reset Password Email
	resetPassowrdData := template.PasswordResetDTO{
		ClientName:  "John Doe",
		ClientEmail: "example@example.com",
		ResetToken:  "token123",
		Validity:    "15 minutes",
	}

	templateBytes, err := tm.RenderToBytes(template.TemplateTypeEmailPasswordReset, resetPassowrdData)
	if err != nil {
		t.Errorf("failed to render template: %v", err)
	}

	templateStr := string(templateBytes)

	if templateStr == "" {
		t.Errorf("template is empty")
	}

	if !strings.Contains(templateStr, "token123") || !strings.Contains(templateStr, "John Doe") || !strings.Contains(templateStr, "example@example.com") || !strings.Contains(templateStr, "15 minutes") {
		t.Errorf("password reset token is not masked, got %s", templateStr)
	}

	// Test Welcome Email
	welcomeData := template.WelcomeTemplateDTO{
		ClientName:  "John Doe",
		ClientEmail: "example@example.com",
	}

	templateBytes, err = tm.RenderToBytes(template.TemplateTypeEmailWelcome, welcomeData)
	if err != nil {
		t.Errorf("failed to render template: %v", err)
	}
	templateStr = string(templateBytes)

	if templateStr == "" {
		t.Errorf("template is empty")
	}

	if !strings.Contains(templateStr, "John Doe") || !strings.Contains(templateStr, "example@example.com") {
		t.Errorf("password reset token is not masked, got %s", templateStr)
	}

	// Test Email Verification
	verificationData := template.EmailVerificationDTO{
		ClientName:  "John Doe",
		ClientEmail: "example@example.com",
	}

	templateBytes, err = tm.RenderToBytes(template.TemplateTypeEmailVerification, verificationData)
	if err != nil {
		t.Errorf("failed to render template: %v", err)
	}
	templateStr = string(templateBytes)

	if templateStr == "" {
		t.Errorf("template is empty")
	}

	if !strings.Contains(templateStr, "John Doe") || !strings.Contains(templateStr, "example@example.com") {
		t.Errorf("password reset token is not masked, got %s", templateStr)
	}
}
