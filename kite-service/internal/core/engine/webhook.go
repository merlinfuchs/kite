package engine

import (
	"context"
	"errors"

	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/webhook"
	"gopkg.in/guregu/null.v4"
)

var (
	// ErrWebhookListenerNotFound means the listener doesn't exist, is
	// disabled, isn't a webhook listener or hasn't been loaded yet.
	ErrWebhookListenerNotFound = errors.New("webhook event listener not found")
	// ErrWebhookAppOffline means the app has no gateway on this cluster, so
	// its flows have no Discord session to run with.
	ErrWebhookAppOffline = errors.New("app is offline")
)

// WebhookRunner runs webhook event listeners for the requests the API
// receives. Sessions come from the gateway manager, which is created after
// the engine.
type WebhookRunner struct {
	engine   *Engine
	sessions SessionProvider
}

func (e *Engine) WebhookRunner(sessions SessionProvider) *WebhookRunner {
	return &WebhookRunner{
		engine:   e,
		sessions: sessions,
	}
}

// WebhookSecret returns the secret of an enabled webhook listener. It reads
// what the engine has loaded, so requests with a wrong URL never reach the
// database.
func (r *WebhookRunner) WebhookSecret(appID string, listenerID string) (string, bool) {
	listener := r.engine.webhookEventListener(appID, listenerID)
	if listener == nil || !listener.listener.WebhookSecret.Valid {
		return "", false
	}
	return listener.listener.WebhookSecret.String, true
}

// RunWebhook starts the flow of a webhook listener in the background.
func (r *WebhookRunner) RunWebhook(ctx context.Context, appID string, listenerID string, event *webhook.Event) error {
	listener := r.engine.webhookEventListener(appID, listenerID)
	if listener == nil {
		return ErrWebhookListenerNotFound
	}

	session, err := r.sessions.AppSession(ctx, appID)
	if err != nil || session == nil {
		return ErrWebhookAppOffline
	}

	go listener.HandleWebhookRun(session, event)
	return nil
}

func (e *Engine) webhookEventListener(appID string, listenerID string) *EventListener {
	e.RLock()
	app := e.apps[appID]
	e.RUnlock()

	if app == nil {
		return nil
	}
	return app.webhookEventListener(listenerID)
}

func (a *App) webhookEventListener(listenerID string) *EventListener {
	a.RLock()
	defer a.RUnlock()

	listener, ok := a.listeners[listenerID]
	if !ok || listener.listener.Source != model.EventSourceWebhook {
		return nil
	}
	return listener
}

func (l *EventListener) HandleWebhookRun(session *state.State, event *webhook.Event) {
	l.env.executeFlowEvent(
		context.Background(),
		l.listener.AppID,
		l.flow,
		session,
		event,
		entityLinks{
			EventListenerID: null.NewString(l.listener.ID, true),
		},
		nil,
	)
}
