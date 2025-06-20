package db_test

import (
	"fmt"
	"log/slog"
	"memorydb/internal/db"
	"strconv"
	"testing"
)

func BenchmarkMemoryDB_Set(b *testing.B) {
	memdb := db.NewMemoryDB(slog.Default())
	defer memdb.Close()

	// Pre-populate the database with some items
	for i := 0; i < 1000; i++ {
		key := "key" + strconv.Itoa(i)
		value := "value" + strconv.Itoa(i)
		err := memdb.Set(key, value)
		if err != nil {
			b.Fatalf("Failed to set initial value: %v", err)
		}
	}
}

func BenchmarkMemoryDB_Get(b *testing.B) {
	m := db.NewMemoryDB(slog.Default())
	defer m.Close()

	// Prepopulate
	for i := 0; i < b.N; i++ {
		key := "key_" + strconv.Itoa(i)
		_ = m.Set(key, fmt.Sprintf("value_%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "key_" + strconv.Itoa(i)
		_, err := m.Get(key)
		if err != nil {
			b.Fatalf("Get failed: %v", err)
		}
	}
}

func BenchmarkMemoryDB_Remove(b *testing.B) {
	m := db.NewMemoryDB(slog.Default())
	defer m.Close()

	// Prepopulate
	for i := 0; i < b.N; i++ {
		key := "key_" + strconv.Itoa(i)
		_ = m.Set(key, fmt.Sprintf("value_%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := "key_" + strconv.Itoa(i)
		_ = m.Remove(key)
	}
}
