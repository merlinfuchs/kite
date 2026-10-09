package flow

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"net/http"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

// maxErrorBodyPreview bounds how much of a failed response is put into the
// error message.
const maxErrorBodyPreview = 300

// effectiveBodyType resolves the body type of nodes that predate body types.
func (d *HTTPRequestData) effectiveBodyType() HTTPRequestBodyType {
	if d.BodyType != "" {
		return d.BodyType
	}

	if d.hasLegacyJSONBody() {
		return HTTPRequestBodyTypeJSON
	}
	return HTTPRequestBodyTypeNone
}

func (d *HTTPRequestData) hasLegacyJSONBody() bool {
	raw := bytes.TrimSpace(d.BodyJSON)
	return len(raw) > 0 && !bytes.Equal(raw, []byte("null"))
}

// bodyTemplate is the JSON body as a template, which is the legacy body for
// nodes that predate body types.
func (d *HTTPRequestData) bodyTemplate() string {
	if strings.TrimSpace(d.Body) == "" && d.hasLegacyJSONBody() {
		return string(d.BodyJSON)
	}
	return d.Body
}

// templates are the settings of the request that can reference secrets. The
// response transform isn't one of them, as its result ends up in the flow.
func (d *HTTPRequestData) templates() []string {
	templates := []string{d.URL, d.bodyTemplate()}
	for _, kv := range slices.Concat(d.Headers, d.Query) {
		templates = append(templates, kv.Value)
	}
	return templates
}

// executeHTTPRequest builds and sends the request of an HTTP request node and
// returns what becomes the node's result.
func (ctx *FlowContext) executeHTTPRequest(d *HTTPRequestData) (thing.Thing, error) {
	secrets, err := newRequestSecrets(ctx, d.templates()...)
	if err != nil {
		return thing.Null, err
	}

	req, _, err := ctx.buildHTTPRequest(d, secrets)
	if err != nil {
		return thing.Null, err
	}

	resp, err := ctx.sendHTTPRequest(req, secrets)
	if err != nil {
		return thing.Null, err
	}

	if d.FailOnErrorStatus {
		if err := checkHTTPResponseStatus(resp); err != nil {
			return thing.Null, secrets.Redact(err)
		}
	}

	if d.ResponseTransform != "" {
		return ctx.transformHTTPResponse(d.ResponseTransform, resp)
	}

	return thing.NewHTTPResponse(resp), nil
}

// buildHTTPRequest evaluates the settings of a node into a request without
// sending it. The body is also returned on its own so it can be shown to the
// user when the request is tested.
func (ctx *FlowContext) buildHTTPRequest(d *HTTPRequestData, secrets *requestSecrets) (*http.Request, []byte, error) {
	method := d.Method
	if method == "" {
		method = "GET"
	}

	url, err := secrets.EvalTemplate(ctx, d.URL)
	if err != nil {
		return nil, nil, err
	}

	// Bound to the flow context so the execution deadline and an explicit
	// Cancel both abort the request. With a background request a slow or
	// non-responding host pins the goroutine and its connection past the
	// end of the flow, in a pool shared with the Discord API client.
	req, err := http.NewRequestWithContext(ctx, method, url.String(), nil)
	if err != nil {
		return nil, nil, secrets.Redact(err)
	}

	for _, header := range d.Headers {
		value, err := secrets.EvalTemplate(ctx, header.Value)
		if err != nil {
			return nil, nil, err
		}

		req.Header.Add(header.Key, value.String())
	}

	query := req.URL.Query()
	for _, queryParam := range d.Query {
		value, err := secrets.EvalTemplate(ctx, queryParam.Value)
		if err != nil {
			return nil, nil, err
		}

		query.Add(queryParam.Key, value.String())
	}
	req.URL.RawQuery = query.Encode()

	body, err := ctx.buildHTTPRequestBody(d, secrets)
	if err != nil {
		return nil, nil, err
	}
	if body != nil {
		// A Content-Type set in the headers takes precedence, e.g. for
		// APIs that want application/vnd.api+json.
		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}
		req.Body = io.NopCloser(bytes.NewReader(body))
		req.ContentLength = int64(len(body))
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(body)), nil
		}
	}

	return req, body, nil
}

