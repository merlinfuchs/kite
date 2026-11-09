package flow

import (
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
// sendHTTP, with its settings as templates that can use the app's secrets.
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
	u := rawURL.String()
	if len(data.Query) > 0 {
		query := make([][2]string, len(data.Query))
		for i, kv := range data.Query {
			value, err := secrets.EvalTemplate(ctx, kv.Value)
			if err != nil {
				return err
			}
			query[i] = [2]string{kv.Key, value.String()}
		}
		if u, err = withQuery(u, query); err != nil {
			return secrets.Redact(err)
		}
	}

	header := http.Header{}
	for _, kv := range data.Headers {
		value, err := secrets.EvalTemplate(ctx, kv.Value)
		if err != nil {
			return err
		}
		header.Add(kv.Key, value.String())
	}

	var body *string
	if data.BodyJSON != nil {
		// This can potentially break the JSON if an expression returns a string containing double quotes
		// We should probably escape the expression results or only run the eval engine on the actual JSON values
		value, err := secrets.EvalTemplate(ctx, string(data.BodyJSON))
		if err != nil {
			return err
		}
		b := value.String()
		body = &b
		header.Set("Content-Type", "application/json")
	}

	res, err := sendHTTP(httpCall{
		ctx:     ctx,
		method:  data.Method,
		url:     u,
		header:  header,
		body:    body,
		secrets: secrets,
	})
	if err != nil {
		return err
	}

	ctx.StoreNodeResult(n, thing.NewHTTPResponse(res))
	return n.ExecuteChildren(ctx)
}

// withQuery adds the query parameters to the URL.
func withQuery(rawURL string, params [][2]string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	query := u.Query()
	for _, p := range params {
		query.Add(p[0], p[1])
	}
	u.RawQuery = query.Encode()
	return u.String(), nil
}
