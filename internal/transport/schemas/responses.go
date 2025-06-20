package schemas

import (
	"time"
)

// OKResponse represents a standard response structure for successful operations.
type OKResponse struct {
	Message string `json:"message"`
}

// RowResponse represents a response structure for a single row in the memory database.
type RowResponse struct {
	Key       string    `json:"key"`
	Kind      string    `json:"kind"`
	Value     any       `json:"value"`
	TTL       time.Time `json:"ttl"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
