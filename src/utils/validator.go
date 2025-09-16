package utils

import "github.com/go-playground/validator/v10"

func FormatValidationError(err error) string {
	if ve, ok := err.(validator.ValidationErrors); ok && len(ve) > 0 {
		fe := ve[0]
		field := fe.Field()
		tag := fe.Tag()
		switch tag {
		case "required":
			return field + " is required"
		case "email":
			return field + " must be a valid email"
		default:
			return field + " failed validation: " + tag
		}
	}
	return "invalid json body: validation failed"
}
