package engine

import (
	"context"
	"testing"

	"github.com/diamondburned/arikawa/v3/state"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/webhook"
	"gopkg.in/guregu/null.v4"
)

type fakeAppStore struct {
	store.AppStore
	loads int
}

func (f *fakeAppStore) AppDiscordUserInstallCount(ctx context.Context, id string) (null.Int, error) {
	f.loads++
	return null.IntFrom(42), nil
}

// The install count is loaded once a flow reads it, including through the
// trigger from before a resume point.
func TestFlowContextUserInstallCount(t *testing.T) {
	apps := &fakeAppStore{}
	s := Env{AppStore: apps}
	flowState := flow.NewFlowContextState()
	flowState.Triggers = []flow.FlowTrigger{{Event: &webhook.Event{}}}

	fCtx := s.flowContext(context.Background(), "app", state.New("Bot test"), &EventData{event: &webhook.Event{}}, entityLinks{}, flowState, nil)
	defer fCtx.Cancel()

	got, err := eval.EvalTemplateToString(fCtx, "{{app.user_install_count}}|{{origin.app.user_install_count}}", fCtx.EvalCtx)
	if err != nil {
		t.Fatal(err)
	}
	if got != "42|42" {
		t.Errorf("got %q", got)
	}
	if apps.loads != 1 {
		t.Errorf("loaded %d times, want 1", apps.loads)
	}
}

// A flow has the values of its parameters, also when it continues after a
// sleep or a click from a node in the middle.
func TestFlowContextParameters(t *testing.T) {
	node, err := flow.CompileCommand(flow.FlowData{
		Nodes:      []flow.FlowNode{{ID: "entry", Type: flow.FlowNodeTypeEntryCommand}},
		Parameters: []flow.FlowParameter{{Key: "greeting", Type: flow.FlowParameterTypeText, Value: "Hi"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	s := Env{AppStore: &fakeAppStore{}}
	fCtx := s.flowContext(context.Background(), "app", state.New("Bot test"), &EventData{event: &webhook.Event{}}, entityLinks{}, nil, node)
	defer fCtx.Cancel()

	got, err := eval.EvalTemplateToString(fCtx, "{{params.greeting}}", fCtx.EvalCtx)
	if err != nil {
		t.Fatal(err)
	}
	if got != "Hi" {
		t.Errorf("got %q", got)
	}
}
