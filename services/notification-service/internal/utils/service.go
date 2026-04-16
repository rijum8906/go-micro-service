package utils

import (
	"fmt"
	"reflect"

	"github.com/rijum8906/relay/packages/core/apperror"
)

func EnsureAllFields(s interface{}) *apperror.AppError {
	val := reflect.ValueOf(s)
	// Handle pointers
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldName := val.Type().Field(i).Name

		// Check if the field is the zero value for its type
		if reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()) {
			return apperror.ErrValidation.WithDetail("missing_field", fmt.Sprintf("field %s is required", fieldName))
		}
	}
	return nil
}
