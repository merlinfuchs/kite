package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitByUserSeparatesAPITokens(t *testing.T) {
	limit := RateLimitByUser(1, time.Minute)

	// request sends a request as the user, with one of their API tokens if
	// tokenID is set, and returns the status.
	request := func(tokenID string) int {
		h := APIHandler(func(c *Context) error {
			c.Session = &model.Session{UserID: "user"}
			if tokenID != "" {
				c.APIToken = &model.APIToken{ID: tokenID, UserID: "user"}
			}
			return limit(func(c *Context) error { return nil })(c)
		})
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
		return rec.Code
	}

	assert.Equal(t, http.StatusOK, request(""), "first login request")
	assert.Equal(t, http.StatusTooManyRequests, request(""), "second login request")
	assert.Equal(t, http.StatusOK, request("a"), "token requests don't use the login's limit")
	assert.Equal(t, http.StatusTooManyRequests, request("b"), "another token of the user shares the tokens' limit")
}
