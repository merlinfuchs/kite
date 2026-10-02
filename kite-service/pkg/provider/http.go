package provider

import (
	"context"
	"net/http"
)

// HTTPProvider provides access to making arbitrary HTTP requests.
type HTTPProvider interface {
	HTTPRequest(ctx context.Context, req *http.Request) (*http.Response, error)
	// HTTPRequestWithoutRedirects returns redirects as they are, for requests
	// with a credential that mustn't be sent to another host.
	HTTPRequestWithoutRedirects(ctx context.Context, req *http.Request) (*http.Response, error)
}

type MockHTTPprovider struct{}

func (p *MockHTTPprovider) HTTPRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	return nil, nil
}

func (p *MockHTTPprovider) HTTPRequestWithoutRedirects(ctx context.Context, req *http.Request) (*http.Response, error) {
	return nil, nil
}
