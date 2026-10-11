package flow

import (
	"bytes"
	"io"
	"net/http"
	"slices"

	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

func init() {
	registerHandlers(map[FlowNodeType]nodeHandler{
		FlowNodeTypeActionHTTPRequest: executeActionHTTPRequest,
	})
}

func executeActionHTTPRequest(n *CompiledFlowNode, ctx *FlowContext) error {
	if n.Data.HTTPRequestData == nil {
		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "http_request_data is nil",
		}
	}

	method := n.Data.HTTPRequestData.Method
	if method == "" {
		method = "GET"
	}

	templates := []string{n.Data.HTTPRequestData.URL, string(n.Data.HTTPRequestData.BodyJSON)}
	for _, kv := range slices.Concat(n.Data.HTTPRequestData.Headers, n.Data.HTTPRequestData.Query) {
		templates = append(templates, kv.Value)
	}
	secrets, err := newRequestSecrets(ctx, templates...)
	if err != nil {
		return traceError(n, err)
	}

	url, err := secrets.EvalTemplate(ctx, n.Data.HTTPRequestData.URL)
	if err != nil {
		return traceError(n, err)
	}

	// Bound to the flow context so the execution deadline and an explicit
	// Cancel both abort the request. With a background request a slow or
	// non-responding host pins the goroutine and its connection past the
	// end of the flow, in a pool shared with the Discord API client.
	req, err := http.NewRequestWithContext(ctx, method, url.String(), nil)
	if err != nil {
		return traceError(n, secrets.Redact(err))
	}

	for _, header := range n.Data.HTTPRequestData.Headers {
		value, err := secrets.EvalTemplate(ctx, header.Value)
		if err != nil {
			return traceError(n, err)
		}

		req.Header.Add(header.Key, value.String())
	}

	query := req.URL.Query()
	for _, queryParam := range n.Data.HTTPRequestData.Query {
		value, err := secrets.EvalTemplate(ctx, queryParam.Value)
		if err != nil {
			return traceError(n, err)
		}

		query.Add(queryParam.Key, value.String())
	}
	req.URL.RawQuery = query.Encode()

	if n.Data.HTTPRequestData.BodyJSON != nil {
		// This can potentially break the JSON if an expression returns a string containing double quotes
		// We should probably escape the expression results or only run the eval engine on the actual JSON values
		body, err := secrets.EvalTemplate(ctx, string(n.Data.HTTPRequestData.BodyJSON))
		if err != nil {
			return traceError(n, err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Body = io.NopCloser(bytes.NewReader([]byte(body.String())))
	}

	resp, err := ctx.HTTP.HTTPRequest(ctx, req)
	if err != nil {
		return traceError(n, secrets.Redact(err))
	}

	// Closed here rather than deferred: the body is fully consumed by
	// NewFromHTTPResponse, and a defer would hold the connection for the
	// whole child subtree. NewFromHTTPResponse also returns early without
	// reading when Content-Length is over its cap, so this has to run on
	// the error path too.
	result, err := thing.NewFromHTTPResponse(resp)
	resp.Body.Close()
	if err != nil {
		return traceError(n, err)
	}

	ctx.StoreNodeResult(n, result)
	return n.ExecuteChildren(ctx)
}
