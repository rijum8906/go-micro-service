package app

import (
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/packages/core/corelogger"
	"github.com/rijum8906/relay/packages/core/dto"
	"github.com/rijum8906/relay/packages/core/template"
	"github.com/rijum8906/relay/services/notification-service/app/config"
	"go.uber.org/zap"
)

func initLogger(config *config.Env) (*zap.Logger, *apperror.AppError) {
	return corelogger.InitLogger(corelogger.LoggerConfig{
		AppEnv:       config.AppEnv,
		EnableJSON:   config.EnableJSON,
		EnableCaller: config.EnableCaller,
		EnableStack:  config.EnableStack,
		LogLevel:     config.LogLevel,
		LogFile:      config.LogFile,
	})
}

func initTemplateManager() (template.TemplateManager, *apperror.AppError) {
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
