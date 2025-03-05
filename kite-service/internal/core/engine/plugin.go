package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
)

type PluginInstance struct {
	model    *model.PluginInstance
	plugin   plugin.Plugin
	instance plugin.PluginInstance
	env      Env
}

func NewPluginInstance(appID string, model *model.PluginInstance, env Env) (*PluginInstance, error) {
	pl := env.PluginRegistry.Plugin(model.PluginID)
	if pl == nil {
		return nil, fmt.Errorf("plugin not found")
	}

	var config plugin.ConfigValues
	err := json.Unmarshal(model.Config, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	instance, err := pl.Instance(appID, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create plugin instance: %w", err)
	}

	return &PluginInstance{
		model:    model,
		plugin:   pl,
		instance: instance,
		env:      env,
	}, nil
}

func (p *PluginInstance) Update(ctx context.Context, discord *state.State) error {
	c := newPluginContext(
		ctx,
		p.env.PluginValueStore,
		discord,
		p.model.AppID,
		p.model.PluginID,
	)

	return p.instance.Update(c)
}

func (p *PluginInstance) HandleEvent(_ string, discord *state.State, event gateway.Event) {
	c := newPluginContext(
		context.Background(),
		p.env.PluginValueStore,
		discord,
		p.model.AppID,
		p.model.PluginID,
	)

	err := p.instance.HandleEvent(c, event)
	if err != nil {
		slog.Error("failed to handle event", "error", err)
	}
}
