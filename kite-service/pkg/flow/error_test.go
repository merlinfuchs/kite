package flow

import (
	"context"
	"testing"

	"github.com/diamondburned/arikawa/v3/utils/httputil"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
)

type discordErrorProvider struct {
	provider.MockDiscordProvider
}

func (p *discordErrorProvider) APIRequest(ctx context.Context, req provider.DiscordAPIRequest) ([]byte, error) {
	return nil, &httputil.HTTPError{Status: 403, Code: 50013, Message: "Missing Permissions"}
}

func TestDescribeErrorDiscord(t *testing.T) {
	_, err := executeBlock(t, &discordErrorProvider{}, "action_role_create", `{}`)

	assert.Equal(t, ErrorInfo{
		NodeID:   "1",
		NodeType: "action_role_create",
		Code:     FlowNodeErrorDiscordAPI,
		Details:  &ErrorDetails{DiscordStatus: 403, DiscordCode: 50013},
	}, DescribeError(err))
}

func TestDescribeErrorSetting(t *testing.T) {
	_, err := executeBlock(t, &blockTestProvider{}, "action_role_create", `{"guild_target": "{{''}}"}`)

	assert.Equal(t, ErrorInfo{
		NodeID:   "1",
		NodeType: "action_role_create",
		Code:     FlowNodeErrorInvalidSetting,
		Details:  &ErrorDetails{Field: "guild_target"},
	}, DescribeError(err))
}

func TestDescribeErrorIntegration(t *testing.T) {
	withTestIntegration(t, IntegrationAuth{Type: "query", Name: "api_key"})
	httpProvider := &redirectCheckingHTTPProvider{status: 401, body: "unauthorized"}

	_, err := executeIntegrationBlock(t, "action_test_api_thing_get",
		FlowNodeData{Fields: map[string]any{"thing_id": "abc"}},
		map[string]string{"test_api": "key"}, httpProvider)

	info := DescribeError(err)
	assert.Equal(t, FlowNodeErrorIntegrationAPI, info.Code)
	assert.Equal(t, &ErrorDetails{Integration: "test_api", HTTPStatus: 401}, info.Details)
}

func TestDescribeErrorLimit(t *testing.T) {
	err := traceError(&CompiledFlowNode{ID: "1", Type: FlowNodeTypeActionLog}, &FlowError{
		Code:    FlowNodeErrorMaxOperationsReached,
		Message: "operations limit exceeded: 100",
	})

	assert.Equal(t, ErrorInfo{
		NodeID:   "1",
		NodeType: FlowNodeTypeActionLog,
		Code:     FlowNodeErrorMaxOperationsReached,
	}, DescribeError(err))
}
