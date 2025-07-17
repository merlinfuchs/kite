package provider

import (
	"context"
	"net/http"
)

// HTTPProvider provides access to making arbitrary HTTP requests.
type HTTPProvider interface {
	HTTPRequest(ctx context.Context, req *http.Request) (*http.Response, error)
}

type MockHTTPprovider struct{}

func (p *MockHTTPprovider) HTTPRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	return nil, nil
}