// buildHTTPRequestBody evaluates the body of the request. A nil body means
// the request is sent without one.
func (ctx *FlowContext) buildHTTPRequestBody(d *HTTPRequestData, secrets *requestSecrets) ([]byte, error) {
	switch d.effectiveBodyType() {
	case HTTPRequestBodyTypeNone:
		return nil, nil
	case HTTPRequestBodyTypeJSON:
		body, err := eval.EvalJSONTemplate(ctx, d.bodyTemplate(), secrets.evalCtx)
		if err != nil {
			return nil, secrets.Redact(fmt.Errorf("failed to evaluate JSON body: %w", err))
		}
		return body, nil
	default:
		return nil, fmt.Errorf("unknown body type %q", d.BodyType)
	}
}

// sendHTTPRequest sends a request through the HTTP provider and reads the
// response.
func (ctx *FlowContext) sendHTTPRequest(req *http.Request, secrets *requestSecrets) (thing.HTTPResponseValue, error) {
	resp, err := ctx.HTTP.HTTPRequest(ctx, req)
	if err != nil {
		return thing.HTTPResponseValue{}, secrets.Redact(err)
	}

	// Closed here rather than deferred: the body is fully consumed below, and
	// a defer in the caller would hold the connection for the whole child
	// subtree. NewHTTPResponseValue also returns early without reading when
	// Content-Length is over its cap, so this has to run on the error path
	// too.
	val, err := thing.NewHTTPResponseValue(resp)
	resp.Body.Close()
	if err != nil {
		return thing.HTTPResponseValue{}, fmt.Errorf("failed to create http response value: %w", err)
	}

	return val, nil
}

// checkHTTPResponseStatus returns an error for 4xx and 5xx responses.
func checkHTTPResponseStatus(resp thing.HTTPResponseValue) error {
	if resp.StatusCode < 400 {
		return nil
	}

	preview := strings.TrimSpace(string(resp.Body))
	if len(preview) > maxErrorBodyPreview {
		preview = preview[:maxErrorBodyPreview]
		// Don't cut a multi-byte character in half.
		for len(preview) > 0 && !utf8.ValidString(preview) {
			preview = preview[:len(preview)-1]
		}
		preview += "…"
	}

	if preview == "" {
		return fmt.Errorf("request failed with status %s", resp.Status)
	}
	return fmt.Errorf("request failed with status %s: %s", resp.Status, preview)
}

// transformHTTPResponse runs the response transform expression of a node.
// The expression sees everything a placeholder would, plus the response as
// `response`.
func (ctx *FlowContext) transformHTTPResponse(expression string, resp thing.HTTPResponseValue) (thing.Thing, error) {
	expression = strings.TrimSpace(expression)

	// Accept the expression wrapped in a placeholder too, as that's how
	// expressions are written everywhere else.
	if strings.HasPrefix(expression, "{{") &&
		strings.HasSuffix(expression, "}}") &&
		strings.Count(expression, "{{") == 1 &&
		strings.Count(expression, "}}") == 1 {
		expression = strings.TrimSpace(expression[2 : len(expression)-2])
	}

	if expression == "" {
		return thing.NewHTTPResponse(resp), nil
	}

	// Copied so the response doesn't leak into the rest of the flow.
	env := make(eval.Env, len(ctx.EvalCtx.Env)+1)
	maps.Copy(env, ctx.EvalCtx.Env)
	env["response"] = eval.NewHTTPResponseEnv(resp)

	res, err := eval.Eval(ctx, expression, eval.Context{
		Env:      env,
		Patchers: ctx.EvalCtx.Patchers,
	})
	if err != nil {
		return thing.Null, fmt.Errorf("failed to transform response: %w", err)
	}

	return res, nil
}
