package db

// DBClient defines the interface for interacting with an in-memory database.
type DBClient interface {
	// Get retrieves an item by its key.
	Get(key string) (*Item, error)

	// Set stores an item with the specified key and optional options.
	Set(key string, value any, opts ...ItemOptions) error

	// Update modifies an existing item with the specified key and value.
	Update(key string, value any, opts ...ItemOptions) error

	// Remove deletes an item by its key.
	Remove(key string) error

	// Push adds a new item to the memory database with the specified key and value.
	Push(key string, values []string, opts ...ItemOptions) (*Item, error)

	// Pop removes and returns the last item from a slice stored at the specified key.
	Pop(key string) (*Item, error)

	// Close releases any resources held by the database client.
	Close()
}
