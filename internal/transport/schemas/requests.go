package schemas

import (
	"encoding/json"
	"memorydb/internal/db"
	"time"
)

// Duration is a wrapper around time.Duration to provide custom JSON marshaling and unmarshaling.
//
// It allows the duration to be represented as a string in JSON format, which is more human-readable.
// For example, a duration of 5 minutes would be represented as "5m" in JSON.
type Duration struct {
	time.Duration
}

// MarshalJSON implements the json.Marshaler interface for the Duration type.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String())
}

// UnmarshalJSON implements the json.Unmarshaler interface for the Duration type.
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = parsed
	return nil
}

// SetRowRequest represents a request to set a new row in the database.
type SetRowRequest struct {
	Key   string           `json:"key" validate:"required"`
	Value db.StringOrSlice `json:"value" validate:"required"` // Value can be any type, but should be string or []string
	TTL   *Duration        `json:"ttl,omitempty"`
}

// UpdateRowRequest represents a request to update an existing row in the database.
type UpdateRowRequest struct {
	Value db.StringOrSlice `json:"value" validate:"required"` // Value can be any type, but should be string or []string
	TTL   *Duration        `json:"ttl,omitempty"`             // Optional TTL for the item
}

// PushItemToSliceRequest represents a request to push an item into a slice stored in the database.
type PushItemToSliceRequest struct {
	Value string    `json:"value" validate:"required"` // Value to push into the slice
	TTL   *Duration `json:"ttl,omitempty"`             // Optional TTL for the item
}
