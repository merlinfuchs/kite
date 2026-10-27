package flow

import (
	"context"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type robloxTestProvider struct {
	users map[int64]thing.RobloxUserValue
}

func (p *robloxTestProvider) UserByID(ctx context.Context, id int64) (*thing.RobloxUserValue, error) {
	user, ok := p.users[id]
	if !ok {
		return nil, provider.ErrNotFound
	}
	return &user, nil
}

// UsersByUsername returns only some fields, like the Roblox API.
func (p *robloxTestProvider) UsersByUsername(ctx context.Context, username string) ([]thing.RobloxUserValue, error) {
	for _, user := range p.users {
		if user.Name == username {
			return []thing.RobloxUserValue{{ID: user.ID, Name: user.Name, DisplayName: user.DisplayName}}, nil
		}
	}
	return nil, nil
}

func executeRobloxUserGet(t *testing.T, mode RobloxLookupType, target string) (thing.Thing, error) {
	return executeRobloxUserGetWith(t, mode, target, func(*FlowContext) {})
}

func executeRobloxUserGetWith(t *testing.T, mode RobloxLookupType, target string, setup func(*FlowContext)) (thing.Thing, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	roblox := &robloxTestProvider{users: map[int64]thing.RobloxUserValue{
		156: {ID: 156, Name: "builderman", DisplayName: "builderman", Description: "hi", CreatedAt: "2006-03-08T00:00:00Z"},
	}}
	c := NewContext(
		ctx,
		5*time.Second,
		&TestContextData{},
		FlowProviders{Roblox: roblox, Log: &provider.MockLogProvider{}},
		FlowContextLimits{MaxStackDepth: 10, MaxOperations: 1000, MaxCredits: 1000},
		eval.NewContext(eval.Env{}),
		nil,
	)
	t.Cleanup(c.Cancel)
	setup(c)

	node := &CompiledFlowNode{
		ID:   "roblox",
		Type: FlowNodeTypeActionRobloxUserGet,
		Data: FlowNodeData{RobloxLookupMode: mode, RobloxUserTarget: target},
	}
	err := node.Execute(c)
	return c.GetNodeResult("roblox"), err
}

func TestRobloxUserGet(t *testing.T) {
	user, err := executeRobloxUserGet(t, RobloxLookupTypeID, " 156 ")
	require.NoError(t, err)
	assert.Equal(t, "hi", user.RobloxUser().Description)

	user, err = executeRobloxUserGet(t, RobloxLookupTypeName, "builderman")
	require.NoError(t, err)
	assert.Equal(t, int64(156), user.RobloxUser().ID)
	assert.Equal(t, "hi", user.RobloxUser().Description)
	assert.Equal(t, "2006-03-08T00:00:00Z", user.RobloxUser().CreatedAt)
}

func TestRobloxUserGetFromResult(t *testing.T) {
	user, err := executeRobloxUserGetWith(t, RobloxLookupTypeID, "{{result('earlier')}}", func(c *FlowContext) {
		c.StoreNodeResult(&CompiledFlowNode{ID: "earlier"}, thing.NewRobloxUser(thing.RobloxUserValue{ID: 156, Name: "builderman"}))
	})
	require.NoError(t, err)
	assert.Equal(t, "hi", user.RobloxUser().Description)
}

func TestRobloxUserGetNotFound(t *testing.T) {
	user, err := executeRobloxUserGet(t, RobloxLookupTypeID, "1")
	require.NoError(t, err)
	assert.True(t, user.IsNil())

	user, err = executeRobloxUserGet(t, RobloxLookupTypeName, "nobody")
	require.NoError(t, err)
	assert.True(t, user.IsNil())
}

func TestRobloxUserGetInvalidTarget(t *testing.T) {
	for _, target := range []string{"", "abc", "<@123>", "0", "-5", "1.5"} {
		_, err := executeRobloxUserGet(t, RobloxLookupTypeID, target)
		assert.ErrorContains(t, err, "must be a Roblox user ID", target)
	}

	_, err := executeRobloxUserGet(t, RobloxLookupTypeName, " ")
	assert.ErrorContains(t, err, "username is empty")
}
