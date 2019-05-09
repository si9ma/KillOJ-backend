package validate

import (
	"gopkg.in/go-playground/validator.v8"
)

// check is validation errors,
// is true, return fields ( ActualTag)
func IsValidationErrors(err error) (fields []string, is bool) {
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		// not ValidationErrors
		return nil, false
	}

	for _, err := range errs {
		fields = append(fields, err.ActualTag)
	}

	return fields, true
}
