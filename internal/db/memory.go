package db

import (
	"fmt"
	"sync"
	"time"
)

var (
	// Ensure memoryDB implements DBClient interface
	_ DBClient = (*memoryDB)(nil)
)

const (
	defaultCleanupInterval = 5 * time.Minute // default interval for cleanup routine
)

type DBOptions interface {
	apply(*memoryDB)
}

// WithCleanupInterval sets the interval for the cleanup routine.
type WithCleanupInterval time.Duration

func (o WithCleanupInterval) apply(db *memoryDB) {
	db.cleanupInterval = time.Duration(o)
}

// memoryDB represents an in-memory database that stores items with optional expiration.
type memoryDB struct {
	store           map[string]*Item // in-memory store for items
	cleanupInterval time.Duration    // interval for cleanup routine
	mu              sync.RWMutex     // mutex for preventing race conditions
	stopChan        chan struct{}    // channel to stop the cleanup routine
}

// NewmemoryDB creates a new instance of memoryDB with an initialized store.
func NewMemoryDB(opts ...DBOptions) DBClient {
	db := &memoryDB{
		store:           make(map[string]*Item),
		cleanupInterval: defaultCleanupInterval,
		stopChan:        make(chan struct{}),
	}

	// Apply options to the memoryDB instance
	for _, opt := range opts {
		opt.apply(db)
	}

	// Start a cleanup routine to remove expired items every 5 minutes
	go db.startCleanupRoutine()
	return db
}

// Get retrieves an item from the memory database by its key.
func (db *memoryDB) Get(key string) (*Item, error) {
	// instead of using RLock, we use Lock here to ensure that we can check for expiration and remove items atomically
	db.mu.Lock()
	defer db.mu.Unlock()
	value, exists := db.store[key]
	if !exists {
		e := ErrDataNotFound
		e.Message = fmt.Sprintf("key '%s' not found in memory database", key)
		e.SysMessage = e.Message
		return nil, e
	}

	if value.isExpired() {
		delete(db.store, key) // Remove expired item
		return nil, ErrKeyHasExpired
	}

	return value, nil
}

// Set stores an item in the memory database with the specified key and value.
func (db *memoryDB) Set(key string, value any, opts ...ItemOptions) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	itemToStore, err := newItem(value, opts...)
	if err != nil {
		return fmt.Errorf("failed to create value for key %s: %w", key, err)
	}

	db.store[key] = itemToStore
	return nil
}

// Update updates an existing item in the memory database with the specified key and value.
func (db *memoryDB) Update(key string, value any, opts ...ItemOptions) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	itemToUpdate, exists := db.store[key]
	if !exists {
		return fmt.Errorf("key %s not found for update", key)
	}

	return itemToUpdate.update(value, opts...)
}

// Remove deletes an item from the memory database by its key.
func (db *memoryDB) Remove(key string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.store[key]; !exists {
		return fmt.Errorf("key %s not found for removal", key)
	}

	delete(db.store, key)
	return nil
}

// Push adds a new item to the memory database with the specified key and value.
func (db *memoryDB) Push(key string, values []string, opts ...ItemOptions) (*Item, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	item, exists := db.store[key]
	if !exists {
		return nil, fmt.Errorf("key %s not found for push", key)
	}

	if err := item.pushToSlice(values...); err != nil {
		return nil, fmt.Errorf("failed to push values to key %s: %w", key, err)
	}

	return item, nil
}

// Pop removes the last item from the slice stored at the specified key in the memory database.
func (db *memoryDB) Pop(key string) (*Item, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	item, exists := db.store[key]
	if !exists {
		return nil, fmt.Errorf("key %s not found for pop", key)
	}

	if err := item.popFromSlice(); err != nil {
		return nil, fmt.Errorf("failed to pop item from key %s: %w", key, err)
	}

	return item, nil
}

// Close stops the cleanup routine and releases resources held by the memoryDB.
func (db *memoryDB) Close() {
	close(db.stopChan) // Signal the cleanup routine to stop
	db.mu.Lock()
	defer db.mu.Unlock()
	db.store = make(map[string]*Item)
}

// startCleanupRoutine starts a goroutine that periodically cleans up expired items from the memory database.
func (db *memoryDB) startCleanupRoutine() {
	ticker := time.NewTicker(db.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			db.cleanExpired()
		case <-db.stopChan:
			return
		}
	}
}

// cleanExpired removes expired items from the memory database.
func (db *memoryDB) cleanExpired() {
	db.mu.Lock()
	defer db.mu.Unlock()

	for key, item := range db.store {
		if item.isExpired() {
			delete(db.store, key)
		}
	}
}
