package godb

import (
	"encoding/json"
	"fmt"
	"memorydb/internal/transport/schemas"
)

type ApiResponse schemas.RowResponse

// UnmarshalJSON customizes the JSON unmarshalling for ApiResponse. In Go, when we
// try to unmarshal a list of strings into a field of type `any`, it is unmarshalled as a slice of empty interfaces (`[]interface{}`).
// Therefore, we force the unmarshalling of the `value` field as either a string or a slice of strings,
// to avoid this issue.
//
// This is just a workaround. In cases where more complex types are used, we should
// handle this differently.
func (r *ApiResponse) UnmarshalJSON(data []byte) error {
	type alias ApiResponse
	aux := &struct {
		Value json.RawMessage `json:"value"`
		*alias
	}{
		alias: (*alias)(r),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Try []string first
	var strSlice []string
	if err := json.Unmarshal(aux.Value, &strSlice); err == nil {
		r.Value = strSlice
		return nil
	}

	// Try string
	var str string
	if err := json.Unmarshal(aux.Value, &str); err == nil {
		r.Value = str
		return nil
	}

	return fmt.Errorf("unsupported type for value")
}
