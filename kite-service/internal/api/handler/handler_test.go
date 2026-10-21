package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIHandlerUnexpectedErrorMasked(t *testing.T) {
	sensitiveMsg := "pq: password authentication failed for user \"postgres\" at 10.0.0.5:5432"
	h := APIHandler(func(c *Context) error {
		return errors.New(sensitiveMsg)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/apps", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var resp wire.APIResponse[struct{}]
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.False(t, resp.Success)
	require.NotNil(t, resp.Error)
	assert.Equal(t, "internal_error", resp.Error.Code)
	assert.Equal(t, "An unexpected error occurred. Please try again later.", resp.Error.Message)

	// Ensure no internal database credentials or sensitive details leaked in the body
	assert.NotContains(t, rec.Body.String(), "postgres")
	assert.NotContains(t, rec.Body.String(), "10.0.0.5")
	assert.NotContains(t, rec.Body.String(), "authentication failed")
}

func TestAPIHandlerPreservesExplicitAPIErrors(t *testing.T) {
	t.Run("bad request", func(t *testing.T) {
		h := APIHandler(func(c *Context) error {
			return ErrBadRequest("custom_bad_request", "param is invalid")
		})

		req := httptest.NewRequest(http.MethodPost, "/v1/apps", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)

		var resp wire.APIResponse[struct{}]
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "custom_bad_request", resp.Error.Code)
		assert.Equal(t, "param is invalid", resp.Error.Message)
	})

	t.Run("explicit internal error", func(t *testing.T) {
		h := APIHandler(func(c *Context) error {
			return ErrInternal("custom safe internal message")
		})

		req := httptest.NewRequest(http.MethodGet, "/v1/test", nil)
		rec := httptest.NewRecorder()

		h.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		var resp wire.APIResponse[struct{}]
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)

		assert.False(t, resp.Success)
		require.NotNil(t, resp.Error)
		assert.Equal(t, "internal_error", resp.Error.Code)
		assert.Equal(t, "custom safe internal message", resp.Error.Message)
	})
}

func TestAPIHandlerValidationErrors(t *testing.T) {
	h := APIHandler(func(c *Context) error {
		return validation.Errors{
			"name": errors.New("cannot be blank"),
		}
	})

	req := httptest.NewRequest(http.MethodPost, "/v1/apps", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp wire.APIResponse[struct{}]
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.False(t, resp.Success)
	require.NotNil(t, resp.Error)
	assert.Equal(t, "validation_failed", resp.Error.Code)
}

func TestAPIHandlerObjectStoreDisabled(t *testing.T) {
	h := APIHandler(func(c *Context) error {
		return store.ErrObjectStoreDisabled
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/assets/123", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var resp wire.APIResponse[struct{}]
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.False(t, resp.Success)
	require.NotNil(t, resp.Error)
	assert.Equal(t, "assets_disabled", resp.Error.Code)
}

func TestSanitizeLogPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "standard webhook path with v1",
			path: "/v1/webhooks/app123/listener456/secret789",
			want: "/v1/webhooks/app123/listener456/[REDACTED]",
		},
		{
			name: "webhook path without v1 prefix",
			path: "/webhooks/app123/listener456/secret789",
			want: "/webhooks/app123/listener456/[REDACTED]",
		},
		{
			name: "webhook path with trailing slash",
			path: "/v1/webhooks/app123/listener456/secret789/",
			want: "/v1/webhooks/app123/listener456/[REDACTED]/",
		},
		{
			name: "incomplete webhook path without secret",
			path: "/v1/webhooks/app123",
			want: "/v1/webhooks/app123",
		},
		{
			name: "non-webhook path",
			path: "/v1/apps/app123/status",
			want: "/v1/apps/app123/status",
		},
		{
			name: "root path",
			path: "/",
			want: "/",
		},
		{
			name: "empty path",
			path: "",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeLogPath(tt.path)
			assert.Equal(t, tt.want, got)
		})
	}
}
