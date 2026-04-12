// Package template
package template

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"
	"sync"
)

type TemplateManager interface {
	RenderToBytes(templateType TemplateType, data any) ([]byte, error)
	RenderToString(templateType TemplateType, data any) (string, error)
	ReloadTemplates() error
}

type templateManager struct {
	templatesDir string
	templates    *template.Template
	mu           sync.RWMutex
}

func NewTemplateManager(templatesDir string) (TemplateManager, error) {
	tm := &templateManager{
		templatesDir: templatesDir,
	}

	if err := tm.loadTemplates(); err != nil {
		return nil, err
	}

	return tm, nil
}

func (m *templateManager) loadTemplates() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pattern := filepath.Join(m.templatesDir, "*.html")
	tmpl, err := template.ParseGlob(pattern)
	if err != nil {
		return fmt.Errorf("failed to load templates from %s: %w", m.templatesDir, err)
	}

	m.templates = tmpl
	return nil
}

func (m *templateManager) ReloadTemplates() error {
	return m.loadTemplates()
}

func (m *templateManager) RenderToBytes(templateType TemplateType, data any) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.templates == nil {
		return nil, fmt.Errorf("templates not loaded")
	}

	templateName := string(templateType) + ".html"
	var buf bytes.Buffer

	if err := m.templates.ExecuteTemplate(&buf, templateName, data); err != nil {
		return nil, fmt.Errorf("failed to render template %s: %w", templateType, err)
	}

	return buf.Bytes(), nil
}

func (m *templateManager) RenderToString(templateType TemplateType, data any) (string, error) {
	bytes, err := m.RenderToBytes(templateType, data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// Type-safe helper methods
func (m *templateManager) RenderWelcomeEmail(data WelcomeTemplateDTO) ([]byte, error) {
	return m.RenderToBytes(TemplateTypeEmailWelcome, data)
}

func (m *templateManager) RenderVerificationEmail(data EmailVerificationDTO) ([]byte, error) {
	return m.RenderToBytes(TemplateTypeEmailVerification, data)
}

func (m *templateManager) RenderPasswordResetEmail(data PasswordResetDTO) ([]byte, error) {
	return m.RenderToBytes(TemplateTypeEmailPasswordReset, data)
}
