package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"memorydb/internal/enums"
	"os"
	"path/filepath"
	"time"
)

// Operation represents a database operation with its command type, timestamp, and associated item.
type Operation struct {
	Command enums.DBCommand `json:"command"`
	Key     string          `json:"key"`
	Time    time.Time       `json:"time"`
	*Item
}

// setupDirectory creates a directory for the database file if it does not exist and returns a file handle to the database log file.
func setupDirectory(dbPath string) (*os.File, error) {
	filename := "test_db.log"

	// create the directory if it does not exist
	if err := os.MkdirAll(dbPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory for database file: %w", err)
	}

	// create or open the database file
	dbFilePath := filepath.Join(dbPath, filename)
	logFile, err := os.OpenFile(dbFilePath, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open database file %s: %w", dbFilePath, err)
	}

	return logFile, nil
}

// logOperation logs a database operation to the log file.
func (db *memoryDB) logOperation(op *Operation) {
	if !db.persistenceEnabled {
		return
	}
	if err := db.logEncoder.Encode(op); err != nil {
		db.logger.Warn("failed to log operation to file", "key", op.Key, "command", op.Command, "error", err)
		return
	}
}

func (db *memoryDB) loadStoredData() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	dbLog := filepath.Join(db.dbPath, "test_db.log")
	fileInfo, err := os.Stat(dbLog)
	if err != nil {
		if os.IsNotExist(err) {
			// If the file does not exist, we can return nil as there is no data to load
			return nil
		}
		return fmt.Errorf("failed to stat database file: %w", err)
	}
	if fileInfo.Size() == 0 {
		// If the file is empty, we can return nil as there is no data to load
		return nil
	}

	// Open the log file for reading
	db.logFile, err = os.Open(dbLog)
	if err != nil {
		return fmt.Errorf("failed to open database file for reading: %w", err)
	}
	defer db.logFile.Close()

	decoder := json.NewDecoder(db.logFile)
	for {
		var op Operation
		if err := decoder.Decode(&op); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("failed to decode operation from database file: %w", err)
		}

		// Reconstruct the item and store it in the memoryDB
		switch op.Command {
		case enums.DBCommandSet:
			db.store[op.Key] = op.Item
		case enums.DBCommandUpdate:
			if item, exists := db.store[op.Key]; exists {
				item.Value = op.Item.Value
				item.UpdatedAt = op.Item.UpdatedAt
			} else {
				return fmt.Errorf("item with key %s not found for update", op.Key)
			}
		case enums.DBCommandRemove:
			if _, exists := db.store[op.Key]; exists {
				delete(db.store, op.Key)
			} else {
				return fmt.Errorf("item with key %s not found for removal", op.Key)
			}
		case enums.DBCommandPush:
			if item, exists := db.store[op.Key]; exists {
				if err := item.pushToSlice(op.UpdatedAt, op.Item.Value.Val.(string)); err != nil {
					return fmt.Errorf("failed to push value to item with key %s: %v", op.Key, err)
				}
			} else {
				return fmt.Errorf("item with key %s not found for push", op.Key)
			}
		case enums.DBCommandPop:
			if item, exists := db.store[op.Key]; exists {
				if err := item.popFromSlice(op.UpdatedAt); err != nil {
					return fmt.Errorf("failed to pop value from item with key %s: %w", op.Key, err)
				}
			} else {
				return fmt.Errorf("item with key %s not found for pop", op.Key)
			}
		default:
			return fmt.Errorf("unknown command %s in operation log", op.Command)
		}
	}

	return nil
}
