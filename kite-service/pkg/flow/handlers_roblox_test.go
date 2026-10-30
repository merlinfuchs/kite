package flow

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type robloxTestProvider struct {
	users     map[int64]thing.RobloxUserValue
	avatarErr error
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

func (p *robloxTestProvider) UserAvatarURL(ctx context.Context, id int64) (string, error) {
	if p.avatarErr != nil {
		return "", p.avatarErr
	}
	return fmt.Sprintf("https://tr.rbxcdn.com/%d.png", id), nil
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
	assert.Equal(t, "https://tr.rbxcdn.com/156.png", user.RobloxUser().AvatarURL)
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

func TestRobloxUserGetAvatarError(t *testing.T) {
	user, err := executeRobloxUserGetWith(t, RobloxLookupTypeID, "156", func(c *FlowContext) {
		c.Roblox.(*robloxTestProvider).avatarErr = errors.New("rate limited")
	})
	require.NoError(t, err)
	assert.Equal(t, "hi", user.RobloxUser().Description)
	assert.Empty(t, user.RobloxUser().AvatarURL)
}

func executeBloxlinkUserGet(t *testing.T, data FlowNodeData, httpProvider *redirectCheckingHTTPProvider) (thing.Thing, error) {
	c, err := executeIntegrationBlock(t, FlowNodeTypeActionBloxlinkUserGet, data, map[string]string{"bloxlink": "k3y"}, httpProvider)
	return c.GetNodeResult("1"), err
}

func TestBloxlinkUserGet(t *testing.T) {
	for _, body := range []string{`{"robloxID":"156","resolved":{}}`, `{"robloxID":156}`} {
		httpProvider := &redirectCheckingHTTPProvider{body: body}
		user, err := executeBloxlinkUserGet(t, FlowNodeData{UserTarget: "{{ '123' }}", GuildTarget: "456"}, httpProvider)
		require.NoError(t, err)

		assert.Equal(t, "https://api.blox.link/v4/public/guilds/456/discord-to-roblox/123", httpProvider.req.URL.String())
		assert.Equal(t, "k3y", httpProvider.req.Header.Get("Authorization"))
		assert.Equal(t, "builderman", user.RobloxUser().Name)
		assert.Equal(t, "https://tr.rbxcdn.com/156.png", user.RobloxUser().AvatarURL)
	}
}

// Members who haven't verified are null instead of an error, so flows can
// tell them to verify.
func TestBloxlinkUserGetNotLinked(t *testing.T) {
	user, err := executeBloxlinkUserGet(t, FlowNodeData{UserTarget: "123", GuildTarget: "456"},
		&redirectCheckingHTTPProvider{status: 404, body: `{"error":"User not found"}`})
	require.NoError(t, err)
	assert.True(t, user.IsNil())
}

// A verified member whose Roblox account was deleted still has its ID.
func TestBloxlinkUserGetDeletedAccount(t *testing.T) {
	user, err := executeBloxlinkUserGet(t, FlowNodeData{UserTarget: "123", GuildTarget: "456"},
		&redirectCheckingHTTPProvider{body: `{"robloxID":"1"}`})
	require.NoError(t, err)
	assert.Equal(t, int64(1), user.RobloxUser().ID)
	assert.Empty(t, user.RobloxUser().Name)
}

func TestBloxlinkUserGetErrors(t *testing.T) {
	_, err := executeBloxlinkUserGet(t, FlowNodeData{UserTarget: "123", GuildTarget: "456"},
		&redirectCheckingHTTPProvider{status: 400, body: `{"error":"Invalid API Key"}`})
	assert.ErrorContains(t, err, "Invalid API Key")

	_, err = executeBloxlinkUserGet(t, FlowNodeData{UserTarget: "abc", GuildTarget: "456"}, &redirectCheckingHTTPProvider{})
	assert.ErrorContains(t, err, "user must be an ID")

	// Outside of a server, like in DMs.
	httpProvider := &redirectCheckingHTTPProvider{}
	_, err = executeBloxlinkUserGet(t, FlowNodeData{UserTarget: "123"}, httpProvider)
	assert.ErrorContains(t, err, "needs a server")
	assert.Nil(t, httpProvider.req)
}
