package apierrors

import (
	"net/http"
)

// well-known error codes for API responses
var (
	// ErrInternalServer is returned when an internal server error occurs.
	ErrInternalServer = NewAPIError("internal_server_error", "internal server error", http.StatusInternalServerError)

	// ErrInvalidJSON is returned when the JSON is invalid
	ErrInvalidJSON = NewAPIError("invalid_json", "invalid JSON", http.StatusBadRequest)

	// ErrInvalidRequest is returned when the request is invalid.
	ErrInvalidRequest = NewAPIError("invalid_request", "invalid request", http.StatusBadRequest)

	// ErrItemNotFound is returned when an item is not found in the database.
	ErrItemNotFound = NewAPIError("item_not_found", "item not found", http.StatusNotFound)

	// ErrURLParamNotFound is returned when a URL parameter is not found in the request.
	ErrURLParamNotFound = NewAPIError("url_param_not_found", "URL parameter not found", http.StatusBadRequest)

	// ErrKeyHasExpired is returned when a key has expired in the database.
	ErrKeyHasExpired = NewAPIError("key_has_expired", "key has expired", http.StatusGone)
)
