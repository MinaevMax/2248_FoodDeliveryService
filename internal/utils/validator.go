package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

func IsBcryptHash(s string) bool {
	// bcrypt-хэши обычно начинаются с $2a$, $2b$ или $2y$ и имеют длину 60 символов
	return len(s) == 60 && (strings.HasPrefix(s, "$2a$") || strings.HasPrefix(s, "$2b$") || strings.HasPrefix(s, "$2y$"))
}

func FormatValidationErrors(errs validator.ValidationErrors) string {
	messages := make([]string, 0, len(errs))
	for _, e := range errs {
		switch e.Tag() {
		case "required":
			messages = append(messages, fmt.Sprintf("%s is required", e.Field()))
		case "oneof":
			messages = append(messages, fmt.Sprintf("%s must be one of %s", e.Field(), e.Param()))
		case "min":
			messages = append(messages, fmt.Sprintf("%s must be at least %s characters", e.Field(), e.Param()))
		default:
			messages = append(messages, fmt.Sprintf("%s is invalid", e.Field()))
		}
	}
	return strings.Join(messages, "; ")
}
