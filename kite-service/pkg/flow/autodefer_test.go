package flow

import (
	"testing"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
)

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

func newResponseNode(id string, ephemeral bool) *CompiledFlowNode {
	return &CompiledFlowNode{
		ID:   id,
		Type: FlowNodeTypeActionResponseCreate,
		Data: FlowNodeData{MessageEphemeral: ephemeral},
	}
}

func entryWith(defaults []*CompiledFlowNode, handles map[string][]*CompiledFlowNode) *CompiledFlowNode {
	return &CompiledFlowNode{
		ID:       "entry",
		Type:     FlowNodeTypeEntryCommand,
		Children: ConnectedFlowNodes{Default: defaults, Handles: handles},
	}
}

func firstResponse(entry *CompiledFlowNode) *CompiledFlowNode {
	return entry.FirstChildMatching(isResponseNode)
}

// The reported bug: the only response sits behind a condition branch, so the
// old walker found nothing and deferred public.
func TestResponseBehindHandleIsFound(t *testing.T) {
	entry := entryWith(nil, map[string][]*CompiledFlowNode{
		"condition_1": {newResponseNode("ephemeral", true)},
	})

	found := firstResponse(entry)
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
		[]*CompiledFlowNode{newResponseNode("public", false)},
		map[string][]*CompiledFlowNode{
			"condition_1": {newResponseNode("ephemeral", true)},
		},
	)

	found := firstResponse(entry)
	if found == nil {
		t.Fatal("no response found")
	}
	if found.Data.MessageEphemeral {
		t.Error("deferred ephemeral, which would force the public response ephemeral")
	}
}

// Direct children are reached before deeper ones.
func TestDirectChildPreferredOverNested(t *testing.T) {
	nested := &CompiledFlowNode{
		ID:       "nested",
		Type:     FlowNodeTypeControlConditionCompare,
		Children: ConnectedFlowNodes{Default: []*CompiledFlowNode{newResponseNode("deep", true)}},
	}

	entry := entryWith([]*CompiledFlowNode{nested, newResponseNode("shallow", false)}, nil)

	found := firstResponse(entry)
	if found == nil || found.ID != "shallow" {
		t.Errorf("found %v, want the direct child", found)
	}
}

// Handle iteration must not depend on Go's random map order, or the same flow
// would defer differently between runs.
func TestHandleOrderIsStable(t *testing.T) {
	build := func() *CompiledFlowNode {
		return entryWith(nil, map[string][]*CompiledFlowNode{
			"a_branch": {newResponseNode("a", false)},
			"b_branch": {newResponseNode("b", true)},
			"c_branch": {newResponseNode("c", true)},
		})
	}

	want := firstResponse(build())
	for i := 0; i < 50; i++ {
		if got := firstResponse(build()); got.ID != want.ID {
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

	if found := firstResponse(entry); found != nil {
		t.Errorf("non-response node %q counted as a response", found.ID)
	}
}

// Flows can loop back on themselves; the walk must terminate.
func TestCycleTerminates(t *testing.T) {
	a := &CompiledFlowNode{ID: "a"}
	b := &CompiledFlowNode{ID: "b"}
	a.Children.Default = []*CompiledFlowNode{b}
	b.Children.Default = []*CompiledFlowNode{a}

	entry := entryWith([]*CompiledFlowNode{a}, nil)

	if found := firstResponse(entry); found != nil {
		t.Errorf("unexpected match %q in a cyclic flow", found.ID)
	}
}

// On a button click only the clicked component's branch runs. The guess used
// to come from all of the message node's children, so the original flow's
// public follow-up decided the defer for an ephemeral button reply (#245).
func TestClickedBranchDecidesDefer(t *testing.T) {
	msgNode := &CompiledFlowNode{
		ID:   "message",
		Type: FlowNodeTypeActionMessageCreate,
		Children: ConnectedFlowNodes{
			Default: []*CompiledFlowNode{newResponseNode("posted", false)},
			Handles: map[string][]*CompiledFlowNode{
				"component_1": {newResponseNode("other_button", false)},
				"component_2": {newResponseNode("clicked", true)},
			},
		},
	}

	found := FirstMatching(msgNode.Children.Handles["component_2"], isResponseNode)
	if found == nil || found.ID != "clicked" {
		t.Fatalf("found %v, want the clicked branch's response", found)
	}

	resp := autoDeferResponse(buttonInteraction(), found)
	if resp.Type != api.DeferredMessageInteractionWithSource || resp.Data.Flags&discord.EphemeralMessage == 0 {
		t.Errorf("got %+v, want an ephemeral deferred response", resp)
	}
}

func buttonInteraction() *discord.InteractionEvent {
	return &discord.InteractionEvent{Data: &discord.ButtonInteraction{CustomID: "x"}}
}

// A component whose branch edits the original message, or doesn't respond at
// all, must only be acknowledged. A "thinking…" message would otherwise be
// edited instead of the component's message.
func TestComponentWithoutNewResponseDefersUpdate(t *testing.T) {
	edit := &CompiledFlowNode{ID: "edit", Type: FlowNodeTypeActionResponseEdit}

	for _, node := range []*CompiledFlowNode{nil, edit} {
		if resp := autoDeferResponse(buttonInteraction(), node); resp.Type != api.DeferredMessageUpdate {
			t.Errorf("response node %v: got type %d, want DeferredMessageUpdate", node, resp.Type)
		}
	}
}

// Commands can't be acknowledged without a response, so they keep deferring
// with a "thinking…" message.
func TestCommandAlwaysDefersWithSource(t *testing.T) {
	command := &discord.InteractionEvent{Data: &discord.CommandInteraction{}}

	if resp := autoDeferResponse(command, nil); resp.Type != api.DeferredMessageInteractionWithSource {
		t.Errorf("got type %d, want DeferredMessageInteractionWithSource", resp.Type)
	}
}
