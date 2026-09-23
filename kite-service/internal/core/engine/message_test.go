package engine

import (
	"context"
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

type fakeMessageStore struct {
	store.MessageStore
	message *model.Message
	err     error
}

func (f *fakeMessageStore) Message(ctx context.Context, id string) (*model.Message, error) {
	return f.message, f.err
}

func flowWithNode(id string) flow.FlowData {
	return flow.FlowData{Nodes: []flow.FlowNode{{ID: id}}}
}

// Edits to a template's button flows have to apply to messages that were sent
// before the edit (#304).
func TestLiveFlowSourcesPreferTemplate(t *testing.T) {
	instance := &model.MessageInstance{
		MessageID: "template",
		FlowSources: map[string]flow.FlowData{
			"edited":  flowWithNode("old"),
			"removed": flowWithNode("snapshot"),
		},
	}
	messages := &fakeMessageStore{message: &model.Message{
		FlowSources: map[string]flow.FlowData{
			"edited": flowWithNode("new"),
			"added":  flowWithNode("added"),
		},
	}}

	got := liveFlowSources(messages, instance)

	if got["edited"].Nodes[0].ID != "new" {
		t.Errorf("edited flow is %q, want the template's current flow", got["edited"].Nodes[0].ID)
	}
	if got["removed"].Nodes[0].ID != "snapshot" {
		t.Error("component removed from the template should keep its snapshot flow")
	}
	if _, ok := got["added"]; ok {
		t.Error("component that isn't on the sent message should not get a flow")
	}
}

func TestLiveFlowSourcesFallBackToSnapshot(t *testing.T) {
	instance := &model.MessageInstance{
		FlowSources: map[string]flow.FlowData{"a": flowWithNode("snapshot")},
	}

	got := liveFlowSources(&fakeMessageStore{err: store.ErrNotFound}, instance)

	if got["a"].Nodes[0].ID != "snapshot" {
		t.Error("expected the snapshot when the template is gone")
	}
}
