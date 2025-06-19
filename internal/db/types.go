package db

import (
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
type ItemOptions interface {
	apply(*Item)
}

// WithTTL sets the time-to-live (TTL) for an item. The TTL is the time after which the item will expire.
type WithTTL time.Duration

func (o WithTTL) apply(opts *Item) {
	opts.TTL = time.Now().Add(time.Duration(o))
}

// item represents a single item in the memory database. It would be similar to a row in a traditional database.
type Item struct {
	Value     any       `json:"value"`         // Value can be string or []string
	TTL       time.Time `json:"ttl,omitempty"` // TTL is optional and will be omitted if not set
	Kind      DataType  `json:"-"`             // Kind is used internally to determine the data type of the value
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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
		dataToBeStored.Value = v
		return dataToBeStored, nil
	case []string:
		dataToBeStored.Kind = StringSliceType
		dataToBeStored.Value = v
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
		dataToBeStored.Value = stringSlice
		return dataToBeStored, nil
	default:
		return nil, ErrInvalidDataType
	}
}

// update modifies the value of an existing item in the database.
func (d *Item) update(value any, opts ...ItemOptions) error {
	if d.Value == nil {
		return ErrDataNotFound
	}

	switch d.Kind {
	case StringType:
		str, ok := value.(string)
		if !ok {
			return ErrInvalidDataType
		}
		d.Value = str
	case StringSliceType:
		slice, ok := value.([]string)
		if !ok {
			return ErrInvalidDataType
		}
		d.Value = slice
	default:
		return ErrInvalidDataType
	}

	// apply options to set TTL and other properties
	for _, opt := range opts {
		opt.apply(d)
	}

	// Update the timestamp to the current time
	d.UpdatedAt = time.Now()
	return nil
}

// pushToSlice adds one or more values to a slice stored in the item.
func (d *Item) pushToSlice(values ...string) error {
	if d.Kind != StringSliceType {
		return ErrInvalidDataType
	}
	slice, ok := d.Value.([]string)
	if !ok {
		return ErrInvalidDataType
	}
	d.Value = append(slice, values...)
	d.UpdatedAt = time.Now()
	return nil
}

// popFromSlice removes the last value from a slice
func (d *Item) popFromSlice() error {
	if d.Kind != StringSliceType {
		return ErrInvalidDataType
	}
	slice, ok := d.Value.([]string)
	if !ok || len(slice) == 0 {
		return ErrDataNotFound
	}
	d.Value = slice[:len(slice)-1]
	d.UpdatedAt = time.Now()
	return nil
}

func (d *Item) isExpired() bool {
	return d.TTL.Before(time.Now())
}

func calculateExpirationTime(ttl time.Duration) time.Time {
	return time.Now().Add(ttl)
}
