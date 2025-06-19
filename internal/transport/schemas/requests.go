package schemas

import (
	"fmt"
	"time"
)

type SetRowRequest struct {
	Key   string         `json:"key" validate:"required"`
	Value any            `json:"value" validate:"required"` // Value can be any type, but should be string or []string
	TTL   *time.Duration `json:"ttl,omitempty"`
}

// Validate checks if the SetRowRequest value field is either a string or a slice of strings.
func (r *SetRowRequest) Validate() error {
	if r.Value == nil {
		return nil // No value to validate
	}

	switch v := r.Value.(type) {
	case string, []string:
		return nil
	case []interface{}:
		// Check if all elements are strings
		for _, elem := range v {
			if _, ok := elem.(string); !ok {
				return fmt.Errorf("invalid value type in slice: %T; expected all elements to be string", elem)
			}
		}
		return nil
	default:
		return fmt.Errorf("invalid value type: %T; expected string or []string", v)
	}
}

// PushItemToSliceRequest represents a request to push an item into a slice stored in the database.
type PushItemToSliceRequest struct {
	Value string         `json:"value" validate:"required"` // Value to push into the slice
	TTL   *time.Duration `json:"ttl,omitempty"`             // Optional TTL for the item
}
