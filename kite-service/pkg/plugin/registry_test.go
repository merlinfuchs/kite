package plugin

import (
	"context"
	"slices"
	"testing"

	"github.com/diamondburned/arikawa/v3/utils/ws"
)

type stubPlugin struct {
	Plugin
	id     string
	events []Event
}

func (p *stubPlugin) ID() string          { return p.id }
func (p *stubPlugin) Events() []Event     { return p.events }
func (p *stubPlugin) Commands() []Command { return nil }
func (p *stubPlugin) Instance(context.Context, string, ConfigValues) (PluginInstance, error) {
	return nil, nil
}

func testRegistry() *Registry {
	r := NewRegistry()
	r.Register(
		&stubPlugin{id: "starboard", events: []Event{
			{ID: "event_message_reaction_add", Source: EventSourceDiscord, Type: EventTypeMessageReactionAdd},
		}},
		&stubPlugin{id: "counting", events: []Event{
			{ID: "event_message_create", Source: EventSourceDiscord, Type: EventTypeMessageCreate},
		}},
	)
	return r
}

func TestEventTypesForResources(t *testing.T) {
	tests := []struct {
		name      string
		resources []string
		want      []ws.EventType
	}{
		{
			name:      "no resources",
			resources: nil,
			want:      nil,
		},
		{
			name:      "resolves a single plugin event",
			resources: []string{"starboard:event_message_reaction_add"},
			want:      []ws.EventType{"MESSAGE_REACTION_ADD"},
		},
		{
			name:      "resolves across plugins",
			resources: []string{"starboard:event_message_reaction_add", "counting:event_message_create"},
			want:      []ws.EventType{"MESSAGE_REACTION_ADD", "MESSAGE_CREATE"},
		},
		{
			// Enabled resources include commands, which subscribe to nothing.
			name:      "ignores resources that are not events",
			resources: []string{"starboard:cmd_starboard"},
			want:      nil,
		},
		{
			// A plugin can be dropped from the registry while instances of it
			// still exist in the database.
			name:      "ignores unknown plugins",
			resources: []string{"removed_plugin:event_message_create"},
			want:      nil,
		},
		{
			name:      "ignores malformed pairs",
			resources: []string{"no_separator"},
			want:      nil,
		},
		{
			name: "deduplicates event types",
			resources: []string{
				"counting:event_message_create",
				"counting:event_message_create",
			},
			want: []ws.EventType{"MESSAGE_CREATE"},
		},
	}

	registry := testRegistry()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := registry.EventTypesForResources(tt.resources)

			if len(got) != len(tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			for _, want := range tt.want {
				if !slices.Contains(got, want) {
					t.Errorf("got %v, missing %q", got, want)
				}
			}
		})
	}
}
