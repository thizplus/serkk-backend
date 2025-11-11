package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps the validator instance
type Validator struct {
	validate *validator.Validate
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag"`
	Value   string `json:"value,omitempty"`
}

// New creates a new validator instance
func New() *Validator {
	v := validator.New()

	// Register custom validators here if needed
	registerCustomValidators(v)

	return &Validator{
		validate: v,
	}
}

// Validate validates a struct and returns validation errors
func (v *Validator) Validate(data interface{}) []ValidationError {
	var validationErrors []ValidationError

	err := v.validate.Struct(data)
	if err == nil {
		return nil
	}

	// Type assert to validator.ValidationErrors
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			validationErrors = append(validationErrors, ValidationError{
				Field:   getFieldName(e),
				Message: getErrorMessage(e),
				Tag:     e.Tag(),
				Value:   fmt.Sprintf("%v", e.Value()),
			})
		}
	}

	return validationErrors
}

// ValidateVar validates a single variable
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	return v.validate.Var(field, tag)
}

// getFieldName extracts the JSON field name or falls back to struct field name
func getFieldName(e validator.FieldError) string {
	field := e.Field()

	// Convert from PascalCase to camelCase for JSON compatibility
	if len(field) > 0 {
		return strings.ToLower(field[:1]) + field[1:]
	}

	return field
}

// getErrorMessage returns a human-readable error message
func getErrorMessage(e validator.FieldError) string {
	field := getFieldName(e)

	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters", field, e.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, e.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, e.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, e.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, e.Param())
	case "alpha":
		return fmt.Sprintf("%s must contain only alphabetic characters", field)
	case "alphanum":
		return fmt.Sprintf("%s must contain only alphanumeric characters", field)
	case "numeric":
		return fmt.Sprintf("%s must be a valid number", field)
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "uri":
		return fmt.Sprintf("%s must be a valid URI", field)
	case "uuid":
		return fmt.Sprintf("%s must be a valid UUID", field)
	case "len":
		return fmt.Sprintf("%s must be exactly %s characters", field, e.Param())
	case "eq":
		return fmt.Sprintf("%s must be equal to %s", field, e.Param())
	case "ne":
		return fmt.Sprintf("%s must not be equal to %s", field, e.Param())
	case "oneof":
		return fmt.Sprintf("%s must be one of [%s]", field, e.Param())
	case "containsany":
		return fmt.Sprintf("%s must contain at least one of the following characters: %s", field, e.Param())
	case "excludes":
		return fmt.Sprintf("%s must not contain '%s'", field, e.Param())
	case "startswith":
		return fmt.Sprintf("%s must start with '%s'", field, e.Param())
	case "endswith":
		return fmt.Sprintf("%s must end with '%s'", field, e.Param())
	default:
		return fmt.Sprintf("%s is invalid (failed %s validation)", field, e.Tag())
	}
}

// registerCustomValidators registers custom validation rules
func registerCustomValidators(v *validator.Validate) {
	// Register custom validators here
	// Example:
	// v.RegisterValidation("custom_tag", customValidationFunc)

	// Username validation: alphanumeric, underscore, hyphen, 3-30 chars
	v.RegisterValidation("username", func(fl validator.FieldLevel) bool {
		username := fl.Field().String()
		if len(username) < 3 || len(username) > 30 {
			return false
		}
		for _, char := range username {
			if !((char >= 'a' && char <= 'z') ||
				(char >= 'A' && char <= 'Z') ||
				(char >= '0' && char <= '9') ||
				char == '_' || char == '-') {
				return false
			}
		}
		return true
	})

	// Password strength validation: min 8 chars, must contain number and letter
	v.RegisterValidation("password", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		if len(password) < 8 {
			return false
		}

		hasLetter := false
		hasNumber := false

		for _, char := range password {
			if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
				hasLetter = true
			}
			if char >= '0' && char <= '9' {
				hasNumber = true
			}
		}

		return hasLetter && hasNumber
	})
}
