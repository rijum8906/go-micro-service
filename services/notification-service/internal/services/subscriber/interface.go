package subscriber

import (
	"github.com/rijum8906/relay/packages/core/apperror"
)

type Service interface {
	Subscribe() *apperror.AppError
}
