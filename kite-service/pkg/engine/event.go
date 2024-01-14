package engine

import (
	"context"
	"log/slog"
	"sync"

	"github.com/merlinfuchs/kite/go-types/event"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
)

func (e *PluginEngine) HandleEvent(ctx context.Context, event *event.Event) {
	var wg sync.WaitGroup

	for _, d := range e.StaticPlugins {
		wg.Add(1)

		d := d
		go func() {
			defer wg.Done()

			err := d.HandleEvent(ctx, event)
			if err != nil {
				slog.With(logattr.Error(err)).Error("Error handling event on plugin")
			}
		}()
	}

	deployments, exists := e.Deployments[event.GuildID]
	if !exists {
		return
	}

	for _, d := range deployments {
		wg.Add(1)

		d := d
		go func() {
			defer wg.Done()

			err := d.HandleEvent(ctx, event)
			if err != nil {
				slog.With(logattr.Error(err)).Error("Error handling event on plugin")
			}
		}()
	}

	wg.Wait()
}
