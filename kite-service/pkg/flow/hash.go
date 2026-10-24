package flow

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/kitecloud/kite/kite-service/internal/util"
)

// ContentHash identifies what a flow does. Moving blocks around and the IDs of
// edges don't change it, so it only differs after a real edit.
func (d FlowData) ContentHash() (string, error) {
	type node struct {
		ID   string       `json:"id"`
		Type FlowNodeType `json:"type"`
		Data FlowNodeData `json:"data"`
	}
	type edge struct {
		Source       string `json:"source"`
		Target       string `json:"target"`
		SourceHandle string `json:"source_handle"`
		TargetHandle string `json:"target_handle"`
	}

	nodes := make([]node, len(d.Nodes))
	for i, n := range d.Nodes {
		nodes[i] = node{ID: n.ID, Type: n.Type, Data: n.Data}
	}
	slices.SortFunc(nodes, func(a, b node) int { return strings.Compare(a.ID, b.ID) })

	edges := make([]edge, len(d.Edges))
	for i, e := range d.Edges {
		edges[i] = edge{
			Source:       e.Source,
			Target:       e.Target,
			SourceHandle: e.SourceHandle.String,
			TargetHandle: e.TargetHandle.String,
		}
	}
	slices.SortFunc(edges, func(a, b edge) int {
		return strings.Compare(
			a.Source+"\x00"+a.SourceHandle+"\x00"+a.Target+"\x00"+a.TargetHandle,
			b.Source+"\x00"+b.SourceHandle+"\x00"+b.Target+"\x00"+b.TargetHandle,
		)
	})

	raw, err := json.Marshal(struct {
		Nodes []node `json:"nodes"`
		Edges []edge `json:"edges"`
	}{nodes, edges})
	if err != nil {
		return "", err
	}

	return util.HashBytes(raw), nil
}
