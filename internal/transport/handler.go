package transport

import (
	"fmt"
	"log/slog"
	"memorydb/internal/apierrors"
	"memorydb/internal/db"
	"memorydb/internal/transport/schemas"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	logger *slog.Logger
	db     db.DBClient
}

// NewHandler creates a new handler with the provided logger.
func NewHandler(logger *slog.Logger, db db.DBClient) *Handler {
	return &Handler{logger: logger, db: db}
}

// HandleSet sets a value in the database.
func (h *Handler) HandleSet(w http.ResponseWriter, r *http.Request) {
	// decode the request body into a SetRequest object
	var body schemas.SetRowRequest
	if err := decodeJSON(r.Body, &body); err != nil {
		wrapError(w, err)
		return
	}

	// store value in the db
	var opts []db.ItemOptions
	if body.TTL != nil {
		opts = append(opts, db.WithTTL(body.TTL.Duration))
	}
	err := h.db.Set(body.Key, body.Value.Val, opts...)
	if err != nil {
		wrapError(w, fmt.Errorf("failed to set item in db: %w", err))
		return
	}

	writeJSON(w, http.StatusOK, schemas.OKResponse{Message: "ok"})
}

// HandleGet retrieves a value from the database by its key. The key must be provided as a URL parameter.
func (h *Handler) HandleGet(w http.ResponseWriter, r *http.Request) {
	keyParam := chi.URLParam(r, "key")
	if keyParam == "" {
		e := apierrors.ErrURLParamNotFound
		wrapError(w, e)
		return
	}

	// get item from db
	item, err := h.db.Get(keyParam)
	if err != nil {
		wrapError(w, h.wrapDBError(err))
		return
	}

	response := schemas.RowResponse{
		Key:       keyParam,
		Value:     item.Value,
		Kind:      db.MappingDataType[item.Kind],
		TTL:       item.TTL,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}

	writeJSON(w, http.StatusOK, response)
}

// HandleRemove deletes a value from the database by its key. The key must be provided as a URL parameter.
func (h *Handler) HandleRemove(w http.ResponseWriter, r *http.Request) {
	keyParam := chi.URLParam(r, "key")
	if keyParam == "" {
		e := apierrors.ErrURLParamNotFound
		wrapError(w, e)
		return
	}

	// remove item from db
	if err := h.db.Remove(keyParam); err != nil {
		wrapError(w, h.wrapDBError(err))
		return
	}

	writeJSON(w, http.StatusOK, schemas.OKResponse{Message: "ok"})
}

// HandleUpdate updates a value in the database by its key. The key must be provided as a URL parameter.
func (h *Handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	keyParam := chi.URLParam(r, "key")
	if keyParam == "" {
		e := apierrors.ErrURLParamNotFound
		wrapError(w, e)
		return
	}

	// decode the request body into a SetRequest object
	var body schemas.UpdateRowRequest
	if err := decodeJSON(r.Body, &body); err != nil {
		wrapError(w, err)
		return
	}

	// update value in the db
	var opts []db.ItemOptions
	if body.TTL != nil {
		opts = append(opts, db.WithTTL(body.TTL.Duration))
	}
	err := h.db.Update(keyParam, body.Value.Val, opts...)
	if err != nil {
		wrapError(w, h.wrapDBError(err))
		return
	}

	writeJSON(w, http.StatusOK, schemas.OKResponse{Message: "ok"})
}

// HandlePush adds a new value to an existing key in the database. The key must be provided as a URL parameter.
func (h *Handler) HandlePush(w http.ResponseWriter, r *http.Request) {
	keyParam := chi.URLParam(r, "key")
	if keyParam == "" {
		e := apierrors.ErrURLParamNotFound
		wrapError(w, e)
		return
	}

	// decode the request body into a PushRequest object
	var body schemas.PushItemToSliceRequest
	if err := decodeJSON(r.Body, &body); err != nil {
		wrapError(w, err)
		return
	}

	// push value to db
	var opts []db.ItemOptions
	if body.TTL != nil {
		opts = append(opts, db.WithTTL(body.TTL.Duration))
	}
	row, err := h.db.Push(keyParam, body.Value, opts...)
	if err != nil {
		wrapError(w, h.wrapDBError(err))
		return
	}

	response := schemas.RowResponse{
		Key:       keyParam,
		Value:     row.Value,
		Kind:      db.MappingDataType[row.Kind],
		TTL:       row.TTL,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
	writeJSON(w, http.StatusOK, response)
}

func (h *Handler) HandlePop(w http.ResponseWriter, r *http.Request) {
	keyParam := chi.URLParam(r, "key")
	if keyParam == "" {
		e := apierrors.ErrURLParamNotFound
		wrapError(w, e)
		return
	}

	// pop value from db
	row, err := h.db.Pop(keyParam)
	if err != nil {
		wrapError(w, h.wrapDBError(err))
		return
	}

	response := schemas.RowResponse{
		Key:       keyParam,
		Value:     row.Value,
		Kind:      db.MappingDataType[row.Kind],
		TTL:       row.TTL,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}

	writeJSON(w, http.StatusOK, response)
}

// wrapDBError wraps a database error into an API error with appropriate messages.
func (h *Handler) wrapDBError(err error) *apierrors.ApiError {
	dbError, ok := err.(*db.DBerror)
	if !ok {
		e := apierrors.ErrInternalServer
		e.Message = fmt.Sprintf("internal server error: %v", err)
		e.SysMessage = fmt.Sprintf("internal server error: %v", err)
		return e
	}

	switch dbError {
	case db.ErrDataNotFound:
		e := apierrors.ErrItemNotFound
		e.Message = dbError.Message
		e.SysMessage = dbError.SysMessage
		return e
	case db.ErrKeyHasExpired:
		e := apierrors.ErrKeyHasExpired
		e.Message = dbError.Message
		e.SysMessage = dbError.SysMessage
		return e
	default:
		e := apierrors.ErrInternalServer
		e.Message = dbError.Message
		e.SysMessage = dbError.SysMessage
		return e
	}
}
