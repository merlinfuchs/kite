package app

import (
	"context"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemapFlow(t *testing.T) {
	ids := &duplicateIDs{
		variables: map[string]string{"var1": "var2"},
		messages:  map[string]string{"msg1": "msg2"},
		assets:    map[string]string{},
	}

	source := flow.FlowData{
		Nodes: []flow.FlowNode{
			{ID: "a", Data: flow.FlowNodeData{VariableID: "var1"}},
			{ID: "b", Data: flow.FlowNodeData{MessageTemplateID: "msg1"}},
			{ID: "c", Data: flow.FlowNodeData{VariableID: "unknown", MessageTemplateID: "other"}},
		},
		Edges: []flow.FlowEdge{{ID: "e", Source: "a", Target: "b"}},
	}

	h := &AppHandler{}
	res, err := h.remapFlow(context.Background(), ids, source, "app", "user")
	require.NoError(t, err)

	assert.Equal(t, "var2", res.Nodes[0].Data.VariableID)
	assert.Equal(t, "msg2", res.Nodes[1].Data.MessageTemplateID)
	assert.Equal(t, "unknown", res.Nodes[2].Data.VariableID)
	assert.Equal(t, "other", res.Nodes[2].Data.MessageTemplateID)
	assert.Equal(t, source.Edges, res.Edges)

	// The original flow is left as it was.
	assert.Equal(t, "var1", source.Nodes[0].Data.VariableID)
	assert.Equal(t, "msg1", source.Nodes[1].Data.MessageTemplateID)
}

func TestCheckContentLimits(t *testing.T) {
	features := model.Features{
		MaxCommands:                2,
		MaxEventListeners:          1,
		MaxScheduledEventListeners: 1,
	}

	listener := func(source model.EventSource) *model.EventListener {
		return &model.EventListener{Source: source}
	}

	tests := []struct {
		name    string
		content appContent
		wantErr bool
	}{
		{"empty", appContent{}, false},
		{"at limit", appContent{commands: make([]*model.Command, 2)}, false},
		{"too many commands", appContent{commands: make([]*model.Command, 3)}, true},
		{"unlimited variables", appContent{variables: make([]*model.Variable, 100)}, false},
		{
			"discord and webhook count separately",
			appContent{eventListeners: []*model.EventListener{
				listener(model.EventSourceDiscord),
				listener(model.EventSourceWebhook),
				listener(model.EventSourceSchedule),
			}},
			false,
		},
		{
			"too many discord listeners",
			appContent{eventListeners: []*model.EventListener{
				listener(model.EventSourceDiscord),
				listener(model.EventSourceDiscord),
			}},
			true,
		},
		{
			"too many scheduled listeners",
			appContent{eventListeners: []*model.EventListener{
				listener(model.EventSourceSchedule),
				listener(model.EventSourceSchedule),
			}},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkContentLimits(&tt.content, features)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInstallationToWire(t *testing.T) {
	res := installationToWire(discordIntegrationTypesConfig{
		// What Discord returns for a new app.
		discordGuildInstall: {OAuth2InstallParams: &discordInstallParams{Scopes: []string{"bot", "applications.commands"}, Permissions: "0"}},
		discordUserInstall:  {OAuth2InstallParams: &discordInstallParams{Scopes: []string{"applications.commands"}, Permissions: "0"}},
	})
	require.NotNil(t, res.GuildInstall)
	require.NotNil(t, res.UserInstall)
	assert.Equal(t, "8", res.GuildInstall.Permissions)
	assert.Equal(t, "0", res.UserInstall.Permissions)

	res = installationToWire(discordIntegrationTypesConfig{
		discordGuildInstall: {OAuth2InstallParams: &discordInstallParams{Scopes: []string{"bot", "applications.commands"}, Permissions: "268446720"}},
	})
	assert.Equal(t, "268446720", res.GuildInstall.Permissions)
	assert.Nil(t, res.UserInstall)

	res = installationToWire(discordIntegrationTypesConfig{discordGuildInstall: {}})
	assert.Equal(t, "8", res.GuildInstall.Permissions)
}
