package godb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"memorydb/internal/transport/schemas"
	"net/http"
	"net/url"
	"time"
)

var (
	_ ApiClient = (*client)(nil)
)

// ApiClient defines the interface for interacting with the in-memory database API.
type ApiClient interface {
	// Get retrieves the value associated with a key from the memory database.
	Get(key string) (*ApiResponse, error)

	// Set stores a key-value pair in the memory database.
	Set(key string, value any, ttl *time.Duration) (*schemas.OKResponse, error)

	// Remove deletes a key-value pair from the memory database.
	Remove(key string) (*schemas.OKResponse, error)

	// Update modifies an existing item in the memory database with the specified key and value.
	Update(key string, value any, ttl *time.Duration) (*schemas.OKResponse, error)

	// Push adds a new item to the memory database with the specified key and value.
	Push(key string, value string, ttl *time.Duration) (*ApiResponse, error)

	// Pop removes the last item from a slice stored at the specified key in the memory database.
	Pop(key string) (*ApiResponse, error)
}

// client is a simple HTTP client for interacting with the memory database.
type client struct {
	url    string       // URL of the memory database server
	prefix string       // Prefix for API endpoints
	client *http.Client // HTTP client for making requests
}

// NewClient creates a new Client instance with the specified URL and a default HTTP client with a timeout.
func NewClient(url string, version string) ApiClient {
	return &client{
		url:    url,
		prefix: "/api/" + version,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Get retrieves the value associated with a key from the memory database.
// It returns a schemas.OKResponse if the operation is successful, or an error if it fails
func (c *client) Get(key string) (*ApiResponse, error) {
	endpoint, err := url.JoinPath(c.url, c.prefix, key)
	if err != nil {
		return nil, fmt.Errorf("failed to join path for key %s: %w", key, err)
	}

	resp, err := c.client.Get(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get item from %s: %w", endpoint, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get item from %s: received status code %d", endpoint, resp.StatusCode)
	}
	defer resp.Body.Close()

	var response ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response from %s: %w", endpoint, err)
	}
	return &response, nil
}

// Set stores a key-value pair in the memory database.
// It returns a schemas.OKResponse if the operation is successful, or an error if it fails
func (c *client) Set(key string, value any, ttl *time.Duration) (*schemas.OKResponse, error) {
	endpoint, err := url.JoinPath(c.url, c.prefix, "set")
	if err != nil {
		return nil, fmt.Errorf("failed to join path for key %s: %w", key, err)
	}

	data := schemas.SetRowRequest{Key: key, Value: value, TTL: ttl}
	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body for %s: %w", endpoint, err)
	}

	resp, err := c.client.Post(endpoint, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to set item in %s: %w", endpoint, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to set item in %s: received status code %d", endpoint, resp.StatusCode)
	}
	defer resp.Body.Close()

	var response schemas.OKResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response from %s: %w", endpoint, err)
	}
	return &response, nil
}

// Remove removes a key-value pair from the memory database.
// It returns a schemas.OKResponse if the operation is successful, or an error if it fails
func (c *client) Remove(key string) (*schemas.OKResponse, error) {
	endpoint, err := url.JoinPath(c.url, c.prefix, key)
	if err != nil {
		return nil, fmt.Errorf("failed to join path for key %s: %w", key, err)
	}
	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create delete request for %s: %w", endpoint, err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to delete item from %s: %w", endpoint, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to delete item from %s: received status code %d", endpoint, resp.StatusCode)
	}
	defer resp.Body.Close()

	var response schemas.OKResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response from %s: %w", endpoint, err)
	}
	return &response, nil
}

// Update updates an existing item in the memory database with the specified key and value.
// It returns a schemas.OKResponse if the operation is successful, or an error if it fails
func (c *client) Update(key string, value any, ttl *time.Duration) (*schemas.OKResponse, error) {
	endpoint, err := url.JoinPath(c.url, c.prefix, key)
	if err != nil {
		return nil, fmt.Errorf("failed to join path for key %s: %w", key, err)
	}

	data := schemas.SetRowRequest{Key: key, Value: value, TTL: ttl}
	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body for %s: %w", endpoint, err)
	}

	req, err := http.NewRequest(http.MethodPatch, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create update request for %s: %w", endpoint, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to update item in %s: %w", endpoint, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to update item in %s: received status code %d", endpoint, resp.StatusCode)
	}
	defer resp.Body.Close()

	var response schemas.OKResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response from %s: %w", endpoint, err)
	}
	return &response, nil
}

// Push adds a new item to the memory database with the specified key and value.
// It returns a schemas.OKResponse if the operation is successful, or an error if it fails
func (c *client) Push(key string, value string, ttl *time.Duration) (*ApiResponse, error) {
	endpoint, err := url.JoinPath(c.url, c.prefix, key, "push")
	if err != nil {
		return nil, fmt.Errorf("failed to join path for key %s: %w", key, err)
	}
	data := schemas.PushItemToSliceRequest{Value: value, TTL: ttl}
	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body for %s: %w", endpoint, err)
	}

	req, err := http.NewRequest(http.MethodPatch, endpoint, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create push request for %s: %w", endpoint, err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to push item to %s: %w", endpoint, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to push item to %s: received status code %d", endpoint, resp.StatusCode)
	}
	defer resp.Body.Close()

	var response ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response from %s: %w", endpoint, err)
	}
	return &response, nil
}

// Pop removes the last item from a slice stored at the specified key in the memory database.
// It returns a schemas.OKResponse if the operation is successful, or an error if it fails
func (c *client) Pop(key string) (*ApiResponse, error) {
	endpoint, err := url.JoinPath(c.url, c.prefix, key, "pop")
	if err != nil {
		return nil, fmt.Errorf("failed to join path for key %s: %w", key, err)
	}
	req, err := http.NewRequest(http.MethodPatch, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create pop request for %s: %w", endpoint, err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to pop item from %s: %w", endpoint, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to pop item from %s: received status code %d", endpoint, resp.StatusCode)
	}
	defer resp.Body.Close()

	var response ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response from %s: %w", endpoint, err)
	}
	return &response, nil
}
