package gateway

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"gopkg.in/guregu/null.v4"
)

func TestPausedMessage(t *testing.T) {
	g := newTestGateway(t)
	g.appBaseURL = "https://kite.onl"
	now := time.Date(2026, 10, 29, 12, 0, 0, 0, time.UTC)
	app := &model.App{ID: "app", CreditsPausedAt: null.TimeFrom(now.Add(-time.Hour))}

	msg := g.pausedMessage(app, now)
	if !strings.Contains(msg, "November 1 (UTC)") || !strings.Contains(msg, "[Kite](<https://kite.onl>)") {
		t.Errorf("message = %q, want reset date and Kite link", msg)
	}

	lastMonth := g.pausedMessage(app, time.Date(2026, 11, 1, 0, 3, 0, 0, time.UTC))
	if !strings.Contains(lastMonth, "in a few minutes") {
		t.Errorf("message after the month of the pause = %q, want it to start in a few minutes", lastMonth)
	}

	december := g.pausedMessage(&model.App{CreditsPausedAt: null.TimeFrom(time.Date(2026, 12, 2, 0, 0, 0, 0, time.UTC))}, time.Date(2026, 12, 31, 23, 0, 0, 0, time.UTC))
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
