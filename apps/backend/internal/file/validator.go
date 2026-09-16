package file

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func formatValidationErrors(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		var errs []string
		for _, e := range validationErrors {
			field := strings.ToLower(e.Field())
			switch e.Tag() {
			case "required":
				errs = append(errs, fmt.Sprintf("%s is required", field))
			case "min":
				errs = append(errs, fmt.Sprintf("%s must be at least %s characters", field, e.Param()))
			case "max":
				errs = append(errs, fmt.Sprintf("%s must be at most %s characters", field, e.Param()))
			default:
				errs = append(errs, fmt.Sprintf("%s is invalid", field))
			}
		}
		return strings.Join(errs, "; ")
	}
	return err.Error()
}
