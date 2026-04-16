package app

import (
	"os"
	"path/filepath"

	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/env"
	"github.com/rijum8906/relay/packages/core/template"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func initLogger(config *env.Config) (*zap.Logger, *apperror.AppError) {
	var zapConfig zap.Config

	if config.AppEnv == "production" {
		zapConfig = zap.NewProductionConfig()
		zapConfig.EncoderConfig.TimeKey = "timestamp"
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		zapConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	}

	if config.EnableJSON {
		zapConfig.Encoding = "json"
	}

	if config.LogLevel != "" {
		level, err := zapcore.ParseLevel(config.LogLevel)
		if err != nil {
			level = zapcore.InfoLevel
		}
		zapConfig.Level = zap.NewAtomicLevelAt(level)
	}

	zapConfig.DisableCaller = !config.EnableCaller
	zapConfig.DisableStacktrace = !config.EnableStack

	// Configure output paths
	if config.LogFile != "" {
		// Ensure log directory exists
		if err := os.MkdirAll(filepath.Dir(config.LogFile), 0o755); err != nil {
			return nil, apperror.ErrInternal.
				WithMessage("failed to create log directory").
				WithDetail("error", err.Error())
		}

		zapConfig.OutputPaths = []string{"stdout", config.LogFile}
		zapConfig.ErrorOutputPaths = []string{"stderr", config.LogFile}
	} else {
		zapConfig.OutputPaths = []string{"stdout"}
		zapConfig.ErrorOutputPaths = []string{"stderr"}
	}

	// Build the logger
	logger, err := zapConfig.Build(
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, apperror.ErrInternal.
			WithMessage("failed to create zap logger").
			WithDetail("error", err.Error())
	}

	return logger, nil
}

func initTemplateManager(config *env.Config) (template.TemplateManager, *apperror.AppError) {
	tm, err := template.NewTemplateManagerWithCompanyInfo("packages/templates", &dto.CompanyInfo{
		Name:       "Relay",
		Emails:     []string{"UfNwO@example.com"},
		Addresses:  []string{"123 Main St, Anytown, USA"},
		WebsiteURL: "https://relay.com",
		SocialLinks: []dto.SocialLink{
			{
				Label: "Twitter",
				URL:   "https://twitter.com/relay",
			},
		},
	})
	if err != nil {
		return nil, apperror.ErrInternal.WithMessage("failed to create template manager").WithDetail("error", err.Error())
	}

	return tm, nil
}
