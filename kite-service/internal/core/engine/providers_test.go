package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
)

type fakeMessageStore struct {
	store.MessageStore
	messages map[string]*model.Message
}

func (f *fakeMessageStore) Message(ctx context.Context, appID string, id string) (*model.Message, error) {
	msg, ok := f.messages[id]
	if !ok || msg.AppID != appID {
		return nil, store.ErrNotFound
	}
	return msg, nil
}

// Template IDs in flow data are user-authored, so a flow must not be able to
// send or link another app's template.
func TestMessageTemplateProviderScopedToApp(t *testing.T) {
	messages := &fakeMessageStore{messages: map[string]*model.Message{
		"own":   {ID: "own", AppID: "app"},
		"other": {ID: "other", AppID: "other_app"},
	}}
	p := NewMessageTemplateProvider("app", messages, nil)

	if _, err := p.MessageTemplate(context.Background(), "own"); err != nil {
		t.Errorf("own template: %v", err)
	}
	if _, err := p.MessageTemplate(context.Background(), "other"); !errors.Is(err, store.ErrNotFound) {
		t.Errorf("other app's template: got %v, want ErrNotFound", err)
	}
}
