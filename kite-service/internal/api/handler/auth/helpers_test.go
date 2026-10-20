package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsSafeRedirectPath(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"", false},
		{"/", true},
		{"/dashboard", true},
		{"/app/123/settings", true},
		{"/dashboard?tab=overview", true},
		{"//evil.com", false},
		{"//evil.com/path", false},
		{"/\\evil.com", false},
		{"https://evil.com", false},
		{"http://evil.com", false},
		{"javascript:alert(1)", false},
		{"dashboard", false},
		{"relative/path", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			assert.Equal(t, tt.expected, isSafeRedirectPath(tt.path))
		})
	}
}

func TestGetOauthRedirectURL(t *testing.T) {
	h := &AuthHandler{
		config: AuthHandlerConfig{
			AppPublicBaseURL: "https://kite.cloud",
		},
	}

	tests := []struct {
		name         string
		cookiePath   string
		expectedDest string
	}{
		{
			name:         "empty cookie returns base URL",
			cookiePath:   "",
			expectedDest: "https://kite.cloud",
		},
		{
			name:         "valid relative path appended to base",
			cookiePath:   "/dashboard",
			expectedDest: "https://kite.cloud/dashboard",
		},
		{
			name:         "path with query params",
			cookiePath:   "/apps/123?tab=billing",
			expectedDest: "https://kite.cloud/apps/123?tab=billing",
		},
		{
			name:         "protocol-relative redirect rejected",
			cookiePath:   "//evil.com",
			expectedDest: "https://kite.cloud",
		},
		{
			name:         "absolute URL rejected",
			cookiePath:   "https://evil.com",
			expectedDest: "https://kite.cloud",
		},
		{
			name:         "unprefixed path rejected",
			cookiePath:   "dashboard",
			expectedDest: "https://kite.cloud",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dest string
			apiHandler := handler.APIHandler(func(c *handler.Context) error {
				dest = h.getOauthRedirectURL(c)
				return nil
			})

			req := httptest.NewRequest(http.MethodGet, "/v1/auth/login/callback", nil)
			if tt.cookiePath != "" {
				req.AddCookie(&http.Cookie{Name: "oauth_redirect", Value: tt.cookiePath})
			}
			rec := httptest.NewRecorder()
			apiHandler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedDest, dest)
		})
	}
}

func TestSetOauthCookies(t *testing.T) {
	h := &AuthHandler{
		config: AuthHandlerConfig{
			SecureCookies: true,
		},
	}

	t.Run("state cookie attributes", func(t *testing.T) {
		apiHandler := handler.APIHandler(func(c *handler.Context) error {
			h.setOauthStateCookie(c)
			return nil
		})

		req := httptest.NewRequest(http.MethodGet, "/v1/auth/login", nil)
		rec := httptest.NewRecorder()
		apiHandler.ServeHTTP(rec, req)

		var stateCookie *http.Cookie
		for _, c := range rec.Result().Cookies() {
			if c.Name == "oauth_state" {
				stateCookie = c
				break
			}
		}
		require.NotNil(t, stateCookie)
		assert.NotEmpty(t, stateCookie.Value)
		assert.True(t, stateCookie.HttpOnly)
		assert.True(t, stateCookie.Secure)
		assert.Equal(t, http.SameSiteLaxMode, stateCookie.SameSite)
		assert.Equal(t, int((15 * time.Minute).Seconds()), stateCookie.MaxAge)
		assert.Equal(t, "/", stateCookie.Path)
	})

	t.Run("safe redirect cookie attributes", func(t *testing.T) {
		apiHandler := handler.APIHandler(func(c *handler.Context) error {
			h.setOauthRedirectCookie(c)
			return nil
		})

		req := httptest.NewRequest(http.MethodGet, "/v1/auth/login?redirect=/dashboard", nil)
		rec := httptest.NewRecorder()
		apiHandler.ServeHTTP(rec, req)

		var redirectCookie *http.Cookie
		for _, c := range rec.Result().Cookies() {
			if c.Name == "oauth_redirect" {
				redirectCookie = c
				break
			}
		}
		require.NotNil(t, redirectCookie)
		assert.Equal(t, "/dashboard", redirectCookie.Value)
		assert.True(t, redirectCookie.HttpOnly)
		assert.True(t, redirectCookie.Secure)
		assert.Equal(t, http.SameSiteLaxMode, redirectCookie.SameSite)
		assert.Equal(t, int((15 * time.Minute).Seconds()), redirectCookie.MaxAge)
		assert.Equal(t, "/", redirectCookie.Path)
	})

	t.Run("unsafe redirect query ignored", func(t *testing.T) {
		apiHandler := handler.APIHandler(func(c *handler.Context) error {
			h.setOauthRedirectCookie(c)
			return nil
		})

		req := httptest.NewRequest(http.MethodGet, "/v1/auth/login?redirect=//evil.com", nil)
		rec := httptest.NewRecorder()
		apiHandler.ServeHTTP(rec, req)

		for _, c := range rec.Result().Cookies() {
			if c.Name == "oauth_redirect" {
				// Cookie should only be set with MaxAge=-1 to clear it
				assert.Equal(t, -1, c.MaxAge)
			}
		}
	})
}
