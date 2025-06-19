package transport_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"memorydb/internal/apierrors"
	"memorydb/internal/db"
	"memorydb/internal/transport"
	"memorydb/internal/transport/schemas"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/suite"
)

type HandlerSuite struct {
	db      *db.MockDBClient
	handler *transport.Handler
	suite.Suite
}

func (suite *HandlerSuite) SetupTest() {
	db := db.NewMockDBClient(suite.T())
	handler := transport.NewHandler(slog.Default(), db)

	suite.db = db
	suite.handler = handler
}

func (s *HandlerSuite) TestSet() {
	s.Run("Set ok", func() {
		body := schemas.SetRowRequest{
			Key:   "testKey",
			Value: "testValue",
		}

		bodyBytes, err := json.Marshal(body)
		s.Require().NoError(err, "failed to marshal request")

		req := httptest.NewRequest(http.MethodPost, "/api/v1/set", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Mock the database to expect a Set call
		s.db.On("Set", "testKey", "testValue").Return(nil)
		s.handler.HandleSet(w, req)

		resp := w.Result()
		s.Equal(http.StatusOK, resp.StatusCode, "expected status code 200 OK")

		var exptectResponse schemas.OKResponse
		err = json.NewDecoder(resp.Body).Decode(&exptectResponse)
		s.Require().NoError(err, "failed to decode response")
		s.Equal("ok", exptectResponse.Message, "expected response message to be 'ok'")
	})

	s.Run("Set error", func() {
		body := schemas.SetRowRequest{
			Key:   "testKey",
			Value: 123, // Invalid value type
		}

		bodyBytes, err := json.Marshal(body)
		s.Require().NoError(err, "failed to marshal request")

		req := httptest.NewRequest(http.MethodPost, "/api/v1/set", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Mock the database to return an error
		s.db.On("Set", "testKey", "testValue").Return(db.ErrInvalidDataType)

		s.handler.HandleSet(w, req)

		resp := w.Result()
		s.NotEqual(http.StatusOK, resp.StatusCode, "expected non-200 status code")

		var errResponse apierrors.ApiError
		err = json.NewDecoder(resp.Body).Decode(&errResponse)
		s.Require().NoError(err, "failed to decode error response")
		s.Equal(apierrors.ErrInvalidRequest.Code, errResponse.Code, "expected error code to be 'invalid_data_type'")
	})
}

func (s *HandlerSuite) TestGet() {
	s.Run("Get ok", func() {
		key := "testKey"
		s.db.On("Get", key).Return(&db.Item{Value: "testValue"}, nil)

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/%v", key), nil)
		req = withUrlParam(req, "key", key)

		w := httptest.NewRecorder()

		s.handler.HandleGet(w, req)

		resp := w.Result()
		s.Equal(http.StatusOK, resp.StatusCode, "expected status code 200 OK")

		var response schemas.RowResponse
		err := json.NewDecoder(resp.Body).Decode(&response)
		s.Require().NoError(err, "failed to decode response")
		s.Equal("testKey", response.Key, "expected key to be 'testKey'")
		s.Equal("testValue", response.Value, "expected value to be 'testValue'")
	})

	s.Run("Get not found", func() {
		nonExistentKey := "nonExistentKey"
		s.db.On("Get", nonExistentKey).Return(nil, db.ErrDataNotFound)

		endpoint := fmt.Sprintf("/api/v1/%v", nonExistentKey)
		req := httptest.NewRequest(http.MethodGet, endpoint, nil)
		req = withUrlParam(req, "key", nonExistentKey)
		w := httptest.NewRecorder()

		s.handler.HandleGet(w, req)

		resp := w.Result()
		s.Equal(http.StatusNotFound, resp.StatusCode, "expected status code 404 Not Found")

		var errResponse apierrors.ApiError
		err := json.NewDecoder(resp.Body).Decode(&errResponse)
		s.Require().NoError(err, "failed to decode error response")
	})
}

func (s *HandlerSuite) TestRemove() {
	s.Run("Delete ok", func() {
		key := "testKey"
		s.db.On("Remove", key).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/%v", key), nil)
		req = withUrlParam(req, "key", key)
		w := httptest.NewRecorder()

		s.handler.HandleRemove(w, req)

		resp := w.Result()
		s.Equal(http.StatusOK, resp.StatusCode, "expected status code 200 OK")

		var response schemas.OKResponse
		err := json.NewDecoder(resp.Body).Decode(&response)
		s.Require().NoError(err, "failed to decode response")
		s.Equal("ok", response.Message, "expected response message to be 'ok'")
	})
}

func (s *HandlerSuite) TestUpdate() {
	s.Run("Update ok", func() {
		key := "testKey"
		s.db.On("Update", key, "updatedValue").Return(nil)

		body := schemas.SetRowRequest{
			Key:   key,
			Value: "updatedValue",
		}
		bodyBytes, err := json.Marshal(body)
		s.Require().NoError(err, "failed to marshal request")

		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/%v", key), bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req = withUrlParam(req, "key", key)
		w := httptest.NewRecorder()

		// Mock the database to expect an Update call
		s.handler.HandleUpdate(w, req)

		resp := w.Result()
		s.Equal(http.StatusOK, resp.StatusCode, "expected status code 200 OK")
		var response schemas.OKResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		s.Require().NoError(err, "failed to decode response")
		s.Equal("ok", response.Message, "expected response message to be 'ok'")
	})
}

func (s *HandlerSuite) TestPush() {
	s.Run("Push ok", func() {
		key := "testKey"
		newValue := "testNew"
		s.db.On("Push", key, []string{newValue}).Return(&db.Item{Value: []string{"test", newValue}}, nil)
		body := schemas.PushItemToSliceRequest{
			Value: newValue,
		}

		bodyBytes, err := json.Marshal(body)
		s.Require().NoError(err, "failed to marshal request")

		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/%v/push", key), bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req = withUrlParam(req, "key", key)
		w := httptest.NewRecorder()

		// Mock the database to expect a Push call
		s.handler.HandlePush(w, req)

		resp := w.Result()
		s.Equal(http.StatusOK, resp.StatusCode, "expected status code 200 OK")

		var response schemas.RowResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		s.Require().NoError(err, "failed to decode response")
		s.Equal(key, response.Key, "expected key to be 'testKey'")
	})
}

func (s *HandlerSuite) TestPop() {
	s.Run("Pop ok", func() {
		key := "testKey"
		s.db.On("Pop", key).Return(&db.Item{Value: []string{}}, nil)

		req := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/%v/pop", key), nil)
		req = withUrlParam(req, "key", key)
		w := httptest.NewRecorder()

		s.handler.HandlePop(w, req)

		resp := w.Result()
		s.Equal(http.StatusOK, resp.StatusCode, "expected status code 200 OK")

		var response schemas.RowResponse
		err := json.NewDecoder(resp.Body).Decode(&response)
		s.Require().NoError(err, "failed to decode response")
		s.Equal(key, response.Key, "expected key to be 'testKey'")
		s.Empty(response.Value, "expected value to be empty after pop")
	})
}

func TestHandlerSuite(t *testing.T) {
	suite.Run(t, new(HandlerSuite))
}

// withUrlParam returns a pointer to a request object with the given URL params
// added to a new chi.Context object.
func withUrlParam(r *http.Request, key, value string) *http.Request {
	chiCtx := chi.NewRouteContext()
	req := r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, chiCtx))
	chiCtx.URLParams.Add(key, value)
	return req
}
