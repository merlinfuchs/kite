package oauth

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllowsRedirectURI(t *testing.T) {
	// The redirect URIs of Claude, Claude Code and VS Code.
	claude := &Client{RedirectURIs: []string{"https://claude.ai/api/mcp/auth_callback"}}
	claudeCode := &Client{RedirectURIs: []string{"http://localhost/callback", "http://127.0.0.1/callback"}}
	vscode := &Client{RedirectURIs: []string{"http://127.0.0.1:33418/", "https://vscode.dev/redirect"}}
	javascript := &Client{RedirectURIs: []string{"javascript:alert(1)", "http://example.com/callback"}}

	tests := []struct {
		name   string
		client *Client
		uri    string
		want   bool
	}{
		{"exact", claude, "https://claude.ai/api/mcp/auth_callback", true},
		{"other path", claude, "https://claude.ai/api/mcp/other", false},
		{"other host", claude, "https://evil.com/api/mcp/auth_callback", false},
		{"query added", claude, "https://claude.ai/api/mcp/auth_callback?x=1", false},
		{"fragment", claude, "https://claude.ai/api/mcp/auth_callback#x", false},
		{"userinfo", claude, "https://user@claude.ai/api/mcp/auth_callback", false},
		{"loopback any port", claudeCode, "http://localhost:51234/callback", true},
		{"loopback ip any port", claudeCode, "http://127.0.0.1:8080/callback", true},
		{"loopback other path", claudeCode, "http://localhost:51234/other", false},
		{"loopback other host", claudeCode, "http://[::1]:51234/callback", false},
		{"loopback registered port", vscode, "http://127.0.0.1:40000/", true},
		{"vscode web", vscode, "https://vscode.dev/redirect", true},
		{"javascript", javascript, "javascript:alert(1)", false},
		{"http not loopback", javascript, "http://example.com/callback", false},
		{"relative", claude, "/api/mcp/auth_callback", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.client.AllowsRedirectURI(tt.uri))
		})
	}
}

func TestCheckClientID(t *testing.T) {
	valid := []string{
		"https://claude.ai/oauth/mcp-oauth-client-metadata",
		"https://vscode.dev/oauth/client-metadata.json",
	}
	for _, id := range valid {
		_, err := checkClientID(id)
		assert.NoError(t, err, id)
	}

	invalid := []string{
		"",
		"claude",
		"http://claude.ai/oauth/mcp-oauth-client-metadata",
		"https://claude.ai",
		"https://claude.ai/",
		"https://user@claude.ai/metadata",
		"https://claude.ai/metadata#x",
		"https://claude.ai/oauth/../metadata",
		"https://claude.ai/" + strings.Repeat("a", maxClientIDLength),
	}
	for _, id := range invalid {
		_, err := checkClientID(id)
		assert.Error(t, err, id)
	}
}

// metadataServer serves the document documents returns for its URL.
func metadataServer(t *testing.T, document func(url string) any) (*httptest.Server, *atomic.Int32) {
	t.Helper()

	var fetches atomic.Int32
	var server *httptest.Server
	server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fetches.Add(1)
		switch r.URL.Path {
		case "/redirect":
			http.Redirect(w, r, server.URL+"/client", http.StatusFound)
		case "/missing":
			http.NotFound(w, r)
		case "/large":
			_, _ = w.Write([]byte(`{"client_name":"` + strings.Repeat("a", maxMetadataSize) + `"}`))
		default:
			_ = json.NewEncoder(w).Encode(document(server.URL + r.URL.Path))
		}
	}))
	t.Cleanup(server.Close)
	return server, &fetches
}

func clientError(t *testing.T, err error) string {
	t.Helper()
	var herr *handler.Error
	require.True(t, errors.As(err, &herr), "got %v", err)
	assert.Equal(t, "invalid_client", herr.Code)
	return herr.Message
}

func TestMetadataClientFetcher(t *testing.T) {
	server, fetches := metadataServer(t, func(url string) any {
		doc := map[string]any{
			"client_id":                  url,
			"client_name":                "Test\nClient",
			"redirect_uris":              []string{"https://client.example/callback"},
			"token_endpoint_auth_method": "none",
		}
		switch {
		case strings.HasSuffix(url, "/other-id"):
			doc["client_id"] = "https://evil.example/client"
		case strings.HasSuffix(url, "/secret"):
			doc["token_endpoint_auth_method"] = "client_secret_basic"
		case strings.HasSuffix(url, "/no-redirects"):
			delete(doc, "redirect_uris")
		case strings.HasSuffix(url, "/no-name"):
			delete(doc, "client_name")
		case strings.HasSuffix(url, "/no-code"):
			doc["grant_types"] = []string{"client_credentials"}
		}
		return doc
	})
	fetcher := NewMetadataClientFetcher(server.Client())
	ctx := context.Background()
	domain := strings.TrimPrefix(server.URL, "https://")

	client, err := fetcher.Client(ctx, server.URL+"/client")
	require.NoError(t, err)
	assert.Equal(t, server.URL+"/client", client.ID)
	assert.Equal(t, "TestClient", client.Name)
	assert.Equal(t, domain, client.Domain)
	assert.Equal(t, []string{"https://client.example/callback"}, client.RedirectURIs)

	// It's cached.
	_, err = fetcher.Client(ctx, server.URL+"/client")
	require.NoError(t, err)
	assert.Equal(t, int32(1), fetches.Load())

	client, err = fetcher.Client(ctx, server.URL+"/no-name")
	require.NoError(t, err)
	assert.Equal(t, domain, client.Name)

	for path, want := range map[string]string{
		"/other-id":     "isn't its URL",
		"/secret":       "without a secret",
		"/no-redirects": "no redirect_uris",
		"/no-code":      "authorization_code",
		"/redirect":     "status 302",
		"/missing":      "status 404",
		"/large":        "larger than",
	} {
		_, err := fetcher.Client(ctx, server.URL+path)
		assert.Contains(t, clientError(t, err), want, path)
	}

	// Failures aren't cached.
	before := fetches.Load()
	_, err = fetcher.Client(ctx, server.URL+"/missing")
	clientError(t, err)
	assert.Equal(t, before+1, fetches.Load())
}
