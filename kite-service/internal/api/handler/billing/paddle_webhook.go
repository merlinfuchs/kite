package billing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/PaddleHQ/paddle-go-sdk/v6"
	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/api/wire"
)

// paddleWebhookMaxBodySize is what the Paddle SDK reads of a webhook.
const paddleWebhookMaxBodySize = 2 << 20

// paddleWebhookTolerance is how old a webhook's signature may be. The SDK
// allows 5 seconds against replays, but a replayed event does no harm here,
// because the subscription is fetched from the API either way.
const paddleWebhookTolerance = 5 * time.Minute

func (h *BillingHandler) HandlePaddleWebhook(c *handler.Context) (*wire.BillingWebhookResponse, error) {
	if h.config.PaddleWebhookSecret == "" {
		return nil, handler.ErrUnauthorized("invalid_signature", "failed to verify webhook signature")
	}

	// The signature is over the exact bytes Paddle sent, so the body is read
	// as it is rather than parsed first.
	body, err := c.ReadBody(paddleWebhookMaxBodySize)
	if err != nil {
		return nil, err
	}

	verifier := paddle.NewWebhookVerifier(
		h.config.PaddleWebhookSecret,
		paddle.VerifierWithTimestampTolerance(paddleWebhookTolerance),
	)
	ok, err := verifier.Verify(&http.Request{
		Header: c.Headers(),
		Body:   io.NopCloser(bytes.NewReader(body)),
	})
	if err != nil || !ok {
		return nil, handler.ErrUnauthorized("invalid_signature", "failed to verify webhook signature")
	}

	var event struct {
		EventType string `json:"event_type"`
		Data      struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, handler.ErrBadRequest("invalid_json", fmt.Sprintf("failed to unmarshal webhook event: %s", err))
	}

	// Every subscription.* event has the subscription as its data. Everything
	// else, like transactions, is acknowledged and ignored, so Paddle doesn't
	// retry it.
	if !strings.HasPrefix(event.EventType, "subscription.") {
		return &wire.BillingWebhookResponse{}, nil
	}

	// Paddle doesn't guarantee the order of events, so the state is fetched
	// from the API instead of taken from the payload.
	if _, err := h.subscriptionManager.RefreshPaddle(c.Context(), event.Data.ID); err != nil {
		slog.Error(
			"Failed to sync Paddle subscription from webhook",
			slog.String("event_type", event.EventType),
			slog.String("paddle_subscription_id", event.Data.ID),
			slog.String("error", err.Error()),
		)
		return nil, err
	}

	return &wire.BillingWebhookResponse{}, nil
}
