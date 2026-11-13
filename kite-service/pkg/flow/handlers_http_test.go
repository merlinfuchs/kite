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

func runHTTPRequestBlock(t *testing.T, data HTTPRequestData, p *codeHTTPProvider, env ...eval.Env) (*FlowContext, error) {
	if len(env) == 0 {
		env = append(env, eval.Env{"user": map[string]any{"id": "42"}})
	}
	c := javaScriptTestContext(t, FlowProviders{
		HTTP:   p,
		Secret: &secretTestProvider{values: map[string]string{"TOKEN": "s3cret-t0ken"}},
		Log:    &provider.MockLogProvider{},
	}, 1000, env[0])
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

// Placeholders in the JSON body can't break it, and stay text like they were
// when the whole body was one template.
func TestHTTPRequestBlockJSONBody(t *testing.T) {
	p := &codeHTTPProvider{status: 200}
	_, err := runHTTPRequestBlock(t, HTTPRequestData{
		URL:    "https://api.example.com",
		Method: "POST",
		BodyJSON: json.RawMessage(`{
			"content": "Hi {{user.name}}!",
			"count": "{{user.count}}",
			"missing": "{{user.missing}}",
			"{{user.id}}": [" {{user.id}} ", 1234567890123456789],
			"token": "{{secrets.TOKEN}}"
		}`),
	}, p, eval.Env{"user": map[string]any{"id": "42", "name": "Jack \"the dev\" \\o/\nhi", "count": 3}})
	require.NoError(t, err)

	assert.JSONEq(t, `{
		"content": "Hi Jack \"the dev\" \\o/\nhi!",
		"count": "3",
		"missing": "",
		"42": [" 42 ", 1234567890123456789],
		"token": "s3cret-t0ken"
	}`, p.reqBody)
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
