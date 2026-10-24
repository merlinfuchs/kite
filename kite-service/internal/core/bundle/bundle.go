// Package bundle has the format of the bundles apps can install. For now those
// are the official templates of kite-web, which writes them to templates.json.
package bundle

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/internal/store"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/message"
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
	Key         string      `json:"key"`
	Label       string      `json:"label"`
	Description string      `json:"description"`
	Type        SettingType `json:"type"`
	Required    bool        `json:"required"`
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
	// SendTo is $setting.<key> of the channel the message is posted in after
	// the install.
	SendTo string `json:"send_to"`
}

// Official returns the bundles of templates.json.
func Official() []*Bundle {
	res := make([]*Bundle, len(official))
	for i, o := range official {
		res[i] = o.bundle
	}
	return res
}

type officialBundle struct {
	bundle  *Bundle
	listing model.BundleListing
}

// Loaded on startup, so the service doesn't start with broken bundles.
var official = func() []officialBundle {
	var raw []json.RawMessage
	if err := json.Unmarshal(templatesJSON, &raw); err != nil {
		panic(fmt.Sprintf("failed to load bundles: %v", err))
	}

	res := make([]officialBundle, len(raw))
	for i, content := range raw {
		b, err := Parse(content)
		if err != nil {
			panic(fmt.Sprintf("failed to load bundles: %v", err))
		}
		res[i].bundle = b
		if err := json.Unmarshal(content, &res[i].listing); err != nil {
			panic(fmt.Sprintf("failed to load bundles: %v", err))
		}
	}
	return res
}()

// Parse reads the stored content of a bundle version.
func Parse(content []byte) (*Bundle, error) {
	var b Bundle
	if err := json.Unmarshal(content, &b); err != nil {
		return nil, fmt.Errorf("failed to parse bundle: %w", err)
	}
	return &b, nil
}

// SyncOfficial stores the official bundles, so they're installed from the
// store like any other. The embedded ones are the source of truth for them,
// so a stored version gets their content.
func SyncOfficial(ctx context.Context, s store.BundleStore) error {
	now := time.Now().UTC()
	for _, o := range official {
		b := o.bundle
		listing := o.listing
		listing.UpdatedAt = now
		content, err := json.Marshal(b)
		if err != nil {
			return fmt.Errorf("failed to encode bundle %s: %w", b.ID, err)
		}
		err = s.UpsertOfficialBundle(ctx, &model.Bundle{
			ID:          b.ID,
			Name:        b.Name,
			Description: b.Description,
			Version:     b.Version,
			Listing:     &listing,
			CreatedAt:   now,
			UpdatedAt:   now,
		}, &model.BundleVersion{
			BundleID:  b.ID,
			Version:   b.Version,
			Content:   content,
			CreatedAt: now,
		})
		if err != nil {
			return fmt.Errorf("failed to store bundle %s: %w", b.ID, err)
		}
	}
	return nil
}

// HasItem returns whether the bundle has a command, event listener or message
// with the key.
func (b *Bundle) HasItem(key string) bool {
	return slices.ContainsFunc(b.Commands, func(c Command) bool { return c.ID == key }) ||
		slices.ContainsFunc(b.EventListeners, func(l EventListener) bool { return l.ID == key }) ||
		slices.ContainsFunc(b.Messages, func(m Message) bool { return m.ID == key })
}

// Selection is what an install adds: the chosen items of a bundle with the
// settings filled in.
type Selection struct {
	Commands       []SelectedCommand
	EventListeners []SelectedEventListener
	Messages       []SelectedMessage
}

type SelectedCommand struct {
	ID         string
	FlowSource flow.FlowData
}

type SelectedEventListener struct {
	ID         string
	Source     string
	FlowSource flow.FlowData
}

type SelectedMessage struct {
	ID          string
	Name        string
	Description null.String
	Data        message.MessageData
	FlowSources map[string]flow.FlowData
}

var (
	settingPattern     = regexp.MustCompile(`\$setting\.[a-z0-9_]+`)
	onlySettingPattern = regexp.MustCompile(`^\$setting\.[a-z0-9_]+$`)
)

// InvalidError is a choice of items or settings that can't be installed.
type InvalidError struct{ msg string }

func (e *InvalidError) Error() string { return e.msg }

func invalid(format string, args ...any) error {
	return &InvalidError{msg: fmt.Sprintf(format, args...)}
}

