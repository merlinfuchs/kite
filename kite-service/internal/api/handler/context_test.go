package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type samplePayload struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

func parseBody(contentType, body string) (samplePayload, error) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	c := &Context{r: req, w: httptest.NewRecorder()}

	var v samplePayload
	err := c.ParseBody(&v)
	return v, err
}

func TestParseBodyAcceptsMediaTypeParams(t *testing.T) {
	for _, ct := range []string{"application/json", "application/json; charset=utf-8", "APPLICATION/JSON"} {
		v, err := parseBody(ct, `{"name": "kite", "count": 42}`)
		require.NoError(t, err, ct)
		assert.Equal(t, samplePayload{Name: "kite", Count: 42}, v)
	}
}

func TestParseBodyErrors(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		body        string
		status      int
		code        string
		message     string
	}{
		{"missing content type", "", `{}`, http.StatusUnsupportedMediaType, "unsupported_media_type", ""},
		{"wrong content type", "text/plain", `{}`, http.StatusUnsupportedMediaType, "unsupported_media_type", ""},
		{"empty body", "application/json", "", http.StatusBadRequest, "invalid_json", ""},
		{"malformed", "application/json", `{"name": foo}`, http.StatusBadRequest, "invalid_json", "invalid JSON at offset"},
		{"trailing value", "application/json", `{"name": "kite"} {}`, http.StatusBadRequest, "invalid_json", "after top-level value"},
		{"type mismatch", "application/json", `{"count": "x"}`, http.StatusBadRequest, "invalid_json", `field "count" has the wrong type (got string)`},
		{"too large", "application/json", `{"name": "` + strings.Repeat("a", MaxJSONBodySize) + `"}`, http.StatusBadRequest, "resource_limit", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseBody(tc.contentType, tc.body)

			var aerr *Error
			require.True(t, errors.As(err, &aerr), "got %v", err)
			assert.Equal(t, tc.status, aerr.Status)
			assert.Equal(t, tc.code, aerr.Code)
			assert.Contains(t, aerr.Message, tc.message)
		})
	}
}

func TestItemChangeSource(t *testing.T) {
	c := &Context{Session: &model.Session{UserID: "user"}}
	assert.Equal(t, store.EditorChange("user", "session"), c.ItemChange("session"))

	c.APIToken = &model.APIToken{ID: "token"}
	assert.Equal(t, store.APITokenChange("user", "token"), c.ItemChange("session"))
}
