package schemas

import (
	"time"
)

// OKResponse represents a standard response structure for successful operations.
type OKResponse struct {
	Message string `json:"message"`
}

// RowResponse represents a response structure for a single row in the memory database.
type RowResponse struct {
	Key       string    `json:"key"`
	Kind      string    `json:"kind"`
	Value     any       `json:"value"`
	TTL       time.Time `json:"ttl"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// // UnmarshalJSON customizes the JSON unmarshalling for RowResponse. Since the database
// // only accepts []string as list values, we force to unmarshal the `value` field
// // as either a string or a slice of strings. If we don't do this, the slice is
// // going to be unmarshalled as a []interface{} which is not what we want.
// //
// // Alternatively, we could handle this in the client code exported in the package gomemdb.
// func (r *RowResponse) UnmarshalJSON(data []byte) error {
// 	type alias RowResponse
// 	aux := &struct {
// 		Value json.RawMessage `json:"value"`
// 		*alias
// 	}{
// 		alias: (*alias)(r),
// 	}

// 	if err := json.Unmarshal(data, &aux); err != nil {
// 		return err
// 	}

// 	// Try []string first
// 	var strSlice []string
// 	if err := json.Unmarshal(aux.Value, &strSlice); err == nil {
// 		r.Value = strSlice
// 		return nil
// 	}

// 	// Try string
// 	var str string
// 	if err := json.Unmarshal(aux.Value, &str); err == nil {
// 		r.Value = str
// 		return nil
// 	}

// 	return fmt.Errorf("unsupported type for value")
// }
