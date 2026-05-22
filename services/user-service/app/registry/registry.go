// Package registry
package registry

import (
	"github.com/rijum8906/relay/packages/core/apperror"
	"github.com/rijum8906/relay/services/user/app"
)

func Run(application *app.Application) *apperror.AppError {
	return application.Run()
}
