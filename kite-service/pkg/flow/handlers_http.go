package flow

import (
	"bytes"
	"io"
	"net/http"
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

	secrets, err := newFlowSecrets(ctx, data.templates()...)
	if err != nil {
		return err
	}
	req, err := buildHTTPRequest(ctx, data, secrets)
	if err != nil {
		return err
	}

	res, err := sendHTTP(ctx, req, secrets, thing.MaxBodySize)
	if err != nil {
		return err
	}
	ctx.StoreNodeResult(n, thing.NewHTTPResponse(res))
	return n.ExecuteChildren(ctx)
}

// templates are the settings of the request that can have placeholders.
func (data *HTTPRequestData) templates() []string {
	templates := []string{data.URL, string(data.BodyJSON)}
	for _, kv := range slices.Concat(data.Headers, data.Query) {
		templates = append(templates, kv.Value)
	}
	return templates
}

// buildHTTPRequest builds the request of a Send API Request block.
func buildHTTPRequest(ctx *FlowContext, data *HTTPRequestData, secrets *flowSecrets) (*http.Request, error) {
	rawURL, err := secrets.EvalTemplate(ctx, data.URL)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	for _, kv := range data.Query {
		value, err := secrets.EvalTemplate(ctx, kv.Value)
		if err != nil {
			return nil, err
		}
		query.Add(kv.Key, value.String())
	}

	var body io.Reader
	if data.BodyJSON != nil {
		value, err := secrets.EvalJSONTemplate(ctx, data.BodyJSON)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(value)
	}

	req, err := newHTTPRequest(ctx, data.Method, rawURL.String(), query, body)
	if err != nil {
		return nil, secrets.Redact(err)
	}
	for _, kv := range data.Headers {
		value, err := secrets.EvalTemplate(ctx, kv.Value)
		if err != nil {
			return nil, err
		}
		req.Header.Add(kv.Key, value.String())
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}
