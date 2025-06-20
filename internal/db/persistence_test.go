package db

import (
	"log/slog"
	"memorydb/internal/enums"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type PersistenceSuite struct {
	suite.Suite
}

func (s *PersistenceSuite) TestSetupDirectory() {
	dbPath := ".db"
	file, err := setupDirectory(dbPath)
	s.Require().NoError(err, "Failed to set up directory for persistence")
	s.Require().NotNil(file, "File should not be nil after setup")

	// Clean up the test file after the test
	err = os.RemoveAll(dbPath)
	s.Require().NoError(err, "Failed to remove test database file")
}

func (s *PersistenceSuite) TestLogOperation() {
	db := NewMemoryDB(slog.Default(), WithPersistenceEnabled(".db"))
	defer os.RemoveAll(".db") // Clean up after test

	item := &Item{Value: &StringOrSlice{"testValue"}, Kind: StringType, UpdatedAt: time.Now(), CreatedAt: time.Now()}
	op := &Operation{
		Key:     "testKey",
		Command: enums.DBCommandSet,
		Time:    time.Now(),
		Item:    item,
	}

	db.(*memoryDB).logOperation(op)

	// Verify that the operation was logged correctly
	fileInfo, err := os.Stat(db.(*memoryDB).logFile.Name())
	s.Require().NoError(err, "Failed to get log file info")
	s.Require().Greater(fileInfo.Size(), int64(0), "Log file should not be empty after logging an operation")
}

func (s *PersistenceSuite) TestLoadStoredData() {
	db := NewMemoryDB(slog.Default(), WithPersistenceEnabled(".db"))
	defer os.RemoveAll(".db") // Clean up after test

	s.Run("ok - set string", func() {
		// Create a test item and log an operation
		key := "testKey"
		item := &Item{Value: &StringOrSlice{"testValue"}, Kind: StringType, UpdatedAt: time.Now(), CreatedAt: time.Now()}
		op := &Operation{
			Key:     key,
			Command: enums.DBCommandSet,
			Time:    time.Now(),
			Item:    item,
		}
		db.(*memoryDB).logOperation(op)

		// Load the stored data
		err := db.(*memoryDB).loadStoredData()
		s.Require().NoError(err, "Failed to load stored data")

		// Verify that the item was loaded correctly
		storedItem, exists := db.(*memoryDB).store["testKey"]
		s.Require().True(exists, "Item should exist in the store after loading")
		s.Require().Equal("testValue", storedItem.Value.Val, "Stored item value should match the logged value")
	})

	s.Run("ok - set string slice", func() {
		// Create a test item with a string slice and log an operation
		key := "testSliceKey"
		item := &Item{Value: &StringOrSlice{[]string{"value1", "value2"}}, Kind: StringSliceType, UpdatedAt: time.Now(), CreatedAt: time.Now()}
		op := &Operation{
			Key:     key,
			Command: enums.DBCommandSet,
			Time:    time.Now(),
			Item:    item,
		}
		db.(*memoryDB).logOperation(op)

		// Load the stored data
		err := db.(*memoryDB).loadStoredData()
		s.Require().NoError(err, "Failed to load stored data")

		// Verify that the item was loaded correctly
		storedItem, exists := db.(*memoryDB).store["testSliceKey"]
		s.Require().True(exists, "Item should exist in the store after loading")

		s.Require().Equal([]string{"value1", "value2"}, storedItem.Value.Val, "Stored item value should match the logged value")
	})

	s.Run("ok - remove item", func() {
		// Create a test item and log an operation to remove it
		key := "testRemoveKey"
		item := &Item{Value: &StringOrSlice{"toBeRemoved"}, Kind: StringType, UpdatedAt: time.Now(), CreatedAt: time.Now()}
		op := &Operation{
			Key:     key,
			Command: enums.DBCommandSet,
			Time:    time.Now(),
			Item:    item,
		}
		db.(*memoryDB).logOperation(op)

		// Now log a remove operation
		removeOp := &Operation{
			Key:     key,
			Command: enums.DBCommandRemove,
			Time:    time.Now(),
		}
		db.(*memoryDB).logOperation(removeOp)

		// Load the stored data
		err := db.(*memoryDB).loadStoredData()
		s.Require().NoError(err, "Failed to load stored data")

		// Verify that the item was removed correctly
		_, exists := db.(*memoryDB).store[key]
		s.Require().False(exists, "Item should not exist in the store after removal")
	})

	s.Run("ok - push to slice", func() {
		// Create a test item with a string slice and log an operation to push new values
		key := "testPushKey"
		item := &Item{Value: &StringOrSlice{[]string{"initialValue"}}, Kind: StringSliceType, UpdatedAt: time.Now(), CreatedAt: time.Now()}
		op := &Operation{
			Key:     key,
			Command: enums.DBCommandSet,
			Time:    time.Now(),
			Item:    item,
		}
		db.(*memoryDB).logOperation(op)

		// Now log a push operation
		pushOp := &Operation{
			Key:     key,
			Command: enums.DBCommandPush,
			Time:    time.Now(),
			Item:    &Item{Value: &StringOrSlice{"newValue1"}, UpdatedAt: time.Now()},
		}
		db.(*memoryDB).logOperation(pushOp)

		// Load the stored data
		err := db.(*memoryDB).loadStoredData()
		s.Require().NoError(err, "Failed to load stored data")

		// Verify that the item was updated correctly
		storedItem, exists := db.(*memoryDB).store[key]
		s.Require().True(exists, "Item should exist in the store after loading")
		s.Require().Equal([]string{"initialValue", "newValue1"}, storedItem.Value.Val, "Stored item value should match the logged values after push")
	})

}

func TestPersistence(t *testing.T) {
	suite.Run(t, new(PersistenceSuite))
}
