package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
)

const (
	// Clients like Claude say the same in Cache-Control.
	clientCacheExpiry  = 5 * time.Minute
	maxCachedClients   = 1000
	maxClientIDLength  = 2048
	maxClientNameRunes = 100
	// Real documents are below 1 KB.
	maxMetadataSize      = 10 * 1024
	metadataFetchTimeout = 5 * time.Second
)

// Client is an OAuth client, described by the metadata document at the URL
// that's its ID. Kite doesn't register clients, so anyone can be one, and the
// domain of the URL is what says who it is.
type Client struct {
	ID           string
	Name         string
	Domain       string
	RedirectURIs []string
}

// AllowsRedirectURI reports whether the client can have its codes sent to
// the URI. Loopback URIs match with any port, as native clients like Claude
// Code listen on a random one.
func (c *Client) AllowsRedirectURI(uri string) bool {
	u, err := url.Parse(uri)
	if err != nil || !u.IsAbs() || u.User != nil || strings.Contains(uri, "#") {
		return false
	}

	// Others, like javascript:, could run on the consent page that redirects.
	loopback := isLoopback(u.Hostname())
	if u.Scheme != "https" && (u.Scheme != "http" || !loopback) {
		return false
	}

	for _, registered := range c.RedirectURIs {
		if registered == uri {
			return true
		}
		if !loopback {
			continue
		}
		r, err := url.Parse(registered)
		if err == nil && r.User == nil && r.Scheme == u.Scheme && r.Hostname() == u.Hostname() &&
			r.EscapedPath() == u.EscapedPath() && r.RawQuery == u.RawQuery {
			return true
		}
	}
	return false
}

func isLoopback(host string) bool {
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

type ClientFetcher interface {
	// Client returns the client with the ID, or an error that can be shown
	// to the user if it isn't a valid client.
	Client(ctx context.Context, id string) (*Client, error)
}

// MetadataClientFetcher gets clients from their Client ID Metadata Document,
// see https://datatracker.ietf.org/doc/draft-ietf-oauth-client-id-metadata-document.
type MetadataClientFetcher struct {
	httpClient *http.Client

	mu    sync.Mutex
	cache map[string]cachedClient
}

type cachedClient struct {
	client    *Client
	expiresAt time.Time
}

// NewMetadataClientFetcher fetches documents with the transport of the
// client, which should go through the egress proxy, as anyone can give their
// URL.
func NewMetadataClientFetcher(httpClient *http.Client) *MetadataClientFetcher {
	client := &http.Client{
		Timeout: metadataFetchTimeout,
		// The document must be at its URL, a redirect could lead anywhere.
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	if httpClient != nil {
		client.Transport = httpClient.Transport
	}

	return &MetadataClientFetcher{
		httpClient: client,
		cache:      make(map[string]cachedClient),
	}
}

func (f *MetadataClientFetcher) Client(ctx context.Context, id string) (*Client, error) {
	u, err := checkClientID(id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	f.mu.Lock()
	cached, ok := f.cache[id]
	f.mu.Unlock()
	if ok && now.Before(cached.expiresAt) {
		return cached.client, nil
	}

	client, err := f.fetch(ctx, id, strings.ToLower(u.Host))
	if err != nil {
		return nil, err
	}

	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.cache) >= maxCachedClients {
		for key, entry := range f.cache {
			if !now.Before(entry.expiresAt) {
				delete(f.cache, key)
			}
		}
		if len(f.cache) >= maxCachedClients {
			clear(f.cache)
		}
	}
	f.cache[id] = cachedClient{client: client, expiresAt: now.Add(clientCacheExpiry)}

	return client, nil
}

func invalidClient(format string, args ...any) error {
	return handler.ErrBadRequest("invalid_client", fmt.Sprintf(format, args...))
}

func checkClientID(id string) (*url.URL, error) {
	if id == "" {
		return nil, invalidClient("client_id is missing")
	}
	if len(id) > maxClientIDLength {
		return nil, invalidClient("client_id is too long")
	}

	u, err := url.Parse(id)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || strings.Contains(id, "#") {
		return nil, invalidClient("client_id must be the https URL of the client's metadata document")
	}
	// The domain is what users trust, and Unicode ones can look like others,
	// like clаude.ai with a Cyrillic a.
	for _, r := range u.Host {
		if r > unicode.MaxASCII {
			return nil, invalidClient("client_id must have an ASCII domain")
		}
	}
	if u.Path == "" || u.Path == "/" {
		return nil, invalidClient("client_id must have a path")
	}
	for _, segment := range strings.Split(u.Path, "/") {
		if segment == "." || segment == ".." {
			return nil, invalidClient("client_id must not have . or .. in its path")
		}
	}
	return u, nil
}

type clientMetadata struct {
	ClientID                string   `json:"client_id"`
	ClientName              string   `json:"client_name"`
	ClientSecret            string   `json:"client_secret"`
	RedirectURIs            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	TokenEndpointAuthMethod string   `json:"token_endpoint_auth_method"`
}

func (f *MetadataClientFetcher) fetch(ctx context.Context, id string, domain string) (*Client, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, id, nil)
	if err != nil {
		return nil, invalidClient("client_id must be the https URL of the client's metadata document")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		slog.Info(
			"Failed to fetch OAuth client metadata",
			slog.String("client_id", id),
			slog.String("error", err.Error()),
		)
		return nil, invalidClient("failed to fetch the client's metadata document")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, invalidClient("the client's metadata document returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxMetadataSize+1))
	if err != nil {
		return nil, invalidClient("failed to read the client's metadata document")
	}
	if len(body) > maxMetadataSize {
		return nil, invalidClient("the client's metadata document is larger than %d bytes", maxMetadataSize)
	}

	var doc clientMetadata
	if err := json.Unmarshal(body, &doc); err != nil {
		return nil, invalidClient("the client's metadata document isn't valid JSON")
	}

	if doc.ClientID != id {
		return nil, invalidClient("the client_id of the client's metadata document isn't its URL")
	}
	if len(doc.RedirectURIs) == 0 {
		return nil, invalidClient("the client's metadata document has no redirect_uris")
	}
	// Clients with a secret could only keep it in a document anyone can read.
	if doc.ClientSecret != "" || (doc.TokenEndpointAuthMethod != "" && doc.TokenEndpointAuthMethod != "none") {
		return nil, invalidClient("only clients without a secret are supported, with the token_endpoint_auth_method none")
	}
	if len(doc.GrantTypes) > 0 && !slices.Contains(doc.GrantTypes, "authorization_code") {
		return nil, invalidClient("the client's grant_types must include authorization_code")
	}
	if len(doc.ResponseTypes) > 0 && !slices.Contains(doc.ResponseTypes, "code") {
		return nil, invalidClient("the client's response_types must include code")
	}

	return &Client{
		ID:           id,
		Name:         clientName(doc.ClientName, domain),
		Domain:       domain,
		RedirectURIs: doc.RedirectURIs,
	}, nil
}

// clientName is the name to show for a client, which is its domain if it
// has none.
func clientName(name string, domain string) string {
	name = strings.TrimSpace(strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, name))
	if name == "" {
		return domain
	}

	if runes := []rune(name); len(runes) > maxClientNameRunes {
		return string(runes[:maxClientNameRunes])
	}
	return name
}
