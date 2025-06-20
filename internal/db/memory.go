package db

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"memorydb/internal/enums"
	"os"
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

// memoryDB represents an in-memory database that stores items with optional expiration.
type memoryDB struct {
	logger          *slog.Logger     // logger for logging operations
	store           map[string]*Item // in-memory store for items
	cleanupInterval time.Duration    // interval for cleanup routine
	mu              sync.RWMutex     // mutex for preventing race conditions
	stopChan        chan struct{}    // channel to stop the cleanup routine

	// Optional features
	persistenceEnabled bool          // flag to indicate if persistence is enabled
	dbPath             string        // path for persistence storage, if enabled
	logFile            *os.File      // file handle for logging operations, if persistence is enabled
	logEncoder         *json.Encoder // encoder for writing operations to the log file
}

// NewmemoryDB creates a new instance of memoryDB with an initialized store.
func NewMemoryDB(logger *slog.Logger, opts ...DBOptions) DBClient {
	db := &memoryDB{
		logger:          logger,
		store:           make(map[string]*Item),
		cleanupInterval: defaultCleanupInterval,
		stopChan:        make(chan struct{}),
	}

	// Apply options to the memoryDB instance
	for _, opt := range opts {
		opt.apply(db)
	}

	// If persistence is enabled, set up the log file and encoder
	if db.persistenceEnabled {
		if err := db.loadStoredData(); err != nil {
			panic(fmt.Sprintf("failed to load stored data: %v", err))
		}
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

	// log the operation
	db.logOperation(&Operation{
		Command: enums.DBCommandSet,
		Key:     key,
		Time:    time.Now(),
		Item:    itemToStore,
	})

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

	if err := itemToUpdate.update(value, time.Now(), opts...); err != nil {
		return fmt.Errorf("failed to update value for key '%s': %w", key, err)
	}

	db.logOperation(&Operation{
		Command: enums.DBCommandUpdate,
		Key:     key,
		Time:    time.Now(),
		Item:    itemToUpdate,
	})

	return nil
}

// Remove deletes an item from the memory database by its key.
func (db *memoryDB) Remove(key string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if _, exists := db.store[key]; !exists {
		return fmt.Errorf("key %s not found for removal", key)
	}

	delete(db.store, key)

	// log the operation
	db.logOperation(&Operation{
		Command: enums.DBCommandRemove,
		Key:     key,
		Time:    time.Now(),
	})

	return nil
}

// Push adds a new item to the memory database with the specified key and value.
func (db *memoryDB) Push(key string, value string, opts ...ItemOptions) (*Item, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	item, exists := db.store[key]
	if !exists {
		return nil, fmt.Errorf("key %s not found for push", key)
	}

	updatedAt := time.Now()
	if err := item.pushToSlice(updatedAt, value); err != nil {
		return nil, fmt.Errorf("failed to push values to key %s: %w", key, err)
	}

	// log the operation
	db.logOperation(&Operation{
		Command: enums.DBCommandPush,
		Key:     key,
		Time:    updatedAt,
		Item: &Item{
			Value:     &StringOrSlice{value},
			UpdatedAt: updatedAt,
		},
	})

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

	updatedAt := time.Now()
	if err := item.popFromSlice(updatedAt); err != nil {
		return nil, fmt.Errorf("failed to pop item from key %s: %w", key, err)
	}

	// log the operation
	db.logOperation(&Operation{
		Command: enums.DBCommandPop,
		Key:     key,
		Time:    updatedAt,
		Item: &Item{
			UpdatedAt: updatedAt,
		},
	})

	return item, nil
}

// Close stops the cleanup routine and releases resources held by the memoryDB.
func (db *memoryDB) Close() {
	close(db.stopChan) // Signal the cleanup routine to stop
	db.mu.Lock()
	defer db.mu.Unlock()
	db.store = make(map[string]*Item)

	// close the log file if persistence is enabled
	if db.persistenceEnabled {
		if db.logFile != nil {
			db.logFile.Close() // Close the log file if persistence is enabled
		}
	}
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
