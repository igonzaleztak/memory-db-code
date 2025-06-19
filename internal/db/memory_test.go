package db_test

import (
	"memorydb/internal/db"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type MemoryDBSuite struct {
	db db.DBClient
	suite.Suite
}

func (suite *MemoryDBSuite) SetupTest() {
	suite.db = db.NewMemoryDB()
}

func (suite *MemoryDBSuite) TeardownTest() {}

func (suite *MemoryDBSuite) TestSetAndGet() {
	values := []struct {
		key           string
		in            any
		out           any
		expectedError bool
	}{
		{"str", "hello", "hello", false},
		{"list", []string{"a", "b"}, []string{"a", "b"}, false},
		{"bad", 123, nil, true},
	}

	for _, v := range values {
		err := suite.db.Set(v.key, v.in)
		if v.expectedError {
			suite.Error(err, "expected error for key: %s", v.key)
		} else {
			suite.NoError(err)
			item, err := suite.db.Get(v.key)
			suite.NoError(err)
			suite.Equal(v.out, item.Value)
		}
	}
}

func (suite *MemoryDBSuite) TestUpdate() {
	values := []struct {
		key           string
		initial       any
		update        any
		out           any
		expectedError bool
	}{
		{"k1", "a", "b", "b", false},
		{"k2", []string{"x"}, []string{"y"}, []string{"y"}, false},
		{"k3", "init", 999, nil, true},
	}

	for _, v := range values {
		_ = suite.db.Set(v.key, v.initial)
		err := suite.db.Update(v.key, v.update)

		if v.expectedError {
			suite.Error(err)
		} else {
			suite.NoError(err)
			item, err := suite.db.Get(v.key)
			suite.NoError(err)
			suite.Equal(v.out, item.Value)
		}
	}
}

func (suite *MemoryDBSuite) TestRemove() {
	values := []struct {
		key           string
		value         any
		expectedError bool
	}{
		{"r1", "toRemove", false},
		{"r2", nil, true}, // not set
	}

	for _, v := range values {
		if !v.expectedError {
			_ = suite.db.Set(v.key, v.value)
		}

		err := suite.db.Remove(v.key)
		if v.expectedError {
			suite.Error(err)
		} else {
			suite.NoError(err)
			item, err := suite.db.Get(v.key)
			suite.Error(err)
			suite.Nil(item)
		}
	}
}

func (suite *MemoryDBSuite) TestPush() {
	_ = suite.db.Set("list", []string{"one"})

	values := []struct {
		key           string
		valuesToPush  []string
		expectedValue []string
		expectedError bool
	}{
		{"list", []string{"two"}, []string{"one", "two"}, false},
		{"missing", []string{"x"}, nil, true},
	}

	for _, v := range values {
		item, err := suite.db.Push(v.key, v.valuesToPush)
		if v.expectedError {
			suite.Error(err)
			suite.Nil(item)
		} else {
			suite.NoError(err)
			suite.NotNil(item)
			suite.Equal(v.expectedValue, item.Value)
		}
	}
}

func (suite *MemoryDBSuite) TestPop() {
	_ = suite.db.Set("list", []string{"a", "b"})

	values := []struct {
		key           string
		expectedValue []string
		expectedError bool
	}{
		{"list", []string{"a"}, false}, // after popping "b"
		{"missing", nil, true},
	}

	for _, v := range values {
		item, err := suite.db.Pop(v.key)
		if v.expectedError {
			suite.Error(err)
			suite.Nil(item)
		} else {
			suite.NoError(err)
			suite.NotNil(item)
			suite.Equal(v.expectedValue, item.Value)
		}
	}
}
func (suite *MemoryDBSuite) TestExpiration() {
	values := []struct {
		key           string
		value         any
		ttl           time.Duration
		sleep         time.Duration
		expectedError bool
	}{
		{"temp1", "short", 500 * time.Millisecond, 1 * time.Second, true},
		{"temp2", "long", 5 * time.Second, 1 * time.Second, false},
	}

	for _, v := range values {
		err := suite.db.Set(v.key, v.value, db.WithTTL(v.ttl))
		suite.NoError(err)

		time.Sleep(v.sleep)

		item, err := suite.db.Get(v.key)
		if v.expectedError {
			suite.Error(err)
			suite.Nil(item)
		} else {
			suite.NoError(err)
			suite.Equal(v.value, item.Value)
		}
	}
}

func TestMemoryDB(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(MemoryDBSuite))
}
