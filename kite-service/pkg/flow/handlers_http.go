package flow

import (
	"bytes"
	"io"
	"net/url"
	"slices"

	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionHTTPRequest: executeActionHTTPRequest,
	})
}

// executeActionHTTPRequest is a preset of the http.request operation, see
// newHTTPRequest, with its settings as templates that can use the app's
// secrets.
func executeActionHTTPRequest(n *CompiledFlowNode, ctx *FlowContext) error {
	data := n.Data.HTTPRequestData
	if data == nil {
		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "http_request_data is nil",
		}
	}

	templates := []string{data.URL, string(data.BodyJSON)}
	for _, kv := range slices.Concat(data.Headers, data.Query) {
		templates = append(templates, kv.Value)
	}
	secrets, err := newFlowSecrets(ctx, templates...)
	if err != nil {
		return err
	}

	rawURL, err := secrets.EvalTemplate(ctx, data.URL)
	if err != nil {
		return err
	}
	query := url.Values{}
	for _, kv := range data.Query {
		value, err := secrets.EvalTemplate(ctx, kv.Value)
		if err != nil {
			return err
		}
		query.Add(kv.Key, value.String())
	}

	var body io.Reader
	if data.BodyJSON != nil {
		value, err := secrets.EvalJSONTemplate(ctx, data.BodyJSON)
		if err != nil {
			return err
		}
		body = bytes.NewReader(value)
	}

	req, err := newHTTPRequest(ctx, data.Method, rawURL.String(), query, body)
	if err != nil {
		return secrets.Redact(err)
	}
	for _, kv := range data.Headers {
		value, err := secrets.EvalTemplate(ctx, kv.Value)
		if err != nil {
			return err
		}
		req.Header.Add(kv.Key, value.String())
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := sendHTTP(ctx, req, secrets, thing.MaxBodySize)
	if err != nil {
		return err
	}
	ctx.StoreNodeResult(n, thing.NewHTTPResponse(res))
	return n.ExecuteChildren(ctx)
}
