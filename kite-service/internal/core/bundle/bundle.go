// Package bundle has the bundles apps can install. For now those are the
// templates of kite-web, which writes them to templates.json.
package bundle

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
	"sync"

	"gopkg.in/guregu/null.v4"
)

//go:embed templates.json
var templatesJSON []byte

type Bundle struct {
	ID             string          `json:"id"`
	Version        int             `json:"version"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	Settings       []Setting       `json:"settings"`
	Commands       []Command       `json:"commands"`
	EventListeners []EventListener `json:"event_listeners"`
	Messages       []Message       `json:"messages"`
}

type SettingType string

const (
	SettingTypeText     SettingType = "text"
	SettingTypeTextarea SettingType = "textarea"
	SettingTypeChannel  SettingType = "channel"
	SettingTypeRole     SettingType = "role"
)

type Setting struct {
	Key      string      `json:"key"`
	Label    string      `json:"label"`
	Type     SettingType `json:"type"`
	Required bool        `json:"required"`
}

// The flows and messages stay JSON until the settings are filled in.
type Command struct {
	ID         string          `json:"id"`
	FlowSource json.RawMessage `json:"flow_source"`
}

type EventListener struct {
	ID         string          `json:"id"`
	Source     string          `json:"source"`
	FlowSource json.RawMessage `json:"flow_source"`
}

type Message struct {
	ID          string                     `json:"id"`
	Name        string                     `json:"name"`
	Description null.String                `json:"description"`
	Data        json.RawMessage            `json:"data"`
	FlowSources map[string]json.RawMessage `json:"flow_sources"`
}

var bundles = sync.OnceValue(func() []*Bundle {
	var res []*Bundle
	if err := json.Unmarshal(templatesJSON, &res); err != nil {
		panic(fmt.Sprintf("failed to load bundles: %v", err))
	}
	return res
})

// Get returns the bundle with the ID, or nil if there is none.
func Get(id string) *Bundle {
	for _, b := range bundles() {
		if b.ID == id {
			return b
		}
	}
	return nil
}

// Selection is what an install adds: the chosen items of a bundle with the
// settings filled in.
type Selection struct {
	Commands       []Command
	EventListeners []EventListener
	Messages       []Message
}

var (
	settingPattern     = regexp.MustCompile(`\$setting\.([a-z0-9_]+)`)
	onlySettingPattern = regexp.MustCompile(`^\$setting\.([a-z0-9_]+)$`)
	snowflakePattern   = regexp.MustCompile(`^[0-9]{17,20}$`)
)

// Select returns the items with the keys and the settings filled in. It fails
// for a key the bundle doesn't have, a setting it doesn't declare, a channel
// or role that isn't an ID, and a missing required setting that one of the
// items uses.
func (b *Bundle) Select(keys []string, values map[string]string) (*Selection, error) {
	settings := make(map[string]Setting, len(b.Settings))
	for _, s := range b.Settings {
		settings[s.Key] = s
	}
	for key, value := range values {
		s, ok := settings[key]
		if !ok {
			return nil, fmt.Errorf("unknown setting %s", key)
		}
		if value != "" && (s.Type == SettingTypeChannel || s.Type == SettingTypeRole) && !snowflakePattern.MatchString(value) {
			return nil, fmt.Errorf("%s must be the ID of a %s", s.Label, s.Type)
		}
	}

	var used []string
	found := 0
	sel := &Selection{}
	apply := func(raw json.RawMessage) (json.RawMessage, error) {
		for _, m := range settingPattern.FindAllSubmatch(raw, -1) {
			used = append(used, string(m[1]))
		}
		return applySettings(raw, values)
	}

	for _, c := range b.Commands {
		if !slices.Contains(keys, c.ID) {
			continue
		}
		found++
		flowSource, err := apply(c.FlowSource)
		if err != nil {
			return nil, err
		}
		sel.Commands = append(sel.Commands, Command{ID: c.ID, FlowSource: flowSource})
	}
	for _, l := range b.EventListeners {
		if !slices.Contains(keys, l.ID) {
			continue
		}
		found++
		flowSource, err := apply(l.FlowSource)
		if err != nil {
			return nil, err
		}
		sel.EventListeners = append(sel.EventListeners, EventListener{ID: l.ID, Source: l.Source, FlowSource: flowSource})
	}
	for _, m := range b.Messages {
		if !slices.Contains(keys, m.ID) {
			continue
		}
		found++
		data, err := apply(m.Data)
		if err != nil {
			return nil, err
		}
		flowSources := make(map[string]json.RawMessage, len(m.FlowSources))
		for id, flowSource := range m.FlowSources {
			if flowSources[id], err = apply(flowSource); err != nil {
				return nil, err
			}
		}
		sel.Messages = append(sel.Messages, Message{ID: m.ID, Name: m.Name, Description: m.Description, Data: data, FlowSources: flowSources})
	}

	if found != len(keys) {
		return nil, fmt.Errorf("the bundle doesn't have all of the items %v", keys)
	}

	for _, s := range b.Settings {
		if s.Required && values[s.Key] == "" && slices.Contains(used, s.Key) {
			return nil, fmt.Errorf("%s is required", s.Label)
		}
	}

	return sel, nil
}

// applySettings replaces $setting.<key> in the strings of the JSON with the
// value of the setting. A field that is only an optional setting without a
// value is left out, so the block uses its default.
func applySettings(raw json.RawMessage, values map[string]string) (json.RawMessage, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	// Keeps large numbers like colors and permissions exact.
	dec.UseNumber()

	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("failed to decode bundle item: %w", err)
	}

	var apply func(v any) any
	apply = func(v any) any {
		switch v := v.(type) {
		case string:
			return settingPattern.ReplaceAllStringFunc(v, func(m string) string {
				return values[settingPattern.FindStringSubmatch(m)[1]]
			})
		case []any:
			for i := range v {
				v[i] = apply(v[i])
			}
			return v
		case map[string]any:
			for key, field := range v {
				if s, ok := field.(string); ok {
					if m := onlySettingPattern.FindStringSubmatch(s); m != nil && values[m[1]] == "" {
						delete(v, key)
						continue
					}
				}
				v[key] = apply(field)
			}
			return v
		default:
			return v
		}
	}

	return json.Marshal(apply(v))
}
