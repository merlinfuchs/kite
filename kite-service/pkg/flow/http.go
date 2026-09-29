package flow

import (
	"bytes"
	"fmt"
	"io"
	"maps"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

// maxHTTPRequestBodySize bounds the request body after evaluation. Every
// template is bounded on its own, but form bodies can hold many of them.
const maxHTTPRequestBodySize = thing.MaxBodySize

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

// executeHTTPRequest builds and sends the request of an HTTP request node and
// returns what becomes the node's result.
func (ctx *FlowContext) executeHTTPRequest(d *HTTPRequestData) (thing.Thing, error) {
	req, _, err := ctx.buildHTTPRequest(d)
	if err != nil {
		return thing.Null, err
	}

	resp, err := ctx.sendHTTPRequest(req)
	if err != nil {
		return thing.Null, err
	}

	if d.FailOnErrorStatus {
		if err := checkHTTPResponseStatus(resp); err != nil {
			return thing.Null, err
		}
	}

	if d.ResponseTransform != "" {
		return ctx.transformHTTPResponse(d.ResponseTransform, resp)
	}

	return thing.NewHTTPResponse(resp), nil
}

// buildHTTPRequest evaluates the templates of a node into a request without
// sending it. The body is also returned on its own so it can be shown to the
// user when the request is tested.
func (ctx *FlowContext) buildHTTPRequest(d *HTTPRequestData) (*http.Request, []byte, error) {
	method := d.Method
	if method == "" {
		method = "GET"
	}

	rawURL, err := ctx.EvalTemplate(d.URL)
	if err != nil {
		return nil, nil, err
	}

	// Bound to the flow context so the execution deadline and an explicit
	// Cancel both abort the request. With a background request a slow or
	// non-responding host pins the goroutine and its connection past the
	// end of the flow, in a pool shared with the Discord API client.
	req, err := http.NewRequestWithContext(ctx, method, rawURL.String(), nil)
	if err != nil {
		return nil, nil, err
	}

	for _, header := range d.Headers {
		value, err := ctx.EvalTemplate(header.Value)
		if err != nil {
			return nil, nil, err
		}

		req.Header.Add(header.Key, value.String())
	}

	query := req.URL.Query()
	for _, queryParam := range d.Query {
		value, err := ctx.EvalTemplate(queryParam.Value)
		if err != nil {
			return nil, nil, err
		}

		query.Add(queryParam.Key, value.String())
	}
	req.URL.RawQuery = query.Encode()

	body, contentType, err := ctx.buildHTTPRequestBody(d)
	if err != nil {
		return nil, nil, err
	}
	if body != nil {
		// A Content-Type set in the headers takes precedence, e.g. for
		// APIs that want application/vnd.api+json.
		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", contentType)
		}
		req.Body = io.NopCloser(bytes.NewReader(body))
		req.ContentLength = int64(len(body))
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(body)), nil
		}
	}

	return req, body, nil
}

// sendHTTPRequest sends a request through the HTTP provider and reads the
// response.
func (ctx *FlowContext) sendHTTPRequest(req *http.Request) (thing.HTTPResponseValue, error) {
	resp, err := ctx.HTTP.HTTPRequest(ctx, req)
	if err != nil {
		return thing.HTTPResponseValue{}, err
	}

	// Closed here rather than deferred by the caller: the body is fully
	// consumed below, and holding it would keep the connection for the whole
	// child subtree. NewHTTPResponseValue also returns early without reading
	// when Content-Length is over its cap, so this has to run on the error
	// path too.
	val, err := thing.NewHTTPResponseValue(resp)
	resp.Body.Close()
	if err != nil {
		return thing.HTTPResponseValue{}, fmt.Errorf("failed to create http response value: %w", err)
	}

	return val, nil
}

// buildHTTPRequestBody evaluates the body of the request. A nil body means
// the request is sent without one.
func (ctx *FlowContext) buildHTTPRequestBody(d *HTTPRequestData) (body []byte, contentType string, err error) {
	switch d.effectiveBodyType() {
	case HTTPRequestBodyTypeNone:
		return nil, "", nil
	case HTTPRequestBodyTypeJSON:
		template := d.Body
		if strings.TrimSpace(template) == "" && d.hasLegacyJSONBody() {
			template = string(d.BodyJSON)
		}

		body, err = eval.EvalJSONTemplate(ctx, template, ctx.EvalCtx)
		if err != nil {
			return nil, "", fmt.Errorf("failed to evaluate JSON body: %w", err)
		}
		if body == nil {
			return nil, "", nil
		}
		contentType = "application/json"
	case HTTPRequestBodyTypeText:
		if d.Body == "" {
			return nil, "", nil
		}

		res, err := ctx.EvalTemplateKeepSpace(d.Body)
		if err != nil {
			return nil, "", fmt.Errorf("failed to evaluate text body: %w", err)
		}

		body = []byte(res.String())
		contentType = strings.TrimSpace(d.BodyContentType)
		if contentType == "" {
			contentType = "text/plain; charset=utf-8"
		}
	case HTTPRequestBodyTypeForm:
		values := url.Values{}
		err := ctx.forEachBodyFormField(d, func(key, value string) error {
			values.Add(key, value)
			return nil
		})
		if err != nil {
			return nil, "", err
		}

		body = []byte(values.Encode())
		contentType = "application/x-www-form-urlencoded"
	case HTTPRequestBodyTypeMultipart:
		var buf bytes.Buffer
		w := multipart.NewWriter(&buf)

		err := ctx.forEachBodyFormField(d, func(key, value string) error {
			return w.WriteField(key, value)
		})
		if err != nil {
			return nil, "", err
		}
		if err := w.Close(); err != nil {
			return nil, "", fmt.Errorf("failed to build multipart body: %w", err)
		}

		body = buf.Bytes()
		contentType = w.FormDataContentType()
	default:
		return nil, "", fmt.Errorf("unknown body type %q", d.BodyType)
	}

	if len(body) > maxHTTPRequestBodySize {
		return nil, "", fmt.Errorf(
			"request body is %d bytes, limit is %d", len(body), maxHTTPRequestBodySize,
		)
	}

	return body, contentType, nil
}

func (ctx *FlowContext) forEachBodyFormField(d *HTTPRequestData, fn func(key, value string) error) error {
	for _, field := range d.BodyForm {
		key := strings.TrimSpace(field.Key)
		if key == "" {
			continue
		}

		value, err := ctx.EvalTemplateKeepSpace(field.Value)
		if err != nil {
			return fmt.Errorf("failed to evaluate form field %q: %w", key, err)
		}

		if err := fn(key, value.String()); err != nil {
			return err
		}
	}
	return nil
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
