package flow

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSendTestHTTPRequest(t *testing.T) {
	// The server echoes the token, which the editor mustn't see.
	p := &codeHTTPProvider{status: 201, body: `{"token":"s3cret-t0ken"}`}
	res := SendTestHTTPRequest(context.Background(), HTTPRequestData{
		URL:      "https://api.example.com/users/{{ user.id }}?key={{secrets.TOKEN}}",
		Method:   "POST",
		Headers:  []HTTPRequestDataKeyValue{{Key: "Authorization", Value: "Bearer {{secrets.TOKEN}}"}},
		BodyJSON: json.RawMessage(`{"name":"{{user.name}}","token":"{{secrets.TOKEN}}"}`),
	}, map[string]string{"user.id": "42", "user.name": `Jack "the dev"`}, FlowProviders{
		HTTP:   p,
		Secret: &secretTestProvider{values: map[string]string{"TOKEN": "s3cret-t0ken"}},
	})
	require.Empty(t, res.Error)

	// The request itself has the values and secrets.
	req := p.reqs[0]
	assert.Equal(t, "https://api.example.com/users/42?key=s3cret-t0ken", req.URL.String())
	assert.Equal(t, "Bearer s3cret-t0ken", req.Header.Get("Authorization"))
	assert.JSONEq(t, `{"name":"Jack \"the dev\"","token":"s3cret-t0ken"}`, p.reqBody)

	assert.Equal(t, "POST", res.Method)
	assert.Equal(t, "https://api.example.com/users/42?key=[secret]", res.URL)
	assert.JSONEq(t, `{"name":"Jack \"the dev\"","token":"[secret]"}`, res.Body)
	require.NotNil(t, res.Response)
	assert.Equal(t, 201, res.Response.StatusCode)
	assert.JSONEq(t, `{"token":"[secret]"}`, string(res.Response.Body))
}

func TestSendTestHTTPRequestError(t *testing.T) {
	p := &codeHTTPProvider{status: 200}
	res := SendTestHTTPRequest(context.Background(), HTTPRequestData{
		URL:    "ftp://example.com",
		Method: "GET",
	}, nil, FlowProviders{HTTP: p})
	assert.Equal(t, "the URL must start with http:// or https://", res.Error)
	assert.Nil(t, res.Response)
	assert.Empty(t, p.reqs)
}
