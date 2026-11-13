package flow

import (
	"context"
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

// TestHTTPRequestResult is the request of a Send API Request block sent from
// the editor, see SendTestHTTPRequest. The values of the app's secrets are
// removed from it.
type TestHTTPRequestResult struct {
	Method string
	URL    string
	// Response is nil if no response was received.
	Response *TestHTTPResponse
	Duration time.Duration
	// Error is why the request couldn't be built or sent.
	Error string
}

type TestHTTPResponse struct {
	Status     string
	StatusCode int
	Headers    map[string]string
	// Body is cut off after maxTestHTTPResponseBody bytes, the editor
	// doesn't show more.
	Body          string
	BodyTruncated bool
}

const maxTestHTTPResponseBody = 64 * 1024

// SendTestHTTPRequest sends the request of a Send API Request block on its
// own, like the editor does to try it out. There's no event or earlier block
// to fill in its placeholders, so values has the text of each placeholder by
// its expression, like "user.id". Other placeholders can only use the app's
// secrets. Nothing is charged, the editor limits how often it's used.
func SendTestHTTPRequest(ctx context.Context, data HTTPRequestData, values map[string]string, providers FlowProviders) TestHTTPRequestResult {
	evalCtx := eval.Context{Env: eval.Env{}, Values: values}
	c := NewContext(ctx, Timeout, nil, providers, FlowContextLimits{}, evalCtx, nil)
	defer c.Cancel()

	var res TestHTTPRequestResult
	secrets, err := newFlowSecrets(c, data.templates()...)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	req, err := buildHTTPRequest(c, &data, secrets)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	res.Method = req.Method
	res.URL = secrets.redactString(req.URL.String())

	start := time.Now()
	resp, err := sendHTTP(c, req, secrets, thing.MaxBodySize)
	res.Duration = time.Since(start)
	if err != nil {
		res.Error = err.Error()
		return res
	}

	// Servers can send back what they got, like an echo service. The body is
	// cut off after the secrets are removed, so no part of one is left.
	body := secrets.redactString(string(resp.Body))
	truncated := len(body) > maxTestHTTPResponseBody
	if truncated {
		body = body[:maxTestHTTPResponseBody]
	}
	for name, value := range resp.Headers {
		resp.Headers[name] = secrets.redactString(value)
	}
	res.Response = &TestHTTPResponse{
		Status:        resp.Status,
		StatusCode:    resp.StatusCode,
		Headers:       resp.Headers,
		Body:          body,
		BodyTruncated: truncated,
	}
	return res
}
