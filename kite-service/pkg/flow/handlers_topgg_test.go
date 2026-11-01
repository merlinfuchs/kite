package flow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTopggVoteGet(t *testing.T) {
	httpProvider := &redirectCheckingHTTPProvider{
		body: `{"created_at":"2026-10-30T12:00:00Z","expires_at":"2026-10-31T00:00:00Z","weight":2}`,
	}

	c, err := executeIntegrationBlock(t, FlowNodeTypeActionTopggVoteGet,
		FlowNodeData{UserTarget: "123456789012345678"},
		map[string]string{"topgg": "t0ken"}, httpProvider)
	require.NoError(t, err)

	assert.Equal(t, "https://top.gg/api/v1/projects/@me/votes/123456789012345678?source=discord", httpProvider.req.URL.String())
	assert.Equal(t, "Bearer t0ken", httpProvider.req.Header.Get("Authorization"))

	res := c.GetNodeResult("1").Object()
	assert.Equal(t, "true", res["voted"].String())
	assert.Equal(t, "2026-10-31T00:00:00Z", res["expires_at"].String())
	assert.Equal(t, "2", res["weight"].String())
}

// Top.gg answers 404 for a user without an active vote.
func TestTopggVoteGetNotVoted(t *testing.T) {
	httpProvider := &redirectCheckingHTTPProvider{
		status: 404,
		body:   `{"type":"about:blank","title":"Not Found","status":404,"detail":"No vote found"}`,
	}

	c, err := executeIntegrationBlock(t, FlowNodeTypeActionTopggVoteGet,
		FlowNodeData{UserTarget: "123456789012345678"},
		map[string]string{"topgg": "t0ken"}, httpProvider)
	require.NoError(t, err)

	res := c.GetNodeResult("1").Object()
	assert.Equal(t, "false", res["voted"].String())
	assert.Equal(t, "", res["expires_at"].String())
	assert.Equal(t, "0", res["weight"].String())
}

func TestTopggVoteGetErrors(t *testing.T) {
	_, err := executeIntegrationBlock(t, FlowNodeTypeActionTopggVoteGet,
		FlowNodeData{UserTarget: "123456789012345678"},
		map[string]string{"topgg": "t0ken"}, &redirectCheckingHTTPProvider{status: 401, body: `{"detail":"bad token"}`})
	assert.ErrorContains(t, err, "Top.gg returned Unauthorized")

	_, err = executeIntegrationBlock(t, FlowNodeTypeActionTopggVoteGet,
		FlowNodeData{UserTarget: "not an id"},
		map[string]string{"topgg": "t0ken"}, &redirectCheckingHTTPProvider{})
	assert.ErrorContains(t, err, "user must be an ID")

	_, err = executeIntegrationBlock(t, FlowNodeTypeActionTopggVoteGet,
		FlowNodeData{UserTarget: "123456789012345678"},
		nil, &redirectCheckingHTTPProvider{})
	assert.ErrorContains(t, err, "Top.gg isn't enabled")
}