// Select returns the items with the keys and the settings filled in. It fails
// with an InvalidError for a key the bundle doesn't have, a setting it doesn't
// declare, a channel or role that isn't an ID, and a missing required setting
// that one of the items uses.
func (b *Bundle) Select(keys []string, values map[string]string) (*Selection, error) {
	settings := make(map[string]Setting, len(b.Settings))
	for _, s := range b.Settings {
		settings[s.Key] = s
	}
	for key, value := range values {
		s, ok := settings[key]
		if !ok {
			return nil, invalid("unknown setting %s", key)
		}
		if value != "" && (s.Type == SettingTypeChannel || s.Type == SettingTypeRole) {
			if id, err := discord.ParseSnowflake(value); err != nil || !id.IsValid() {
				return nil, invalid("%s must be the ID of a %s", s.Label, s.Type)
			}
		}
	}

	wanted := make(map[string]bool, len(keys))
	for _, key := range keys {
		wanted[key] = true
	}
	pick := func(id string) bool {
		if !wanted[id] {
			return false
		}
		delete(wanted, id)
		return true
	}

	f := filler{values: values, used: map[string]bool{}}
	sel := &Selection{}

	for _, c := range b.Commands {
		if !pick(c.ID) {
			continue
		}
		cmd := SelectedCommand{ID: c.ID}
		if err := f.fill(c.FlowSource, &cmd.FlowSource); err != nil {
			return nil, err
		}
		sel.Commands = append(sel.Commands, cmd)
	}
	for _, l := range b.EventListeners {
		if !pick(l.ID) {
			continue
		}
		listener := SelectedEventListener{ID: l.ID, Source: l.Source}
		if err := f.fill(l.FlowSource, &listener.FlowSource); err != nil {
			return nil, err
		}
		sel.EventListeners = append(sel.EventListeners, listener)
	}
	for _, m := range b.Messages {
		if !pick(m.ID) {
			continue
		}
		msg := SelectedMessage{
			ID:          m.ID,
			Name:        m.Name,
			Description: m.Description,
			FlowSources: make(map[string]flow.FlowData, len(m.FlowSources)),
		}
		if err := f.fill(m.Data, &msg.Data); err != nil {
			return nil, err
		}
		for id, raw := range m.FlowSources {
			var flowSource flow.FlowData
			if err := f.fill(raw, &flowSource); err != nil {
				return nil, err
			}
			msg.FlowSources[id] = flowSource
		}
		f.recordUsed(m.SendTo)
		sel.Messages = append(sel.Messages, msg)
	}

	if len(wanted) > 0 {
		unknown := slices.Sorted(maps.Keys(wanted))
		return nil, invalid("the bundle has no items %s", strings.Join(unknown, ", "))
	}

	for _, s := range b.Settings {
		if s.Required && strings.TrimSpace(values[s.Key]) == "" && f.used[s.Key] {
			return nil, invalid("%s is required", s.Label)
		}
	}

	return sel, nil
}

// filler fills in the settings and remembers which ones the items use.
type filler struct {
	values map[string]string
	used   map[string]bool
}

func (f *filler) recordUsed(s string) {
	for _, m := range settingPattern.FindAllString(s, -1) {
		f.used[strings.TrimPrefix(m, "$setting.")] = true
	}
}

// fill replaces $setting.<key> in the strings of the JSON with the value of
// the setting, and decodes the result into v. A field that is only an optional
// setting without a value is left out, so the block uses its default.
func (f *filler) fill(raw json.RawMessage, v any) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	// Keeps large numbers like colors and permissions exact.
	dec.UseNumber()

	var tree any
	if err := dec.Decode(&tree); err != nil {
		return fmt.Errorf("failed to decode bundle item: %w", err)
	}

	var apply func(v any) any
	apply = func(v any) any {
		switch v := v.(type) {
		case string:
			f.recordUsed(v)
			return settingPattern.ReplaceAllStringFunc(v, func(m string) string {
				return f.values[strings.TrimPrefix(m, "$setting.")]
			})
		case []any:
			for i := range v {
				v[i] = apply(v[i])
			}
			return v
		case map[string]any:
			for key, field := range v {
				if s, ok := field.(string); ok && onlySettingPattern.MatchString(s) && f.values[strings.TrimPrefix(s, "$setting.")] == "" {
					f.recordUsed(s)
					delete(v, key)
					continue
				}
				v[key] = apply(field)
			}
			return v
		default:
			return v
		}
	}

	filled, err := json.Marshal(apply(tree))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(filled, v); err != nil {
		return fmt.Errorf("failed to decode bundle item: %w", err)
	}
	return nil
}
