package flow

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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

// openCloudTestProvider answers Open Cloud requests by method and path,
// and records them.
type openCloudTestProvider struct {
	responses map[string]string
	requests  []openCloudTestRequest
}

type openCloudTestRequest struct {
	key  string
	body string
}

func (p *openCloudTestProvider) HTTPRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
	return p.HTTPRequestWithoutRedirects(ctx, req)
}

func (p *openCloudTestProvider) HTTPRequestWithoutRedirects(ctx context.Context, req *http.Request) (*http.Response, error) {
	if req.Header.Get("x-api-key") != "k3y" {
		return nil, errors.New("missing API key")
	}
	key := req.Method + " " + strings.TrimPrefix(req.URL.Path, "/cloud/v2/")
	if query, _ := url.QueryUnescape(req.URL.RawQuery); query != "" {
		key += "?" + query
	}
	var body []byte
	if req.Body != nil {
		body, _ = io.ReadAll(req.Body)
	}
	p.requests = append(p.requests, openCloudTestRequest{key: key, body: string(body)})

	res, ok := p.responses[key]
	status := http.StatusOK
	if !ok {
		status, res = http.StatusNotFound, `{"code":"NOT_FOUND","message":"Not found"}`
	}
	return &http.Response{StatusCode: status, Status: http.StatusText(status), Body: io.NopCloser(strings.NewReader(res))}, nil
}

const (
	testMembersPath = "GET groups/7/memberships?maxPageSize=1&filter=user == 'users/156'"
	testRolesPath   = "GET groups/7/roles?maxPageSize=20"
)

func newOpenCloudTestProvider() *openCloudTestProvider {
	return &openCloudTestProvider{responses: map[string]string{
		testMembersPath:                 `{"groupMemberships":[{"path":"groups/7/memberships/abc","user":"users/156","role":"groups/7/roles/30","roles":["groups/7/roles/10"]}]}`,
		testRolesPath:                   `{"groupRoles":[{"path":"groups/7/roles/1","id":"1","displayName":"Guest","rank":0},{"path":"groups/7/roles/10","id":"10","displayName":"Member","rank":1}],"nextPageToken":"p2"}`,
		testRolesPath + "&pageToken=p2": `{"groupRoles":[{"path":"groups/7/roles/20","id":"20","displayName":"Moderator","rank":100},{"path":"groups/7/roles/30","id":"30","displayName":"Admin","rank":200}]}`,
	}}
}

// executeRobloxGroupBlock runs a Roblox group block in group 7, unless data
// has another group.
func executeRobloxGroupBlock(t *testing.T, nodeType FlowNodeType, data FlowNodeData, httpProvider *openCloudTestProvider) (thing.Thing, error) {
	if data.RobloxGroupID == "" {
		data.RobloxGroupID = "7"
	}
	c, err := executeIntegrationBlock(t, nodeType, data, map[string]string{"roblox_open_cloud": "k3y"}, httpProvider)
	return c.GetNodeResult("1"), err
}

func roleNames(member thing.Thing) []string {
	var names []string
	for _, role := range member.Object()["roles"].Array() {
		names = append(names, role.Object()["name"].String())
	}
	return names
}

func TestRobloxGroupMemberGet(t *testing.T) {
	member, err := executeRobloxGroupBlock(t, FlowNodeTypeActionRobloxGroupMemberGet,
		FlowNodeData{RobloxUserTarget: "builderman", RobloxLookupMode: RobloxLookupTypeName}, newOpenCloudTestProvider())
	require.NoError(t, err)

	// The highest role isn't always in the roles Roblox lists.
	assert.Equal(t, []string{"Admin", "Member"}, roleNames(member))
	assert.Equal(t, "Admin", member.Object()["role"].String())
	assert.Equal(t, "200", member.Object()["rank"].String())
	assert.Equal(t, "156", member.Object()["user_id"].String())
}

func TestRobloxGroupMemberGetNotMember(t *testing.T) {
	httpProvider := newOpenCloudTestProvider()
	httpProvider.responses[testMembersPath] = `{"groupMemberships":[]}`
	member, err := executeRobloxGroupBlock(t, FlowNodeTypeActionRobloxGroupMemberGet,
		FlowNodeData{RobloxUserTarget: "156"}, httpProvider)
	require.NoError(t, err)
	assert.True(t, member.IsNil())

	member, err = executeRobloxGroupBlock(t, FlowNodeTypeActionRobloxGroupMemberGet,
		FlowNodeData{RobloxUserTarget: "nobody", RobloxLookupMode: RobloxLookupTypeName}, httpProvider)
	require.NoError(t, err)
	assert.True(t, member.IsNil())
}

