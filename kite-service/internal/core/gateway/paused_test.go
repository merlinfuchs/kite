package gateway

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

type fakeMemberAppStore struct {
	store.AppStore
	memberIDs []string
}

func (s *fakeMemberAppStore) AppMemberDiscordIDs(ctx context.Context, appID string) ([]string, error) {
	return s.memberIDs, nil
}

func TestPausedMessage(t *testing.T) {
	g := newTestGateway(t)
	g.appStore = &fakeMemberAppStore{memberIDs: []string{"1"}}
	g.appBaseURL = "https://kite.onl"
	now := time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)
	app := &model.App{ID: "app", CreditsPausedAt: null.TimeFrom(now.Add(-time.Hour))}

	member := g.pausedMessage(app, discord.UserID(1), now)
	if !strings.Contains(member, "November 1") || !strings.Contains(member, "https://kite.onl/apps/app/premium") {
		t.Errorf("member message = %q, want reset date and upgrade link", member)
	}

	other := g.pausedMessage(app, discord.UserID(2), now)
	if !strings.Contains(other, "November 1") || strings.Contains(other, "kite.onl") {
		t.Errorf("message for others = %q, want reset date and no link", other)
	}

	lastMonth := g.pausedMessage(app, discord.UserID(1), time.Date(2026, 11, 1, 0, 3, 0, 0, time.UTC))
	if !strings.Contains(lastMonth, "in a few minutes") {
		t.Errorf("message after the month of the pause = %q, want it to start in a few minutes", lastMonth)
	}

	december := g.pausedMessage(&model.App{CreditsPausedAt: null.TimeFrom(time.Date(2026, 12, 2, 0, 0, 0, 0, time.UTC))}, discord.UserID(2), time.Date(2026, 12, 31, 23, 0, 0, 0, time.UTC))
	if !strings.Contains(december, "January 1") {
		t.Errorf("message in December = %q, want January 1", december)
	}
}

func TestAppSessionHidesPausedApps(t *testing.T) {
	g := newTestGateway(t)
	m := &GatewayManager{gateways: map[string]*Gateway{"app": g}}

	if s, err := m.AppSession(context.Background(), "app"); err != nil || s == nil {
		t.Fatalf("AppSession of enabled app = %v, %v, want session", s, err)
	}

	g.app = &model.App{ID: "app", Enabled: true, CreditsPausedAt: null.TimeFrom(time.Now())}
	if s, err := m.AppSession(context.Background(), "app"); err == nil || s != nil {
		t.Errorf("AppSession of paused app = %v, %v, want none", s, err)
	}
	if c, err := m.AppClient(context.Background(), "app"); err != nil || c == nil {
		t.Errorf("AppClient of paused app = %v, %v, want client for the dashboard", c, err)
	}
}
