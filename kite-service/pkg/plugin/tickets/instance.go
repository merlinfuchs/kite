package tickets

import (
	"context"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/kitecloud/kite/kite-service/pkg/plugin"
)

type TicketsPluginInstance struct {
	appID  string
	config plugin.ConfigValues
}

func (p *TicketsPluginInstance) Update(ctx context.Context, config plugin.ConfigValues) error {
	p.config = config
	return nil
}

func (p *TicketsPluginInstance) HandleEvent(c plugin.Context, event gateway.Event) error {
	return nil
}

func (p *TicketsPluginInstance) HandleCommand(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	return nil
}

func (p *TicketsPluginInstance) HandleComponent(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	return nil
}

func (p *TicketsPluginInstance) HandleModal(c plugin.Context, event *gateway.InteractionCreateEvent) error {
	return nil
}

func (p *TicketsPluginInstance) Close() error {
	return nil
}
