package flowtest

import (
	"context"
	"net/http"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/core/engine"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

// httpRequestTestTimeout bounds a test request end to end. It's shorter than
// the engine's client timeout so a hanging host doesn't tie up the API.
const httpRequestTestTimeout = 15 * time.Second

type FlowTestHandler struct {
	// The client flows use, so a test goes through the same egress proxy and
	// can't reach anything a deployed flow couldn't. Nil when no proxy is
	// configured, which turns testing off.
	httpClient     *http.Client
	appSecretStore store.AppSecretStore
	tokenCrypt     *util.SymmetricCrypt
}

func NewFlowTestHandler(httpClient *http.Client, appSecretStore store.AppSecretStore, tokenCrypt *util.SymmetricCrypt) *FlowTestHandler {
	return &FlowTestHandler{
		httpClient:     httpClient,
		appSecretStore: appSecretStore,
		tokenCrypt:     tokenCrypt,
	}
}

func (h *FlowTestHandler) HandleHTTPRequestTest(c *handler.Context, req wire.FlowHTTPRequestTestRequest) (*wire.FlowHTTPRequestTestResponse, error) {
	// Without an egress proxy nothing stops a request from reaching the
	// host's internal network, and a test would return the response straight
	// to the browser.
	if h.httpClient == nil {
		return nil, handler.ErrForbidden(
			"http_request_test_disabled",
			"Testing requests is turned off on this instance. It needs an egress proxy, set engine.http_proxy_url to enable it.",
		)
	}

	res := flow.RunHTTPRequestTest(c.Context(), flow.HTTPRequestTestOpts{
		Data:    req.Data,
		Values:  req.Values,
		HTTP:    clientHTTPProvider{client: h.httpClient},
		Secret:  engine.NewSecretProvider(c.App.ID, h.appSecretStore, h.tokenCrypt),
		Timeout: httpRequestTestTimeout,
	})

	out := &wire.FlowHTTPRequestTestResponse{
		Method:               res.Method,
		URL:                  res.URL,
		RequestHeaders:       res.RequestHeaders,
		RequestBody:          res.RequestBody,
		RequestBodyTruncated: res.RequestBodyCut,
		Result:               res.Result,
		Error:                res.Error,
		ErrorStage:           res.ErrorStage,
		DurationMS:           res.Duration.Milliseconds(),
	}

	if res.Response != nil {
		out.Response = &wire.FlowHTTPRequestTestResponseData{
			Status:        res.Response.Status,
			StatusCode:    res.Response.StatusCode,
			Headers:       res.Response.Headers,
			Body:          res.ResponseBody,
			BodySize:      res.ResponseBodySize,
			BodyTruncated: res.ResponseBodyCut,
			BodyBinary:    res.ResponseBinary,
		}
	}

	return out, nil
}

type clientHTTPProvider struct {
	client *http.Client
}

func (p clientHTTPProvider) HTTPRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	return p.client.Do(req.WithContext(ctx))
}

// HTTPRequestWithoutRedirects isn't used by HTTP request blocks, but is part
// of the provider.
func (p clientHTTPProvider) HTTPRequestWithoutRedirects(ctx context.Context, req *http.Request) (*http.Response, error) {
	client := *p.client
	client.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return client.Do(req.WithContext(ctx))
}
