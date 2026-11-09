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

// newHTTPRequest builds a request of the http.request operation. The Send API
// Request block is a preset of it, and code calls it with fetch. The query is
// added to the URL's own. The method is GET if it's empty.
func newHTTPRequest(c *FlowContext, method string, rawURL string, query url.Values, body io.Reader) (*http.Request, error) {
	method = strings.ToUpper(method)
	if method == "" {
		method = http.MethodGet
	}
	if !slices.Contains(httpRequestMethods, method) {
		return nil, fmt.Errorf("the method must be one of %s", strings.Join(httpRequestMethods, ", "))
	}
	if !isHTTPURL(rawURL) {
		return nil, errors.New("the URL must start with http:// or https://")
	}

	// Bound to the flow context so the execution deadline and an explicit
	// Cancel both abort the request. With a background request a slow or
	// non-responding host pins the goroutine and its connection past the
	// end of the flow, in a pool shared with the Discord API client.
	req, err := http.NewRequestWithContext(c, method, rawURL, body)
	if err != nil {
		return nil, err
	}
	if len(query) > 0 {
		q := req.URL.Query()
		for name, values := range query {
			q[name] = append(q[name], values...)
		}
		req.URL.RawQuery = q.Encode()
	}
	return req, nil
}

// sendHTTP sends the request out of Kite and reads the response, whatever
// its status, with a body of up to maxBody bytes. Test runs don't send it.
// secrets are the secrets its values used, which are removed from errors.
func sendHTTP(c *FlowContext, req *http.Request, secrets *flowSecrets, maxBody int) (thing.HTTPResponseValue, error) {
	if c.Test != nil {
		return thing.HTTPResponseValue{}, errors.New("test runs don't send requests outside of Discord")
	}
	resp, err := c.HTTP.HTTPRequest(c, req)
	if err != nil {
		return thing.HTTPResponseValue{}, secrets.Redact(err)
	}
	defer resp.Body.Close()
	return thing.NewHTTPResponseValue(resp, maxBody)
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

	// Code has no placeholders, only secrets are filled in.
	c := call.ctx
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
	var body io.Reader
	if arg.Body != nil {
		body = strings.NewReader(secrets.Fill(*arg.Body))
	}
	req, err := newHTTPRequest(c, arg.Method, secrets.Fill(arg.URL), nil, body)
	if err != nil {
		return nil, secrets.Redact(err)
	}
	for name, v := range arg.Headers {
		req.Header.Set(name, secrets.Fill(v))
	}
	// Like fetch, which sends text unless the code says otherwise.
	if arg.Body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
	}

	// Like sendHTTP, before the call is charged.
	if c.Test != nil {
		return nil, errors.New("test runs don't send requests outside of Discord")
	}
	if err := call.charge(); err != nil {
		return nil, err
	}
	// The code parses it in its 8 MB of memory.
	res, err := sendHTTP(c, req, secrets, js.MaxInputLength)
	if err != nil {
		return nil, err
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
