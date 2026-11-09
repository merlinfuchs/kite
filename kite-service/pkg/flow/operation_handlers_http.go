package flow

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/kitecloud/kite/kite-service/pkg/js"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

var httpRequestMethods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}

// httpCall is a request of the http.request operation. The Send API Request
// block is a preset of it, and code calls it with fetch.
type httpCall struct {
	ctx *FlowContext
	// GET if empty.
	method string
	url    string
	header http.Header
	// No body if nil.
	body *string
	// secrets are the secrets the values used, which are removed from
	// errors.
	secrets *flowSecrets
	// charge is operationCall.charge for code. Blocks are charged when they
	// start, so it's nil for them.
	charge func() error
}

// sendHTTP sends the request and reads the response, whatever its status.
func sendHTTP(c httpCall) (thing.HTTPResponseValue, error) {
	method := strings.ToUpper(c.method)
	if method == "" {
		method = http.MethodGet
	}
	if !slices.Contains(httpRequestMethods, method) {
		return thing.HTTPResponseValue{}, fmt.Errorf("the method must be one of %s", strings.Join(httpRequestMethods, ", "))
	}
	u, err := url.Parse(c.url)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return thing.HTTPResponseValue{}, errors.New("the URL must start with http:// or https://")
	}

	var body io.Reader
	if c.body != nil {
		body = strings.NewReader(*c.body)
	}
	// Bound to the flow context so the execution deadline and an explicit
	// Cancel both abort the request. With a background request a slow or
	// non-responding host pins the goroutine and its connection past the
	// end of the flow, in a pool shared with the Discord API client.
	req, err := http.NewRequestWithContext(c.ctx, method, u.String(), body)
	if err != nil {
		return thing.HTTPResponseValue{}, c.secrets.Redact(err)
	}
	if c.header != nil {
		req.Header = c.header
	}

	if c.charge != nil {
		if err := c.charge(); err != nil {
			return thing.HTTPResponseValue{}, err
		}
	}
	resp, err := c.ctx.HTTP.HTTPRequest(c.ctx, req)
	if err != nil {
		return thing.HTTPResponseValue{}, c.secrets.Redact(err)
	}
	// Closed here rather than deferred: the body is fully read, and a defer
	// would hold the connection for longer. NewHTTPResponseValue also returns
	// early without reading when Content-Length is over its cap, so this has
	// to run on the error path too.
	res, err := thing.NewHTTPResponseValue(resp)
	resp.Body.Close()
	return res, err
}

// httpRequestOperation is http.request for code, which gets the response
// whatever its status.
func httpRequestOperation(call operationCall) (any, error) {
	var arg struct {
		URL     string            `json:"url"`
		Method  string            `json:"method"`
		Headers map[string]string `json:"headers"`
		Body    *string           `json:"body"`
	}
	if err := decodeOperationArg(call.arg, &arg); err != nil {
		return nil, err
	}
	if arg.URL == "" {
		return nil, errors.New("missing parameter url")
	}

	c := call.ctx
	if c.Test != nil {
		return nil, errors.New("test runs don't send requests outside of Discord")
	}

	// Code has no placeholders, only secrets are filled in.
	templates := []string{arg.URL}
	for _, v := range arg.Headers {
		templates = append(templates, v)
	}
	if arg.Body != nil {
		templates = append(templates, *arg.Body)
	}
	secrets, err := newFlowSecrets(c, templates...)
	if err != nil {
		return nil, err
	}
	header := http.Header{}
	for name, v := range arg.Headers {
		header.Set(name, secrets.Fill(v))
	}
	var body *string
	if arg.Body != nil {
		b := secrets.Fill(*arg.Body)
		body = &b
		// Like fetch, which sends text unless the code says otherwise.
		if header.Get("Content-Type") == "" {
			header.Set("Content-Type", "text/plain;charset=UTF-8")
		}
	}

	res, err := sendHTTP(httpCall{
		ctx:     c,
		method:  arg.Method,
		url:     secrets.Fill(arg.URL),
		header:  header,
		body:    body,
		secrets: secrets,
		charge:  call.charge,
	})
	if err != nil {
		return nil, err
	}
	// The code parses it in its 8 MB of memory.
	if len(res.Body) > js.MaxInputLength {
		return nil, fmt.Errorf("the response is larger than %d MB", js.MaxInputLength>>20)
	}

	headers := make(map[string]string, len(res.Headers))
	for name, v := range res.Headers {
		headers[strings.ToLower(name)] = v
	}
	return map[string]any{
		"status":      res.StatusCode,
		"status_text": strings.TrimSpace(strings.TrimPrefix(res.Status, fmt.Sprint(res.StatusCode))),
		"headers":     headers,
		"body":        string(res.Body),
	}, nil
}
