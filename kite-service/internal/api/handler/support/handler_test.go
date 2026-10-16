package support

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/api/handler"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"gopkg.in/guregu/null.v4"
)

type testAppStore struct {
	store.AppStore
	requestedID string
}

func (s *testAppStore) App(_ context.Context, id string) (*model.App, error) {
	s.requestedID = id
	return &model.App{ID: id}, nil
}

type testEventListenerStore struct {
	store.EventListenerStore
	listeners []*model.EventListener
}

func (s *testEventListenerStore) EventListenersByApp(_ context.Context, _ string) ([]*model.EventListener, error) {
	return s.listeners, nil
}

func TestEventListenerListDoesNotExposeWebhookSecret(t *testing.T) {
	appStore := &testAppStore{}
	eventListenerStore := &testEventListenerStore{
		listeners: []*model.EventListener{{
			ID:            "listener-id",
			AppID:         "kite-app-id",
			Source:        model.EventSourceWebhook,
			WebhookSecret: null.StringFrom("private-webhook-secret"),
		}},
	}
	h := NewSupportHandler(appStore, nil, nil, eventListenerStore)

	mux := http.NewServeMux()
	group := handler.Group(mux, "/v1/support/apps/{appID}")
	group.Get("/event-listeners", handler.Typed(h.HandleEventListenerList))
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/v1/support/apps/kite-app-id/event-listeners", nil))

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", res.Code, http.StatusOK, res.Body.String())
	}
	if appStore.requestedID != "kite-app-id" {
		t.Fatalf("looked up app ID %q, want Kite App ID", appStore.requestedID)
	}
	if strings.Contains(res.Body.String(), "private-webhook-secret") || strings.Contains(res.Body.String(), "webhook_secret") {
		t.Fatalf("support response exposed a webhook secret: %s", res.Body.String())
	}
}
