package flow

import (
	"testing"

	"gopkg.in/guregu/null.v4"
)

func TestContentHash(t *testing.T) {
	base := func() FlowData {
		return FlowData{
			Nodes: []FlowNode{
				{ID: "1", Type: FlowNodeTypeEntryCommand, Data: FlowNodeData{Name: "ban"}, Position: FlowNodePosition{X: 0, Y: 0}},
				{ID: "2", Type: FlowNodeTypeActionResponseCreate, Position: FlowNodePosition{X: 0, Y: 100}},
			},
			Edges: []FlowEdge{
				{ID: "e1", Source: "1", Target: "2", SourceHandle: null.StringFrom("a")},
			},
		}
	}

	hash := func(t *testing.T, d FlowData) string {
		t.Helper()
		h, err := d.ContentHash()
		if err != nil {
			t.Fatal(err)
		}
		return h
	}

	want := hash(t, base())

	moved := base()
	moved.Nodes[1].Position = FlowNodePosition{X: 300, Y: 50}
	moved.Nodes[0], moved.Nodes[1] = moved.Nodes[1], moved.Nodes[0]
	moved.Edges[0].ID = "other"
	if got := hash(t, moved); got != want {
		t.Errorf("moving blocks changed the hash")
	}

	edited := base()
	edited.Nodes[0].Data.Name = "kick"
	if got := hash(t, edited); got == want {
		t.Errorf("editing a block didn't change the hash")
	}

	rewired := base()
	rewired.Edges[0].SourceHandle = null.StringFrom("b")
	if got := hash(t, rewired); got == want {
		t.Errorf("changing an edge didn't change the hash")
	}
}
