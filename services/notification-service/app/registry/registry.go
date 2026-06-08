package registry

import (
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/notification-service/app"
	organizationservice "github.com/rijum8906/relay/services/notification-service/internal/services/organization"
	userservice "github.com/rijum8906/relay/services/notification-service/internal/services/user"
)

func Run(application *app.Application) *apperror.AppError {
	userService, appErr := userservice.New()
	if appErr != nil {
		return appErr
	}

	if appErr = userService.CreateStream(); appErr != nil {
		return appErr
	}

	if appErr = userService.Run(); appErr != nil {
		return appErr
	}

	orgService, appErr := organizationservice.New()
	if appErr != nil {
		return appErr
	}

	if appErr = orgService.CreateStream(); appErr != nil {
		return appErr
	}

	if appErr = orgService.Run(); appErr != nil {
		return appErr
	}

	return application.Run()
}
