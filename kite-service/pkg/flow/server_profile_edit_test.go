package flow

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type serverProfileDiscordProvider struct {
	provider.MockDiscordProvider

	guildID discord.GuildID
	data    *provider.ServerProfileData
}

func (p *serverProfileDiscordProvider) EditServerProfile(ctx context.Context, guildID discord.GuildID, data provider.ServerProfileData) error {
	p.guildID = guildID
	p.data = &data
	return nil
}

func runServerProfileEdit(t *testing.T, contextData FlowContextData, data FlowNodeData, httpProvider provider.HTTPProvider) (*serverProfileDiscordProvider, error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	discordProvider := &serverProfileDiscordProvider{}

	c := NewContext(
		ctx,
		5*time.Second,
		contextData,
		FlowProviders{
			Discord: discordProvider,
			HTTP:    httpProvider,
			Log:     &provider.MockLogProvider{},
		}, FlowContextLimits{
			MaxStackDepth: 10,
			MaxOperations: 1000,
			MaxCredits:    1000,
		},
		eval.NewContext(eval.Env{}),
		nil,
	)
	defer c.Cancel()

	node := CompiledFlowNode{
		ID:   "0",
		Type: FlowNodeTypeActionServerProfileEdit,
		Data: data,
	}

	return discordProvider, node.Execute(c)
}

func profileJSON(t *testing.T, data *provider.ServerProfileData) string {
	t.Helper()
	b, err := json.Marshal(data)
	require.NoError(t, err)
	return string(b)
}

func ptr(s string) *string {
	return &s
}

// Unset fields stay as they are, and empty ones reset to the bot's own profile.
func TestServerProfileEditSendsOnlySetFields(t *testing.T) {
	p, err := runServerProfileEdit(t, &blockTestContextData{}, FlowNodeData{
		ServerProfileData: &ServerProfileData{
			Nick:      ptr("Helper"),
			AvatarURL: ptr(""),
		},
	}, &provider.MockHTTPprovider{})
	require.NoError(t, err)
	assert.Equal(t, discord.GuildID(5), p.guildID)
	assert.JSONEq(t, `{"nick":"Helper","avatar":null}`, profileJSON(t, p.data))
}

func TestServerProfileEditDownloadsImage(t *testing.T) {
	png := "\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR"
	httpProvider := &redirectCheckingHTTPProvider{body: png}

	p, err := runServerProfileEdit(t, &blockTestContextData{}, FlowNodeData{
		ServerProfileData: &ServerProfileData{BannerURL: ptr("https://example.com/banner.png")},
	}, httpProvider)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/banner.png", httpProvider.req.URL.String())
	assert.JSONEq(t, `{"banner":"data:image/png;base64,iVBORw0KGgoAAAANSUhEUg=="}`, profileJSON(t, p.data))
}

func TestServerProfileEditRejectsNonImage(t *testing.T) {
	p, err := runServerProfileEdit(t, &blockTestContextData{}, FlowNodeData{
		ServerProfileData: &ServerProfileData{AvatarURL: ptr("https://example.com/page")},
	}, &redirectCheckingHTTPProvider{body: "<html></html>"})
	require.ErrorContains(t, err, "image must be a PNG, JPEG, GIF or WebP")
	assert.Nil(t, p.data)
}

func TestServerProfileEditRejectsNonHTTPURL(t *testing.T) {
	_, err := runServerProfileEdit(t, &blockTestContextData{}, FlowNodeData{
		ServerProfileData: &ServerProfileData{AvatarURL: ptr("file:///etc/passwd")},
	}, &redirectCheckingHTTPProvider{})
	require.ErrorContains(t, err, "http:// or https://")
}

func TestServerProfileEditFailsOutsideServer(t *testing.T) {
	p, err := runServerProfileEdit(t, &TestContextData{}, FlowNodeData{
		ServerProfileData: &ServerProfileData{Nick: ptr("Helper")},
	}, &provider.MockHTTPprovider{})
	require.Error(t, err)
	assert.Nil(t, p.data)
}
