package flow

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runHTTPRequestBlock(t *testing.T, data HTTPRequestData, p *codeHTTPProvider) (*FlowContext, error) {
	c := javaScriptTestContext(t, FlowProviders{
		HTTP:   p,
		Secret: &secretTestProvider{values: map[string]string{"TOKEN": "s3cret-t0ken"}},
		Log:    &provider.MockLogProvider{},
	}, 1000, eval.Env{"user": map[string]any{"id": "42"}})
	return c, httpNode(data).Execute(c)
}

func TestHTTPRequestBlock(t *testing.T) {
	p := &codeHTTPProvider{status: 404, body: `{"found":false}`}
	c, err := runHTTPRequestBlock(t, HTTPRequestData{
		URL:      "https://api.example.com/users/{{user.id}}?a=1",
		Method:   "post",
		Headers:  []HTTPRequestDataKeyValue{{Key: "Authorization", Value: "Bearer {{secrets.TOKEN}}"}, {Key: "X-Many", Value: "1"}, {Key: "X-Many", Value: "2"}},
		Query:    []HTTPRequestDataKeyValue{{Key: "b", Value: "{{user.id}}"}},
		BodyJSON: json.RawMessage(`{"user":"{{user.id}}"}`),
	}, p)
	require.NoError(t, err)

	req := p.reqs[0]
	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, "https://api.example.com/users/42?a=1&b=42", req.URL.String())
	assert.Equal(t, "Bearer s3cret-t0ken", req.Header.Get("Authorization"))
	assert.Equal(t, []string{"1", "2"}, req.Header.Values("X-Many"))
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	assert.JSONEq(t, `{"user":"42"}`, p.reqBody)

	// The result flows read, whatever the status.
	res := c.GetNodeResult("1").HTTPResponse()
	assert.Equal(t, 404, res.StatusCode)
	assert.JSONEq(t, `{"found":false}`, string(res.Body))
	// Blocks are charged when they start, not per request like code.
	assert.Equal(t, 3, c.CreditsUsed())
}

func TestHTTPRequestBlockRejectsOtherSchemes(t *testing.T) {
	p := &codeHTTPProvider{status: 200}
	_, err := runHTTPRequestBlock(t, HTTPRequestData{URL: "file:///etc/passwd", Method: "GET"}, p)
	assert.ErrorContains(t, err, "the URL must start with http:// or https://")
	assert.Empty(t, p.reqs)
}

func TestHTTPRequestBlockResponseLimit(t *testing.T) {
	// Like before it ran through http.request, a long body without a
	// Content-Length is cut off.
	p := &codeHTTPProvider{status: 200, body: strings.Repeat("a", thing.MaxBodySize+1)}
	c, err := runHTTPRequestBlock(t, HTTPRequestData{URL: "https://example.com", Method: "GET"}, p)
	require.NoError(t, err)
	assert.Len(t, c.GetNodeResult("1").HTTPResponse().Body, thing.MaxBodySize)

	// Any method works, also ones the editor doesn't offer.
	_, err = runHTTPRequestBlock(t, HTTPRequestData{URL: "https://example.com", Method: "HEAD"}, p)
	require.NoError(t, err)
	assert.Equal(t, "HEAD", p.reqs[len(p.reqs)-1].Method)
}
