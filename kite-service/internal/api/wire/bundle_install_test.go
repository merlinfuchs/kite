package wire

import (
	"strings"
	"testing"

	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/stretchr/testify/assert"
)

func TestBundleInstallCreateRequestValidate(t *testing.T) {
	flowSource := flow.FlowData{Nodes: []flow.FlowNode{{
		ID:   "1",
		Type: flow.FlowNodeTypeEntryCommand,
		Data: flow.FlowNodeData{Name: "ban", Description: "Ban a user."},
	}}}
	command := func(key string) BundleInstallCommandCreateRequest {
		return BundleInstallCommandCreateRequest{Key: key, Command: CommandCreateRequest{FlowSource: flowSource}}
	}
	listener := func(key string) BundleInstallEventListenerCreateRequest {
		return BundleInstallEventListenerCreateRequest{Key: key, EventListener: EventListenerCreateRequest{Source: "discord", FlowSource: flowSource}}
	}
	valid := func() BundleInstallCreateRequest {
		return BundleInstallCreateRequest{
			BundleID:       "moderation",
			BundleVersion:  1,
			Name:           "Moderation",
			Settings:       map[string]string{"channel_id": "123"},
			Commands:       []BundleInstallCommandCreateRequest{command("ban")},
			EventListeners: []BundleInstallEventListenerCreateRequest{listener("welcome")},
		}
	}

	tests := []struct {
		name    string
		change  func(r *BundleInstallCreateRequest)
		wantErr bool
	}{
		{"valid", func(r *BundleInstallCreateRequest) {}, false},
		{"no settings", func(r *BundleInstallCreateRequest) { r.Settings = nil }, false},
		{"no items", func(r *BundleInstallCreateRequest) { r.Commands, r.EventListeners = nil, nil }, true},
		{"only event listeners", func(r *BundleInstallCreateRequest) { r.Commands = nil }, false},
		{"duplicate key", func(r *BundleInstallCreateRequest) { r.EventListeners[0].Key = "ban" }, true},
		{"missing key", func(r *BundleInstallCreateRequest) { r.Commands[0].Key = "" }, true},
		{"missing bundle", func(r *BundleInstallCreateRequest) { r.BundleID = "" }, true},
		{"missing version", func(r *BundleInstallCreateRequest) { r.BundleVersion = 0 }, true},
		{"invalid source", func(r *BundleInstallCreateRequest) { r.EventListeners[0].EventListener.Source = "other" }, true},
		{"long setting", func(r *BundleInstallCreateRequest) { r.Settings["channel_id"] = strings.Repeat("a", 4001) }, true},
		{"too many settings", func(r *BundleInstallCreateRequest) {
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
