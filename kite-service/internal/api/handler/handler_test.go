package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diamondburned/arikawa/v3/utils/httputil"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIHandlerDiscordErrors(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		code   string
	}{
		{
			name:   "client error",
			err:    &httputil.HTTPError{Status: http.StatusForbidden, Code: 50013, Message: "Missing Permissions"},
			status: http.StatusBadRequest,
			code:   "discord_error",
		},
		{
			name:   "server error",
			err:    &httputil.HTTPError{Status: http.StatusBadGateway},
			status: http.StatusInternalServerError,
			code:   "internal_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := APIHandler(func(c *Context) error {
				return fmt.Errorf("failed to send message: %w", tt.err)
			})

			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/", nil))

			var resp wire.APIResponse[struct{}]
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, tt.status, rec.Code)
			assert.Equal(t, tt.code, resp.Error.Code)
			assert.Equal(t, "failed to send message: "+tt.err.Error(), resp.Error.Message)
		})
	}
}
