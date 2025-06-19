package transport

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"memorydb/internal/apierrors"
	"memorydb/internal/validator"
	"net/http"
)

// writeJSON writes a response in JSON format.
func writeJSON[T any](w http.ResponseWriter, status int, v T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	bytes, err := json.Marshal(v)
	if err != nil {
		wrapError(w, fmt.Errorf("failed to marshal response: %w", err))
		return
	}

	if _, err := w.Write(bytes); err != nil {
		wrapError(w, fmt.Errorf("failed to write response: %w", err))
		return
	}
}

// wrapError wraps an error into an API error and writes it to the response writer.
func wrapError(w http.ResponseWriter, err error) {
	apiError, ok := err.(*apierrors.ApiError)
	if !ok {
		uknownError := apierrors.ErrInternalServer
		slog.Error("API error", "code", uknownError.Code, "message", err.Error(), "status", uknownError.HTTPStatus)
		writeJSON(w, uknownError.HTTPStatus, uknownError)
		return
	}
	slog.Error("API error", "code", apiError.Code, "message", apiError.Message, "status", apiError.HTTPStatus)
	writeJSON(w, apiError.HTTPStatus, apiError)
}

// decodeJSON decodes JSON from the request body into the given object.
func decodeJSON(r io.Reader, v any) error {
	if err := json.NewDecoder(r).Decode(v); err != nil {
		e := apierrors.ErrInvalidJSON
		e.Message = fmt.Sprintf("failed to decode JSON: %v", err)
		e.SysMessage = fmt.Sprintf("failed to decode JSON: %v", err)
		return e
	}

	// validate the object
	if err := validator.ValidateJSON(v); err != nil {
		e := apierrors.ErrInvalidRequest
		e.Message = err.Error()
		e.SysMessage = err.Error()
		return e
	}
	return nil
}
