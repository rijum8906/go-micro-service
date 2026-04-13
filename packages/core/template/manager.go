// Package template
package template

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"path/filepath"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/rijum8906/relay/packages/core/apperror"
)

var validate = validator.New()

const templateFileExtension = ".html"

type TemplateManager interface {
	RenderToBytes(templateType TemplateType, data any) ([]byte, *apperror.AppError)
	RenderToString(templateType TemplateType, data any) (string, *apperror.AppError)
	ReloadTemplates() *apperror.AppError
	ValidateData(data any) *apperror.AppError
}

type templateManager struct {
	templatesDir string
	info         *CompanyInfo
	templates    *template.Template
	mu           sync.RWMutex
}

func NewTemplateManager(templatesDir string) (TemplateManager, error) {
	return NewTemplateManagerWithCompanyInfo(templatesDir, nil)
}

func NewTemplateManagerWithCompanyInfo(templatesDir string, companyInfo *CompanyInfo) (TemplateManager, error) {
	templatesDir = strings.TrimSpace(templatesDir)
	if templatesDir == "" {
		return nil, errors.New("templates directory is required")
	}

	if companyInfo != nil {
		if err := validate.Struct(companyInfo); err != nil {
			return nil, fmt.Errorf("invalid company info: %w", err)
		}
	}

	tm := &templateManager{
		templatesDir: templatesDir,
		info:         companyInfo,
	}

	if err := tm.loadTemplates(); err != nil {
		return nil, err
	}

	return tm, nil
}

func (m *templateManager) loadTemplates() error {
	if m == nil {
		return errors.New("template manager is nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	pattern := filepath.Join(m.templatesDir, "*"+templateFileExtension)
	tmpl, err := template.ParseGlob(pattern)
	if err != nil {
		return fmt.Errorf("failed to load templates from %s: %w", m.templatesDir, err)
	}
	if tmpl == nil {
		return fmt.Errorf("failed to load templates from %s: no templates parsed", m.templatesDir)
	}

	m.templates = tmpl
	return nil
}

func (m *templateManager) ValidateData(data any) *apperror.AppError {
	if data == nil {
		return apperror.New(apperror.CodeValidation, "failed to validate template data").
			WithDetail("error", "template data is required")
	}

	err := validate.Struct(data)
	if err != nil {
		return apperror.New(apperror.CodeValidation, "failed to validate template data").
			WithDetail("error", err.Error())
	}

	return nil
}

func (m *templateManager) ReloadTemplates() *apperror.AppError {
	err := m.loadTemplates()
	if err != nil {
		return apperror.New(apperror.CodeInternal, "failed to reload templates").WithDetail("error", err.Error())
	}

	return nil
}

func (m *templateManager) RenderToBytes(templateType TemplateType, data any) ([]byte, *apperror.AppError) {
	if err := m.ValidateData(data); err != nil {
		return nil, err
	}

	templateName, err := normalizeTemplateName(templateType)
	if err != nil {
		return nil, err
	}

	if m == nil {
		return nil, apperror.New(apperror.CodeInternal, "template manager is not initialized")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.templates == nil {
		return nil, apperror.New(apperror.CodeInternal, "templates are not loaded")
	}

	if m.templates.Lookup(templateName) == nil {
		return nil, apperror.New(apperror.CodeInternal, "template not found").
			WithDetail("template", templateName)
	}

	templateData, err := m.buildTemplateData(data)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer

	if err := m.templates.ExecuteTemplate(&buf, templateName, templateData); err != nil {
		return nil, apperror.New(apperror.CodeInternal, "failed to render template").
			WithDetail("template", templateName).
			WithDetail("error", err.Error())
	}

	return buf.Bytes(), nil
}

func (m *templateManager) RenderToString(templateType TemplateType, data any) (string, *apperror.AppError) {
	bytes, err := m.RenderToBytes(templateType, data)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func normalizeTemplateName(templateType TemplateType) (string, *apperror.AppError) {
	name := strings.TrimSpace(string(templateType))
	if name == "" {
		return "", apperror.New(apperror.CodeValidation, "template type is required")
	}

	return name + templateFileExtension, nil
}

func (m *templateManager) buildTemplateData(data any) (any, *apperror.AppError) {
	if m.info == nil {
		return data, nil
	}

	if data == nil {
		return nil, apperror.New(apperror.CodeValidation, "template data is required")
	}

	value := reflect.ValueOf(data)
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nil, apperror.New(apperror.CodeValidation, "template data is required")
		}
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Struct:
		templateData := map[string]any{}
		flattenTemplateFields(templateData, value)
		templateData["CompanyInfo"] = m.info
		return templateData, nil
	case reflect.Map:
		if value.Type().Key().Kind() == reflect.String {
			templateData := map[string]any{}
			iter := value.MapRange()
			for iter.Next() {
				templateData[iter.Key().String()] = iter.Value().Interface()
			}
			templateData["CompanyInfo"] = m.info
			return templateData, nil
		}
	}

	return map[string]any{
		"Data":        data,
		"CompanyInfo": m.info,
	}, nil
}

func flattenTemplateFields(out map[string]any, value reflect.Value) {
	for value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return
		}
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return
	}

	valueType := value.Type()
	for i := range value.NumField() {
		field := valueType.Field(i)
		fieldValue := value.Field(i)

		if field.Anonymous {
			flattenTemplateFields(out, fieldValue)
			continue
		}

		if !field.IsExported() {
			continue
		}

		out[field.Name] = fieldValue.Interface()
	}
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
