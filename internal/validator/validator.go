package validator

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator reprents an interface that objects can implement to specify custom validation rules.
type Validator interface {
	Validate() error // Validate validates an object.
}

// ValidateJSON validates a given struct with json tags, and returns an error if the validation fails.
func ValidateJSON(obj any) error {
	v := validator.New(validator.WithRequiredStructEnabled())
	if err := handleValidationErrors(v.Struct(obj)); err != nil {
		return err
	}

	// run additional validation rules if any
	// check if object has Validate method, if so, call it
	if v, ok := obj.(Validator); ok {
		if err := v.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// handleValidationErrors converts validation errors to a human readable format.
//
// By modifying this function, you can customize the error messages returned to the user
// when validation fails. For example, in this case we haved defined that when a field is required,
// the error message will be "field 'x' is required".
func handleValidationErrors(err error) error {
	if err == nil {
		return nil
	}

	validationErrors := err.(validator.ValidationErrors)
	errorsMsg := make([]string, len(validationErrors))
	for i, validationError := range validationErrors {
		fieldName := validationError.Field()
		switch validationError.Tag() {
		case "required":
			errorsMsg[i] = "field '" + fieldName + "' is required"
		default:
			errorsMsg[i] = validationError.Error()
		}
	}

	return errors.New(strings.Join(errorsMsg, ", "))
}
