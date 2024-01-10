package engine

import (
	"context"
	"slices"

	"github.com/merlinfuchs/kite/go-types/event"
)

func (e *PluginEngine) HandleEvent(ctx context.Context, event *event.Event) error {
	for _, plugin := range e.Plugins {
		if _, ok := plugin.GuildIDs[event.GuildID]; len(plugin.GuildIDs) == 0 || ok {
			plugin.Plugin.Lock()
			defer plugin.Plugin.Unlock()

			if !slices.Contains(plugin.Plugin.Manifest().Events, string(event.Type)) {
				continue
			}

			err := plugin.Plugin.Handle(ctx, event)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
