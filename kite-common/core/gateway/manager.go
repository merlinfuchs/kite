package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/kitecloud/kite/kite-common/model"
	"github.com/kitecloud/kite/kite-common/store"
)

type EventHandler interface {
	HandleEvent(appID string, event gateway.Event)
}

type GatewayManager struct {
	sync.Mutex

	appStore     store.AppStore
	eventHandler EventHandler

	lastUpdate time.Time
	gateways   map[string]*Gateway
}

func NewGatewayManager(appStore store.AppStore, eventHandler EventHandler) *GatewayManager {
	return &GatewayManager{
		appStore:     appStore,
		eventHandler: eventHandler,
		gateways:     make(map[string]*Gateway),
	}
}

func (m *GatewayManager) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)

	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				if err := m.populateGateways(ctx); err != nil {
					slog.With("error", err).Error("failed to populate gateways")
				}
			}
		}
	}()
}

func (m *GatewayManager) populateGateways(ctx context.Context) error {
	appIDs, err := m.appStore.AppIDs(ctx)
	if err != nil {
		return fmt.Errorf("failed to get enabled apps: %w", err)
	}

	lastUpdate := m.lastUpdate
	m.lastUpdate = time.Now().UTC()

	apps, err := m.appStore.AppsUpdatedSince(ctx, lastUpdate)
	if err != nil {
		return fmt.Errorf("failed to get apps updated since %s: %w", lastUpdate, err)
	}

	if err := m.removeDanglingGateways(ctx, appIDs); err != nil {
		return fmt.Errorf("failed to remove dangling apps: %w", err)
	}

	for _, integration := range apps {
		if err := m.addGateway(ctx, integration); err != nil {
			slog.With("error", err).Error("failed to add gateway")
		}
	}

	return nil
}

func (m *GatewayManager) removeDanglingGateways(ctx context.Context, appIDs []string) error {
	m.Lock()
	defer m.Unlock()

	lookupMap := make(map[string]struct{}, len(appIDs))
	for _, id := range appIDs {
		lookupMap[id] = struct{}{}
	}

	for id, gateway := range m.gateways {
		if _, ok := lookupMap[id]; !ok {
			if err := gateway.Close(ctx); err != nil {
				slog.With("error", err).Error("failed to close gateway")
			}

			delete(m.gateways, id)
		}
	}

	return nil
}

func (m *GatewayManager) addGateway(ctx context.Context, app *model.App) error {
	m.Lock()
	defer m.Unlock()

	if g, ok := m.gateways[app.ID]; ok {
		g.Update(ctx, app)
	} else {
		g := NewGateway(app, m.eventHandler)
		m.gateways[app.ID] = g
	}

	return nil
}
