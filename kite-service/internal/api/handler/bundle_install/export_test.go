package bundleinstall

import (
	"testing"

	"github.com/kitecloud/kite/kite-service/internal/core/bundle"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

func TestAddUsed(t *testing.T) {
	node := func(data flow.FlowNodeData) flow.FlowData {
		return flow.FlowData{Nodes: []flow.FlowNode{{ID: "n", Data: data}}}
	}
	// The panel's button sends another message, which counts.
	panel := &model.Message{ID: "panel", FlowSources: map[string]flow.FlowData{
		"button": node(flow.FlowNodeData{MessageTemplateID: "reply", VariableID: "clicks"}),
	}}
	reply := &model.Message{ID: "reply"}
	unused := &model.Message{ID: "unused"}
	clicks := &model.Variable{ID: "clicks"}

	cmd := &model.Command{ID: "cmd", FlowSource: flow.FlowData{Nodes: []flow.FlowNode{
		{ID: "send", Data: flow.FlowNodeData{MessageTemplateID: "panel"}},
		{ID: "gone", Data: flow.FlowNodeData{MessageTemplateID: "deleted", VariableID: "deleted"}},
	}}}

	items := bundle.ExportItems{Commands: []*model.Command{cmd}}
	addUsed(&items, []*model.Message{panel, reply, unused}, []*model.Variable{clicks})

	assert.Equal(t, []*model.Message{panel, reply}, items.Messages)
	assert.Equal(t, []*model.Variable{clicks}, items.Variables)
	// Blocks can't keep ones the app doesn't have anymore.
	assert.Equal(t, "", cmd.FlowSource.Nodes[1].Data.MessageTemplateID)
	assert.Equal(t, "", cmd.FlowSource.Nodes[1].Data.VariableID)
	assert.Equal(t, "panel", cmd.FlowSource.Nodes[0].Data.MessageTemplateID)
}

func TestUsedByOthers(t *testing.T) {
	install := null.StringFrom("install")
	all := store.BundleInstallContents{
		Commands: []*model.Command{
			{ID: "own", BundleInstallID: install, FlowSource: flow.FlowData{Nodes: []flow.FlowNode{{Data: flow.FlowNodeData{MessageTemplateID: "only_own"}}}}},
			{ID: "later", BundleInstallID: null.StringFrom("later"), FlowSource: flow.FlowData{Nodes: []flow.FlowNode{{Data: flow.FlowNodeData{VariableID: "shared_var"}}}}},
		},
		Messages: []*model.Message{
			{ID: "user", FlowSources: map[string]flow.FlowData{"b": {Nodes: []flow.FlowNode{{Data: flow.FlowNodeData{MessageTemplateID: "shared_msg"}}}}}},
			// Kept because the user's message uses it, so what its buttons use
			// is kept too.
			{ID: "shared_msg", BundleInstallID: install, FlowSources: map[string]flow.FlowData{"b": {Nodes: []flow.FlowNode{{Data: flow.FlowNodeData{VariableID: "button_var"}}}}}},
		},
	}

	assert.Equal(t, map[string]bool{"shared_var": true, "shared_msg": true, "button_var": true}, usedByOthers(all, "install"))
}
