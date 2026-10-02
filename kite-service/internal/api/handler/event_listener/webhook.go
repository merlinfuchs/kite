package eventlistener

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
	"github.com/kitecloud/kite/kite-service/internal/core/engine"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/internal/util"
	"github.com/kitecloud/kite/kite-service/pkg/webhook"
	"gopkg.in/guregu/null.v4"
)

const (
	// Requests a webhook listener accepts per interval. Every request runs a
	// flow, and anyone the URL was shared with can send them.
	webhookRateLimit         = 60
	webhookRateLimitInterval = time.Minute
)

// WebhookRunner runs the flows of webhook event listeners. It's implemented
// by the engine.
type WebhookRunner interface {
	WebhookSecret(appID string, listenerID string) (string, bool)
	RunWebhook(ctx context.Context, appID string, listenerID string, event *webhook.Event) error
}

func newWebhookSecret(source model.EventSource) null.String {
	if source != model.EventSourceWebhook {
		return null.String{}
	}
	return null.NewString(util.SecureURLKey(), true)
}

func (h *EventListenerHandler) HandleEventListenerWebhookSecretRegenerate(c *handler.Context) (*wire.EventListenerWebhookSecretRegenerateResponse, error) {
	if c.EventListener.Source != model.EventSourceWebhook {
		return nil, handler.ErrBadRequest("invalid_source", "only webhook event listeners have a webhook URL")
	}

	eventListener, err := h.eventListenerStore.UpdateEventListenerWebhookSecret(
		c.Context(),
		c.EventListener.ID,
		util.SecureURLKey(),
		time.Now().UTC(),
	)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil, handler.ErrNotFound("unknown_event_listener", "Event listener not found")
		}
		return nil, fmt.Errorf("failed to update webhook secret: %w", err)
	}

	return wire.EventListenerToWire(eventListener), nil
}

// HandleEventListenerWebhook receives a request to the webhook URL of a
// listener and runs its flow in the background. It's public, the secret in
// the URL is the only authentication.
func (h *EventListenerHandler) HandleEventListenerWebhook(c *handler.Context) error {
	appID := c.Param("appID")
	listenerID := c.Param("listenerID")
	secret := c.Param("secret")

	// The same for a wrong listener and a wrong secret, so listener IDs can't
	// be probed.
	notFound := handler.ErrNotFound("unknown_webhook", "Webhook not found")

	expected, ok := h.webhookRunner.WebhookSecret(appID, listenerID)
	if !ok || !secretsEqual(secret, expected) {
		return notFound
	}

	// After the secret check, so requests with a wrong secret can't use up
	// the limit of a listener.
	_, _, reset, ok, err := h.webhookLimiter.Take(c.Context(), listenerID)
	if err != nil {
		return handler.ErrInternal("failed to take rate limit token")
	}
	if !ok {
		c.SetHeader("Retry-After", time.Unix(0, int64(reset)).UTC().Format(time.RFC1123))
		return handler.ErrRateLimit("This webhook is receiving too many requests. Please try again later.")
	}

	// The engine only notices deleted listeners after a while, and changes
	// like a new secret after its next poll.
	listener, err := h.eventListenerStore.EventListener(c.Context(), listenerID)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return notFound
		}
		// Not returned as is, unexpected errors are logged with the request
		// path, which contains the secret.
		slog.Error(
			"Failed to get event listener for webhook",
			slog.String("event_listener_id", listenerID),
			slog.String("error", err.Error()),
		)
		return handler.ErrInternal("failed to get event listener")
	}
	if listener.AppID != appID ||
		!listener.Enabled ||
		listener.Source != model.EventSourceWebhook ||
		!listener.WebhookSecret.Valid ||
		!secretsEqual(secret, listener.WebhookSecret.String) {
		return notFound
	}

	event, err := webhookEvent(c)
	if err != nil {
		return err
	}

	if err := h.webhookRunner.RunWebhook(c.Context(), appID, listenerID, event); err != nil {
		if errors.Is(err, engine.ErrWebhookListenerNotFound) {
			return notFound
		}
		if errors.Is(err, engine.ErrWebhookAppOffline) {
			return handler.ErrServiceUnavailable("app_offline", "The app is offline, try again later")
		}
		return handler.ErrInternal("failed to run webhook")
	}

	return c.JSON(http.StatusAccepted, wire.APIResponse[*wire.EventListenerWebhookResponse]{
		Success: true,
		Data:    &wire.EventListenerWebhookResponse{},
	})
}

func secretsEqual(a string, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func webhookEvent(c *handler.Context) (*webhook.Event, error) {
	event := &webhook.Event{
		Headers: make(map[string]string),
		Query:   make(map[string]string),
	}

	metadataSize := 0
	for name, values := range c.Headers() {
		name = strings.ToLower(name)
		// Browsers send the cookies of the API with requests to it, which
		// would hand the session of a logged in user to the flow.
		if name == "cookie" {
			continue
		}

		event.Headers[name] = strings.Join(values, ", ")
		metadataSize += len(name) + len(event.Headers[name])
	}
	for name, values := range c.QueryValues() {
		event.Query[name] = strings.Join(values, ",")
		metadataSize += len(name) + len(event.Query[name])
	}
	if metadataSize > webhook.MaxMetadataSize {
		return nil, handler.ErrBadRequest("resource_limit", fmt.Sprintf(
			"request headers and query exceed the maximum allowed size (%d)", webhook.MaxMetadataSize,
		))
	}

	body, err := c.ReadBody(webhook.MaxBodySize)
	if err != nil {
		var aerr *handler.Error
		if errors.As(err, &aerr) {
			return nil, err
		}
		return nil, handler.ErrBadRequest("invalid_body", "failed to read request body")
	}
	event.Body = string(body)

	return event, nil
}
