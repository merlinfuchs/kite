package wire

import (
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/stretchr/testify/assert"
)

func TestModuleCreateRequestValidate(t *testing.T) {
	flowSource := flow.FlowData{Nodes: []flow.FlowNode{{
		ID:   "1",
		Type: flow.FlowNodeTypeEntryCommand,
		Data: flow.FlowNodeData{Name: "ban", Description: "Ban a user."},
	}}}
	command := func(key string) ModuleCommandCreateRequest {
		return ModuleCommandCreateRequest{Key: key, Command: CommandCreateRequest{FlowSource: flowSource}}
	}
	listener := func(key string) ModuleEventListenerCreateRequest {
		return ModuleEventListenerCreateRequest{Key: key, EventListener: EventListenerCreateRequest{Source: "discord", FlowSource: flowSource}}
	}
	valid := func() ModuleCreateRequest {
		return ModuleCreateRequest{
			TemplateID:      "moderation",
			TemplateVersion: 1,
			Name:            "Moderation",
			Settings:        map[string]string{"channel_id": "123"},
			Commands:        []ModuleCommandCreateRequest{command("ban")},
			EventListeners:  []ModuleEventListenerCreateRequest{listener("welcome")},
		}
	}

	tests := []struct {
		name    string
		change  func(r *ModuleCreateRequest)
		wantErr bool
	}{
		{"valid", func(r *ModuleCreateRequest) {}, false},
		{"no settings", func(r *ModuleCreateRequest) { r.Settings = nil }, false},
		{"no items", func(r *ModuleCreateRequest) { r.Commands, r.EventListeners = nil, nil }, true},
		{"only event listeners", func(r *ModuleCreateRequest) { r.Commands = nil }, false},
		{"duplicate key", func(r *ModuleCreateRequest) { r.EventListeners[0].Key = "ban" }, true},
		{"missing key", func(r *ModuleCreateRequest) { r.Commands[0].Key = "" }, true},
		{"missing template", func(r *ModuleCreateRequest) { r.TemplateID = "" }, true},
		{"missing version", func(r *ModuleCreateRequest) { r.TemplateVersion = 0 }, true},
		{"invalid source", func(r *ModuleCreateRequest) { r.EventListeners[0].EventListener.Source = "other" }, true},
		{"long setting", func(r *ModuleCreateRequest) { r.Settings["channel_id"] = strings.Repeat("a", 4001) }, true},
		{"too many settings", func(r *ModuleCreateRequest) {
			for i := range 51 {
				r.Settings[strings.Repeat("k", i+1)] = "v"
			}
		}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := valid()
			tt.change(&req)
			err := req.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
