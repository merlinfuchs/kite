package flow

import "testing"

// The auto-defer has to declare ephemeral-ness before the flow picks a branch,
// so it guesses from the first response node it can reach. The guess used to
// be made with FindChildWithType, which only walks Children.Default -- so a
// response behind a condition or button handle was invisible and the flow
// deferred public even when its only response was ephemeral.
//
// These tests deliberately do NOT assert that an ephemeral response anywhere
// forces an ephemeral defer. The defer only binds a response that edits the
// original, so promoting every flow with an ephemeral branch would force
// intended-public first responses ephemeral.

func responseNode(id string, ephemeral bool) *CompiledFlowNode {
	return &CompiledFlowNode{
		ID:   id,
		Type: FlowNodeTypeActionResponseCreate,
		Data: FlowNodeData{MessageEphemeral: ephemeral},
	}
}

func entryWith(defaults []*CompiledFlowNode, handles map[string][]*CompiledFlowNode) *CompiledFlowNode {
	if handles == nil {
		handles = map[string][]*CompiledFlowNode{}
	}
	return &CompiledFlowNode{
		ID:       "entry",
		Type:     FlowNodeTypeEntryCommand,
		Children: ConnectedFlowNodes{Default: defaults, Handles: handles},
	}
}

func firstResponse(t *testing.T, entry *CompiledFlowNode) *CompiledFlowNode {
	t.Helper()
	return entry.FirstChildMatching(isResponseNode)
}

// The reported bug: the only response sits behind a condition branch, so the
// old walker found nothing and deferred public.
func TestResponseBehindHandleIsFound(t *testing.T) {
	entry := entryWith(nil, map[string][]*CompiledFlowNode{
		"condition_1": {responseNode("ephemeral", true)},
	})

	found := firstResponse(t, entry)
	if found == nil {
		t.Fatal("response behind a handle was not found")
	}
	if !found.Data.MessageEphemeral {
		t.Error("found response is not the ephemeral one")
	}
}

// A public first response must stay public even when a later branch responds
// ephemerally, otherwise the defer forces the public reply ephemeral.
func TestPublicFirstResponseWinsOverEphemeralBranch(t *testing.T) {
	entry := entryWith(
		[]*CompiledFlowNode{responseNode("public", false)},
		map[string][]*CompiledFlowNode{
			"condition_1": {responseNode("ephemeral", true)},
		},
	)

	found := firstResponse(t, entry)
	if found == nil {
		t.Fatal("no response found")
	}
	if found.Data.MessageEphemeral {
		t.Error("deferred ephemeral, which would force the public response ephemeral")
	}
}

// Direct children are reached before deeper ones.
func TestDirectChildPreferredOverNested(t *testing.T) {
	nested := entryWith([]*CompiledFlowNode{responseNode("deep", true)}, nil)
	nested.ID = "nested"
	nested.Type = FlowNodeTypeControlConditionCompare

	entry := entryWith([]*CompiledFlowNode{nested, responseNode("shallow", false)}, nil)

	found := firstResponse(t, entry)
	if found == nil || found.ID != "shallow" {
		t.Errorf("found %v, want the direct child", found)
	}
}

// Handle iteration must not depend on Go's random map order, or the same flow
// would defer differently between runs.
func TestHandleOrderIsStable(t *testing.T) {
	build := func() *CompiledFlowNode {
		return entryWith(nil, map[string][]*CompiledFlowNode{
			"a_branch": {responseNode("a", false)},
			"b_branch": {responseNode("b", true)},
			"c_branch": {responseNode("c", true)},
		})
	}

	want := firstResponse(t, build())
	for i := 0; i < 50; i++ {
		if got := firstResponse(t, build()); got.ID != want.ID {
			t.Fatalf("iteration %d picked %q, want %q", i, got.ID, want.ID)
		}
	}
}

// Non-response nodes carry MessageEphemeral too and must not be mistaken for a
// response.
func TestNonResponseNodeIgnored(t *testing.T) {
	entry := entryWith([]*CompiledFlowNode{{
		ID:   "send",
		Type: FlowNodeTypeActionMessageCreate,
		Data: FlowNodeData{MessageEphemeral: true},
	}}, nil)

	if found := firstResponse(t, entry); found != nil {
		t.Errorf("non-response node %q counted as a response", found.ID)
	}
}

// Flows can loop back on themselves; the walk must terminate.
func TestCycleTerminates(t *testing.T) {
	a := &CompiledFlowNode{ID: "a", Children: ConnectedFlowNodes{Handles: map[string][]*CompiledFlowNode{}}}
	b := &CompiledFlowNode{ID: "b", Children: ConnectedFlowNodes{Handles: map[string][]*CompiledFlowNode{}}}
	a.Children.Default = []*CompiledFlowNode{b}
	b.Children.Default = []*CompiledFlowNode{a}

	entry := entryWith([]*CompiledFlowNode{a}, nil)

	if found := firstResponse(t, entry); found != nil {
		t.Errorf("unexpected match %q in a cyclic flow", found.ID)
	}
}