func TestRobloxGroupRoleEdit(t *testing.T) {
	tests := []struct {
		name    string
		mode    RobloxGroupRoleMode
		role    string
		request string
		body    string
		roles   []string
	}{
		{"set by name", RobloxGroupRoleModeSet, "moderator", "PATCH groups/7/memberships/abc", `{"role":"groups/7/roles/20"}`, []string{"Moderator"}},
		{"set by rank", "", "100", "PATCH groups/7/memberships/abc", `{"role":"groups/7/roles/20"}`, []string{"Moderator"}},
		{"add by ID", RobloxGroupRoleModeAdd, "20", "POST groups/7/memberships/abc:assignRole", `{"role":"groups/7/roles/20"}`, []string{"Admin", "Moderator", "Member"}},
		{"remove", RobloxGroupRoleModeRemove, "{{ 'Admin' }}", "POST groups/7/memberships/abc:unassignRole", `{"role":"groups/7/roles/30"}`, []string{"Member"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpProvider := newOpenCloudTestProvider()
			httpProvider.responses[tt.request] = `{}`
			member, err := executeRobloxGroupBlock(t, FlowNodeTypeActionRobloxGroupRoleEdit,
				FlowNodeData{RobloxUserTarget: "156", RobloxGroupRole: tt.role, RobloxGroupRoleMode: tt.mode}, httpProvider)
			require.NoError(t, err)

			last := httpProvider.requests[len(httpProvider.requests)-1]
			assert.Equal(t, tt.request, last.key)
			assert.JSONEq(t, tt.body, last.body)
			assert.Equal(t, tt.roles, roleNames(member))
		})
	}
}

// Adding a role the member has, or removing one they don't, only reads.
func TestRobloxGroupRoleEditNoChange(t *testing.T) {
	for _, data := range []FlowNodeData{
		{RobloxUserTarget: "156", RobloxGroupRole: "Member", RobloxGroupRoleMode: RobloxGroupRoleModeAdd},
		{RobloxUserTarget: "156", RobloxGroupRole: "Moderator", RobloxGroupRoleMode: RobloxGroupRoleModeRemove},
	} {
		httpProvider := newOpenCloudTestProvider()
		_, err := executeRobloxGroupBlock(t, FlowNodeTypeActionRobloxGroupRoleEdit, data, httpProvider)
		require.NoError(t, err)
		for _, req := range httpProvider.requests {
			assert.True(t, strings.HasPrefix(req.key, "GET "), req.key)
		}
	}
}

func TestRobloxGroupRoleEditErrors(t *testing.T) {
	httpProvider := newOpenCloudTestProvider()
	_, err := executeRobloxGroupBlock(t, FlowNodeTypeActionRobloxGroupRoleEdit,
		FlowNodeData{RobloxUserTarget: "156", RobloxGroupRole: "Owner"}, httpProvider)
	assert.ErrorContains(t, err, `has no role named, ranked or with the ID "Owner"`)

	// Roblox's error, like when the key's owner can't rank that high.
	_, err = executeRobloxGroupBlock(t, FlowNodeTypeActionRobloxGroupRoleEdit,
		FlowNodeData{RobloxUserTarget: "156", RobloxGroupRole: "Admin"}, httpProvider)
	assert.ErrorContains(t, err, `Roblox Open Cloud returned Not Found: {"code":"NOT_FOUND"`)

	httpProvider.responses[testMembersPath] = `{"groupMemberships":[]}`
	_, err = executeRobloxGroupBlock(t, FlowNodeTypeActionRobloxGroupRoleEdit,
		FlowNodeData{RobloxUserTarget: "156", RobloxGroupRole: "Admin"}, httpProvider)
	assert.ErrorContains(t, err, "isn't in the group")

	_, err = executeRobloxGroupBlock(t, FlowNodeTypeActionRobloxGroupMemberGet,
		FlowNodeData{RobloxGroupID: "groups/7", RobloxUserTarget: "156"}, httpProvider)
	assert.ErrorContains(t, err, "must be a Roblox group ID")
}

func TestRobloxJoinRequestRespond(t *testing.T) {
	const listPath = "GET groups/7/join-requests?maxPageSize=1&filter=user == 'users/156'"
	for _, action := range []RobloxJoinRequestAction{RobloxJoinRequestActionAccept, RobloxJoinRequestActionDecline} {
		httpProvider := newOpenCloudTestProvider()
		httpProvider.responses[listPath] = `{"groupJoinRequests":[{"path":"groups/7/join-requests/156","user":"users/156"}]}`
		httpProvider.responses["POST groups/7/join-requests/156:"+string(action)] = `{}`

		found, err := executeRobloxGroupBlock(t, FlowNodeTypeActionRobloxJoinRequest,
			FlowNodeData{RobloxUserTarget: "156", RobloxJoinRequestAction: action}, httpProvider)
		require.NoError(t, err)
		assert.True(t, found.Bool())
		require.Len(t, httpProvider.requests, 2)
		assert.Equal(t, "POST groups/7/join-requests/156:"+string(action), httpProvider.requests[1].key)
	}

	httpProvider := newOpenCloudTestProvider()
	httpProvider.responses[listPath] = `{"groupJoinRequests":[]}`
	found, err := executeRobloxGroupBlock(t, FlowNodeTypeActionRobloxJoinRequest,
		FlowNodeData{RobloxUserTarget: "156"}, httpProvider)
	require.NoError(t, err)
	assert.False(t, found.Bool())
}
