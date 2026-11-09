package flow

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

// operationCall is a call of a handler operation from code.
type operationCall struct {
	// A copy of the flow's context that's canceled with the call.
	ctx *FlowContext
	arg json.RawMessage
	// charge counts the call against the flow's operations and charges its
	// credits. A handler calls it once the argument is valid, before it does
	// anything, so a wrong argument isn't billed.
	charge func() error
}

// operationHandler runs a handler operation, see
// kite-web/src/lib/integrations/handlerOperations.ts, and returns its result,
// which the code gets as JSON.
type operationHandler func(call operationCall) (any, error)

// operationHandlers are the handlers of the handler operations, by their
// name, like http.request. TestEveryHandlerOperationHasAHandler checks they
// match operations.json.
var operationHandlers = map[string]operationHandler{
	"http.request": httpRequestOperation,
}

// decodeOperationArg decodes the argument of a call into v, which has a field
// for each of its parameters.
func decodeOperationArg(arg json.RawMessage, v any) error {
	if len(bytes.TrimSpace(arg)) == 0 {
		return nil
	}
	dec := json.NewDecoder(bytes.NewReader(arg))
	dec.DisallowUnknownFields()
	err := dec.Decode(v)
	if err == nil {
		return nil
	}

	var typeErr *json.UnmarshalTypeError
	switch {
	case errors.As(err, &typeErr) && typeErr.Field == "":
		return errors.New("the argument must be an object of its parameters")
	case errors.As(err, &typeErr):
		return fmt.Errorf("invalid value for parameter %s", typeErr.Field)
	case strings.HasPrefix(err.Error(), "json: unknown field "):
		return fmt.Errorf("it has no parameter %s", strings.Trim(strings.TrimPrefix(err.Error(), "json: unknown field "), `"`))
	}
	return errors.New("the argument must be an object of its parameters")
}

// codeSecretRe is a secret in a value of code, like {{secrets.TOKEN}}. Code
// has no placeholders, so other {{...}} stay as they are.
var codeSecretRe = regexp.MustCompile(`\{\{\s*secrets\.([A-Za-z_][A-Za-z0-9_]*)\s*\}\}`)

var httpRequestMethods = []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}

// httpRequestOperation sends a request like the Send API Request block, and
// returns the response whatever its status.
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
	method := strings.ToUpper(arg.Method)
	if method == "" {
		method = http.MethodGet
	}
	if !slices.Contains(httpRequestMethods, method) {
		return nil, fmt.Errorf("invalid value for parameter method: must be one of %s", strings.Join(httpRequestMethods, ", "))
	}

	c := call.ctx
	if c.Test != nil {
		return nil, errors.New("test runs don't send requests outside of Discord")
	}

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
	fill := func(s string) string {
		return codeSecretRe.ReplaceAllStringFunc(s, func(m string) string {
			if v, ok := secrets.values[codeSecretRe.FindStringSubmatch(m)[1]]; ok {
				return v
			}
			return m
		})
	}

	u, err := url.Parse(fill(arg.URL))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return nil, errors.New("invalid value for parameter url: must start with http:// or https://")
	}
	var body io.Reader
	if arg.Body != nil {
		body = strings.NewReader(fill(*arg.Body))
	}
	req, err := http.NewRequestWithContext(c, method, u.String(), body)
	if err != nil {
		return nil, secrets.Redact(err)
	}
	for name, v := range arg.Headers {
		req.Header.Set(name, fill(v))
	}
	// Like fetch, which sends text unless the code says otherwise.
	if arg.Body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
	}

	if err := call.charge(); err != nil {
		return nil, err
	}
	resp, err := c.HTTP.HTTPRequest(c, req)
	if err != nil {
		return nil, secrets.Redact(err)
	}
	res, err := thing.NewHTTPResponseValue(resp)
	resp.Body.Close()
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
