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
		case "len":
			return field + " must be " + fe.Param() + " characters long"
		case "numeric":
			return field + " must be numeric"
		case "uuid4":
			return field + " must be a valid UUIDv4"
		case "oneof":
			return field + " must be one of: " + fe.Param()
		case "min":
			return field + " must be at least " + fe.Param()
		case "gte":
			return field + " must be greater than or equal to " + fe.Param()
		case "lte":
			return field + " must be less than or equal to " + fe.Param()
		default:
			return field + " failed validation: " + tag
		}
	}
	return "invalid json body: validation failed"
}
