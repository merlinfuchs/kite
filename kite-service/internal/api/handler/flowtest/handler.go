package flowtest

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

// httpRequestTestTimeout bounds a test request end to end. It's shorter than
// the engine's client timeout so a hanging host doesn't tie up the API.
const httpRequestTestTimeout = 15 * time.Second

type FlowTestHandler struct {
	// The same client flows use, so a test goes through the same egress
	// proxy and can't reach anything a deployed flow couldn't.
	httpClient *http.Client
}

func NewFlowTestHandler(httpClient *http.Client) *FlowTestHandler {
	return &FlowTestHandler{
		httpClient: httpClient,
	}
}

func (h *FlowTestHandler) HandleHTTPRequestTest(c *handler.Context, req wire.FlowHTTPRequestTestRequest) (*wire.FlowHTTPRequestTestResponse, error) {
	if h.httpClient == nil {
		return nil, handler.ErrServiceUnavailable("http_unavailable", "Testing requests is not available on this instance")
	}

	slog.Debug(
		"Testing HTTP request block",
		slog.String("app_id", c.App.ID),
		slog.String("user_id", c.Session.UserID),
	)

	res := flow.RunHTTPRequestTest(c.Context(), flow.HTTPRequestTestOpts{
		Data:    req.Data,
		Values:  req.Values,
		HTTP:    clientHTTPProvider{client: h.httpClient},
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
			BodySize:      len(res.Response.Body),
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
