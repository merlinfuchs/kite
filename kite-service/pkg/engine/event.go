package engine

import (
	"context"
	"log/slog"
	"sync"

	"github.com/merlinfuchs/kite/go-types/event"
	"github.com/merlinfuchs/kite/kite-service/internal/logging/logattr"
)

func (e *Engine) HandleEvent(ctx context.Context, event *event.Event) {
	var wg sync.WaitGroup

	for _, d := range e.Deployments[event.GuildID] {
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
