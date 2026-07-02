// Package validator wraps go-playground/validator with a single shared instance
// and provides a human-readable field-error formatter for DTO validation.
//
// Architectural decision: validator.New() is expensive — it uses reflection to
// parse struct tags. A single instance is created here and reused across all
// DTOs in the application. This follows the same pattern as database connection
// pooling: create once, share everywhere.
//
// Usage in a handler:
//
//	if err := validator.Validate(dto); err != nil {
//	    return response.ValidationError(c, err)
//	}
package validator

import (
	"reflect"
	"sync"

	"github.com/go-playground/validator/v10"
)

// FieldError represents a single failed validation rule on a DTO field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

var (
	instance *validator.Validate
	once     sync.Once
)

// getInstance returns the singleton validator instance, creating it on first call.
func getInstance() *validator.Validate {
	once.Do(func() {
		v := validator.New()

		// Use JSON field names in error messages instead of Go struct field names.
		// This ensures the API client sees "first_name" instead of "FirstName".
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := fld.Tag.Get("json")
			if name == "" || name == "-" {
				name = fld.Tag.Get("form")
			}
			// Strip options like "omitempty": take only the part before the first comma.
			for i := 0; i < len(name); i++ {
				if name[i] == ',' {
					return name[:i]
				}
			}
			return name
		})

		instance = v
	})
	return instance
}

// Validate validates a DTO struct using struct tags.
// Returns a slice of FieldError on failure, or nil on success.
// The returned value is ready to be passed directly to response.ValidationError().
func Validate(dto any) []FieldError {
	v := getInstance()
	err := v.Struct(dto)
	if err == nil {
		return nil
	}

	var errs validator.ValidationErrors
	if !asValidationErrors(err, &errs) {
		// Not a validation error — propagate as-is (caller should handle separately).
		return []FieldError{{Field: "_", Message: err.Error()}}
	}

	fieldErrors := make([]FieldError, 0, len(errs))
	for _, e := range errs {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   e.Field(),
			Message: humanizeTag(e),
		})
	}
	return fieldErrors
}

// asValidationErrors is a type-assertion helper for validator.ValidationErrors.
func asValidationErrors(err error, target *validator.ValidationErrors) bool {
	switch e := err.(type) {
	case validator.ValidationErrors:
		*target = e
		return true
	}
	return false
}

// humanizeTag converts a validator tag name into a human-readable message.
// Extend this switch to cover custom validation tags added in the future.
func humanizeTag(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return e.Field() + " is required"
	case "email":
		return e.Field() + " must be a valid email address"
	case "min":
		return e.Field() + " must be at least " + e.Param() + " characters"
	case "max":
		return e.Field() + " must be at most " + e.Param() + " characters"
	case "len":
		return e.Field() + " must be exactly " + e.Param() + " characters"
	case "numeric":
		return e.Field() + " must be a numeric value"
	case "url":
		return e.Field() + " must be a valid URL"
	case "uuid4":
		return e.Field() + " must be a valid UUID v4"
	case "oneof":
		return e.Field() + " must be one of: " + e.Param()
	case "gt":
		return e.Field() + " must be greater than " + e.Param()
	case "gte":
		return e.Field() + " must be greater than or equal to " + e.Param()
	case "lt":
		return e.Field() + " must be less than " + e.Param()
	case "lte":
		return e.Field() + " must be less than or equal to " + e.Param()
	default:
		return e.Field() + " failed validation: " + e.Tag()
	}
}
