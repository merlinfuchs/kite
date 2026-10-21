package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type samplePayload struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func newTestContext(method, target string, body io.Reader, contentType string) (*Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, target, body)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rec := httptest.NewRecorder()
	return &Context{
		r: req,
		w: rec,
	}, rec
}

func TestContextParseBody_Success(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		expected    samplePayload
	}{
		{
			name:        "standard application/json",
			contentType: "application/json",
			body:        `{"name": "kite", "count": 42}`,
			expected:    samplePayload{Name: "kite", Count: 42},
		},
		{
			name:        "application/json with charset utf-8",
			contentType: "application/json; charset=utf-8",
			body:        `{"name": "kite", "count": 42}`,
			expected:    samplePayload{Name: "kite", Count: 42},
		},
		{
			name:        "case-insensitive content type",
			contentType: "APPLICATION/JSON",
			body:        `{"name": "kite", "count": 42}`,
			expected:    samplePayload{Name: "kite", Count: 42},
		},
		{
			name:        "content type with whitespace and quoted parameter",
			contentType: `application/json; charset="utf-8"`,
			body:        `{"name": "kite", "count": 42}`,
			expected:    samplePayload{Name: "kite", Count: 42},
		},
		{
			name:        "body with trailing whitespace",
			contentType: "application/json",
			body:        "{\"name\": \"kite\", \"count\": 42}  \n\t\r\n",
			expected:    samplePayload{Name: "kite", Count: 42},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, _ := newTestContext(http.MethodPost, "/test", strings.NewReader(tc.body), tc.contentType)

			var payload samplePayload
			err := ctx.ParseBody(&payload)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, payload)
		})
	}
}

func TestContextParseBody_UnsupportedMediaType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
	}{
		{
			name:        "missing content type",
			contentType: "",
			body:        `{"name": "kite"}`,
		},
		{
			name:        "text plain",
			contentType: "text/plain",
			body:        `{"name": "kite"}`,
		},
		{
			name:        "form urlencoded",
			contentType: "application/x-www-form-urlencoded",
			body:        `name=kite`,
		},
		{
			name:        "malformed media type header",
			contentType: "application/json; ;",
			body:        `{"name": "kite"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, _ := newTestContext(http.MethodPost, "/test", strings.NewReader(tc.body), tc.contentType)

			var payload samplePayload
			err := ctx.ParseBody(&payload)
			require.Error(t, err)

			var aerr *Error
			require.True(t, errors.As(err, &aerr), "error should be of type *handler.Error")
			assert.Equal(t, http.StatusUnsupportedMediaType, aerr.Status)
			assert.Equal(t, "unsupported_media_type", aerr.Code)
		})
	}
}

func TestContextParseBody_InvalidJSON(t *testing.T) {
	tests := []struct {
		name            string
		body            string
		expectedMessage string
	}{
		{
			name:            "empty body",
			body:            "",
			expectedMessage: "request body is empty",
		},
		{
			name:            "whitespace only body",
			body:            "   \n\t ",
			expectedMessage: "request body is empty",
		},
		{
			name:            "truncated JSON body",
			body:            `{"name": `,
			expectedMessage: "invalid JSON: unexpected EOF",
		},
		{
			name:            "illegal syntax character",
			body:            `{"name": foo}`,
			expectedMessage: "invalid JSON syntax at offset",
		},
		{
			name:            "type mismatch for field",
			body:            `{"name": "kite", "count": "not-a-number"}`,
			expectedMessage: `invalid value for field "count": expected int`,
		},
		{
			name:            "multiple top level JSON objects",
			body:            `{"name": "kite", "count": 1} {"name": "second", "count": 2}`,
			expectedMessage: "request body must only contain a single JSON value",
		},
		{
			name:            "trailing non-whitespace tokens",
			body:            `{"name": "kite", "count": 1} trailing garbage`,
			expectedMessage: "request body must only contain a single JSON value",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, _ := newTestContext(http.MethodPost, "/test", strings.NewReader(tc.body), "application/json")

			var payload samplePayload
			err := ctx.ParseBody(&payload)
			require.Error(t, err)

			var aerr *Error
			require.True(t, errors.As(err, &aerr), "error should be of type *handler.Error")
			assert.Equal(t, http.StatusBadRequest, aerr.Status)
			assert.Equal(t, "invalid_json", aerr.Code)
			assert.Contains(t, aerr.Message, tc.expectedMessage)
		})
	}
}

func TestContextParseBody_InvalidJSON_TopLevelType(t *testing.T) {
	ctx, _ := newTestContext(http.MethodPost, "/test", strings.NewReader(`"string-instead-of-object"`), "application/json")
	var payload samplePayload
	err := ctx.ParseBody(&payload)
	require.Error(t, err)

	var aerr *Error
	require.True(t, errors.As(err, &aerr))
	assert.Equal(t, http.StatusBadRequest, aerr.Status)
	assert.Equal(t, "invalid_json", aerr.Code)
	assert.Contains(t, aerr.Message, "invalid value: expected")
}

func TestContextParseBody_NonPointerTargetError(t *testing.T) {
	ctx, _ := newTestContext(http.MethodPost, "/test", strings.NewReader(`{"name": "kite"}`), "application/json")
	var payload samplePayload
	err := ctx.ParseBody(payload)
	require.Error(t, err)

	var aerr *Error
	assert.False(t, errors.As(err, &aerr), "developer error should not be wrapped as a client *handler.Error")
	var invalidUnmarshalErr *json.InvalidUnmarshalError
	assert.True(t, errors.As(err, &invalidUnmarshalErr))
}

func TestContextParseBody_ExceedsLimit(t *testing.T) {
	largeBody := strings.Repeat("a", int(MaxJSONBodySize)+1024)
	jsonBody := `{"name": "` + largeBody + `"}`

	ctx, _ := newTestContext(http.MethodPost, "/test", strings.NewReader(jsonBody), "application/json")

	var payload samplePayload
	err := ctx.ParseBody(&payload)
	require.Error(t, err)

	var aerr *Error
	require.True(t, errors.As(err, &aerr), "error should be of type *handler.Error")
	assert.Equal(t, http.StatusBadRequest, aerr.Status)
	assert.Equal(t, "resource_limit", aerr.Code)
}

func TestAPIHandler_WithTypedWithBodyIntegration(t *testing.T) {
	type testReq struct {
		Title string `json:"title"`
	}
	type testResp struct {
		Echo string `json:"echo"`
	}

	handler := APIHandler(TypedWithBody(func(c *Context, r testReq) (*testResp, error) {
		return &testResp{Echo: r.Title}, nil
	}))

	t.Run("valid request returns 200", func(t *testing.T) {
		body := `{"title": "hello"}`
		req := httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var resp wire.APIResponse[*testResp]
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.True(t, resp.Success)
		assert.Equal(t, "hello", resp.Data.Echo)
	})

	t.Run("unsupported media type returns 415", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/items", strings.NewReader(`{"title": "hello"}`))
		req.Header.Set("Content-Type", "text/plain")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnsupportedMediaType, rec.Code)
		var resp wire.APIResponse[struct{}]
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Equal(t, "unsupported_media_type", resp.Error.Code)
	})

	t.Run("malformed json returns 400 invalid_json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/items", bytes.NewReader([]byte(`{"title": `)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		var resp wire.APIResponse[struct{}]
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Equal(t, "invalid_json", resp.Error.Code)
	})
}
