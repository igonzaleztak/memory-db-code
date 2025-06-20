package db

import (
	"encoding/json"
	"time"
)

// DataType represents the type of data stored in the item. There are two types:
// StringType for a single string value and StringSliceType for a slice of strings.
type DataType int

const (
	StringType DataType = iota
	StringSliceType
)

var MappingDataType = map[DataType]string{
	StringType:      "string",
	StringSliceType: "string_slice",
}

const (
	defaultTTL = 5 * time.Minute
)

// ItemOptions is an interface that allows for applying options to an item.
// In this case, the only option that is available is WithTTL, which sets the time-to-live for the item to a custom value.
type ItemOptions interface {
	apply(*Item)
}

// WithTTL sets the time-to-live (TTL) for an item. The TTL is the time after which the item will expire.
type WithTTL time.Duration

func (o WithTTL) apply(opts *Item) {
	opts.TTL = time.Now().Add(time.Duration(o))
}

// StringOrSlice is a custom type that can hold either a string or a slice of strings.
//
// It implements the json.Unmarshaler and json.Marshaler interfaces to handle JSON serialization and deserialization.
// This ensures that the value is correctly interpreted as either a single string or a slice of strings when unmarshaling from JSON.
type StringOrSlice struct {
	Val any // Value can be string or []string
}

// UnmarshalJSON implements the json.Unmarshaler interface for StringOrSlice.
// It attempts to unmarshal the JSON data into either a string or a slice of strings,
// avoiding it to be unmarshaled into a generic []interface{} type.
func (s *StringOrSlice) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		s.Val = str
		return nil
	}

	var slice []string
	if err := json.Unmarshal(data, &slice); err == nil {
		s.Val = slice
		return nil
	}

	return ErrInvalidDataType // Return an error if neither type matches
}

// MarshalJSON implements the json.Marshaler interface for StringOrSlice.
// In this case the marshalJSON method writes the value within the StringOrSlice wrapper and not the wrapper itself.
//
// Example: It will marshal the struct as this:
//
//		{
//	       "key": "value",
//		   "value": "some string"
//		}
//
// Instead of:
//
//		{
//	       "key": "value",
//		   "value": {
//		       "Val": "some string"
//		   }
func (s StringOrSlice) MarshalJSON() ([]byte, error) {
	switch v := s.Val.(type) {
	case nil:
		return json.Marshal(nil)
	case string:
		return json.Marshal(v)
	case []string:
		return json.Marshal(v)
	default:
		return nil, ErrInvalidDataType
	}
}

// item represents a single item in the memory database. It would be similar to a row in a traditional database.
type Item struct {
	Value     *StringOrSlice `json:"value"`         // Value can be string or []string
	TTL       time.Time      `json:"ttl,omitempty"` // TTL is optional and will be omitted if not set
	Kind      DataType       `json:"kind"`          // Kind is used internally to determine the data type of the value
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// newItem creates a new item with the given value and options
func newItem(value any, opts ...ItemOptions) (*Item, error) {
	dataToBeStored := &Item{
		TTL:       calculateExpirationTime(defaultTTL),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, opt := range opts {
		opt.apply(dataToBeStored)
	}

	// Determine the type of value and set the Kind and Value fields accordingly, so it's easier to work with later.
	switch v := value.(type) {
	case string:
		dataToBeStored.Kind = StringType
		dataToBeStored.Value = &StringOrSlice{Val: v}
		return dataToBeStored, nil
	case []string:
		dataToBeStored.Kind = StringSliceType
		dataToBeStored.Value = &StringOrSlice{Val: v}
		return dataToBeStored, nil
	case []interface{}:
		// check if all elements are strings
		stringSlice := make([]string, len(v))
		for i, elem := range v {
			if _, ok := elem.(string); !ok {
				return nil, ErrInvalidDataType
			}
			stringSlice[i] = elem.(string)
		}
		dataToBeStored.Kind = StringSliceType
		dataToBeStored.Value = &StringOrSlice{Val: stringSlice}
		return dataToBeStored, nil
	default:
		return nil, ErrInvalidDataType
	}
}

// update modifies the value of an existing item in the database.
func (d *Item) update(value any, updatedAt time.Time, opts ...ItemOptions) error {
	if d.Value == nil {
		return ErrDataNotFound
	}

	// check whether the value is []interface{}. If it is, convert it to []string.
	// Otherwise, if the value is a string, set it directly.
	switch v := value.(type) {
	case string:
		d.Kind = StringType
		d.Value = &StringOrSlice{Val: v}
	case []string:
		d.Kind = StringSliceType
		d.Value = &StringOrSlice{Val: v}
	case []interface{}:
		// check if all elements are strings
		stringSlice := make([]string, len(v))
		for i, elem := range v {
			if str, ok := elem.(string); ok {
				stringSlice[i] = str
			} else {
				return ErrInvalidDataType
			}
		}
		d.Kind = StringSliceType
		d.Value = &StringOrSlice{Val: stringSlice}
	default:
		return ErrInvalidDataType // Return an error if the value type is not supported
	}

	// apply options to set TTL and other properties
	for _, opt := range opts {
		opt.apply(d)
	}

	// Update the timestamp to the current time
	d.UpdatedAt = updatedAt
	return nil
}

// pushToSlice adds one or more values to a slice stored in the item.
func (d *Item) pushToSlice(updatedAt time.Time, value string) error {
	if d.Kind != StringSliceType {
		return ErrInvalidDataType
	}
	slice, ok := d.Value.Val.([]string)
	if !ok {
		return ErrInvalidDataType
	}
	d.Value = &StringOrSlice{Val: append(slice, value)}
	d.UpdatedAt = updatedAt
	return nil
}

// popFromSlice removes the last value from a slice
func (d *Item) popFromSlice(updatedAt time.Time) error {
	if d.Kind != StringSliceType {
		return ErrInvalidDataType
	}
	slice, ok := d.Value.Val.([]string)
	if !ok || len(slice) == 0 {
		return ErrDataNotFound
	}
	d.Value = &StringOrSlice{Val: slice[:len(slice)-1]}
	d.UpdatedAt = updatedAt
	return nil
}

// isExpired checks if the item has expired based on its TTL.
func (d *Item) isExpired() bool {
	return d.TTL.Before(time.Now())
}

// calculateExpirationTime calculates the expiration time based on the given TTL.
func calculateExpirationTime(ttl time.Duration) time.Time {
	return time.Now().Add(ttl)
}
