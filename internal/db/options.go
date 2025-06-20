package db

import (
	"encoding/json"
	"time"
)

// DBOptions defines an interface for applying options to the memoryDB instance.
//
// It enablesa a flexible configuration of the database, such as setting cleanup intervals or enabling persistence.
// This functional options pattern has been made following uber-go's best practices for configuring components.
type DBOptions interface {
	apply(*memoryDB)
}

// WithCleanupInterval sets the interval for the cleanup routine.
type WithCleanupInterval time.Duration

func (o WithCleanupInterval) apply(db *memoryDB) {
	db.cleanupInterval = time.Duration(o)
}

// WithPersistenceEnabled sets whether persistence is enabled for the database.
type WithPersistenceEnabled string

func (o WithPersistenceEnabled) apply(db *memoryDB) {
	db.persistenceEnabled = true
	db.dbPath = string(o)

	logFile, err := setupDirectory(db.dbPath)
	if err != nil {
		panic("failed to set up directory for persistence: " + err.Error())
	}
	db.logFile = logFile
	db.logEncoder = json.NewEncoder(db.logFile)
}
