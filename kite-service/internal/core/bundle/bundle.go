// Package bundle has the format of the bundles apps can install. For now those
// are the official templates of kite-web, which writes them to templates.json.
package bundle

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
	"time"

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
	Parameters     []Parameter     `json:"parameters"`
	Commands       []Command       `json:"commands"`
	EventListeners []EventListener `json:"event_listeners"`
	Messages       []Message       `json:"messages"`
	Variables      []Variable      `json:"variables"`
}

// Variable is a stored variable the blocks of the bundle use by its ID.
type Variable struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Scoped bool   `json:"scoped"`
}

// Parameter is a parameter the flows of the bundle use as {{params.<key>}}.
// Flows that use the same key get the same value when the bundle is added.
type Parameter struct {
	Key         string                 `json:"key"`
	Label       string                 `json:"label"`
	Description string                 `json:"description"`
	Type        flow.FlowParameterType `json:"type"`
	Required    bool                   `json:"required"`
}

// flowParameter is the parameter as the flows that use it get it.
func (p Parameter) flowParameter() flow.FlowParameter {
	return flow.FlowParameter{
		Key:         p.Key,
		Label:       p.Label,
		Description: p.Description,
		Type:        p.Type,
		Required:    p.Required,
	}
}

// The flows and messages stay JSON until they're selected.
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
	// SendTo is the key of the channel parameter the message is posted in
	// after the install.
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
	for i, o := range official {
		b := o.bundle
		listing := o.listing
		listing.Position = i
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
			Content:     content,
			Listing:     &listing,
			CreatedAt:   now,
			UpdatedAt:   now,
		})
		if err != nil {
			return fmt.Errorf("failed to store bundle %s: %w", b.ID, err)
		}
	}
	return nil
}

// HasItem returns whether the bundle has a command, event listener, message or
// variable with the key.
func (b *Bundle) HasItem(key string) bool {
	return slices.ContainsFunc(b.Commands, func(c Command) bool { return c.ID == key }) ||
		slices.ContainsFunc(b.EventListeners, func(l EventListener) bool { return l.ID == key }) ||
		slices.ContainsFunc(b.Messages, func(m Message) bool { return m.ID == key }) ||
		slices.ContainsFunc(b.Variables, func(v Variable) bool { return v.ID == key })
}

// Message returns the message with the key.
func (b *Bundle) Message(key string) (Message, bool) {
	i := slices.IndexFunc(b.Messages, func(m Message) bool { return m.ID == key })
	if i == -1 {
		return Message{}, false
	}
	return b.Messages[i], true
}

// Variable returns the variable with the key.
func (b *Bundle) Variable(key string) (Variable, bool) {
	i := slices.IndexFunc(b.Variables, func(v Variable) bool { return v.ID == key })
	if i == -1 {
		return Variable{}, false
	}
	return b.Variables[i], true
}

// Flows returns the flows of the selected items, including the ones of the
// buttons of messages.
func (s *Selection) Flows() []flow.FlowData {
	var flows []flow.FlowData
	for _, cmd := range s.Commands {
		flows = append(flows, cmd.FlowSource)
	}
	for _, listener := range s.EventListeners {
		flows = append(flows, listener.FlowSource)
	}
	for _, msg := range s.Messages {
		for _, flowSource := range msg.FlowSources {
			flows = append(flows, flowSource)
		}
	}
	return flows
}

// Selection is what an install adds: the chosen items of a bundle with the
// values of their parameters.
type Selection struct {
	Commands       []SelectedCommand
	EventListeners []SelectedEventListener
	Messages       []SelectedMessage
	Variables      []Variable
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

// InvalidError is a choice of items or parameters that can't be installed.
type InvalidError struct{ msg string }

func (e *InvalidError) Error() string { return e.msg }

func invalid(format string, args ...any) error {
	return &InvalidError{msg: fmt.Sprintf(format, args...)}
}

// Select returns the items with the keys, with the values of the parameters
// their flows use. It fails with an InvalidError for a key the bundle doesn't
// have, a parameter it doesn't declare, a channel or role that isn't an ID,
// and a missing required parameter that one of the items uses.
func (b *Bundle) Select(keys []string, values map[string]string) (*Selection, error) {
	params := make(map[string]flow.FlowParameter, len(b.Parameters))
	for _, p := range b.Parameters {
		params[p.Key] = p.flowParameter()
	}
	for key, value := range values {
		p, ok := params[key]
		if !ok {
			return nil, invalid("unknown parameter %s", key)
		}
		if err := p.CheckValue(value); err != nil {
			return nil, invalid("%s", err)
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

	used := map[string]bool{}
	// decodeFlow sets the parameters of the flow to the ones it uses, with
	// their values.
	decodeFlow := func(raw json.RawMessage) (flow.FlowData, error) {
		var data flow.FlowData
		if err := json.Unmarshal(raw, &data); err != nil {
			return data, fmt.Errorf("failed to decode bundle flow: %w", err)
		}
		data.Parameters = data.UsedParameters()
		for i, p := range data.Parameters {
			used[p.Key] = true
			if declared, ok := params[p.Key]; ok {
				p = declared
			}
			p.Value = values[p.Key]
			data.Parameters[i] = p
		}
		return data, nil
	}

	sel := &Selection{}
	var err error

	for _, c := range b.Commands {
		if !pick(c.ID) {
			continue
		}
		cmd := SelectedCommand{ID: c.ID}
		if cmd.FlowSource, err = decodeFlow(c.FlowSource); err != nil {
			return nil, err
		}
		sel.Commands = append(sel.Commands, cmd)
	}
	for _, l := range b.EventListeners {
		if !pick(l.ID) {
			continue
		}
		listener := SelectedEventListener{ID: l.ID, Source: l.Source}
		if listener.FlowSource, err = decodeFlow(l.FlowSource); err != nil {
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
		if err := json.Unmarshal(m.Data, &msg.Data); err != nil {
			return nil, fmt.Errorf("failed to decode bundle message: %w", err)
		}
		for id, raw := range m.FlowSources {
			if msg.FlowSources[id], err = decodeFlow(raw); err != nil {
				return nil, err
			}
		}
		if m.SendTo != "" {
			used[m.SendTo] = true
		}
		sel.Messages = append(sel.Messages, msg)
	}

	for _, v := range b.Variables {
		if pick(v.ID) {
			sel.Variables = append(sel.Variables, v)
		}
	}

	if len(wanted) > 0 {
		unknown := slices.Sorted(maps.Keys(wanted))
		return nil, invalid("the bundle has no items %s", strings.Join(unknown, ", "))
	}

	for _, p := range b.Parameters {
		if p.Required && strings.TrimSpace(values[p.Key]) == "" && used[p.Key] {
			return nil, invalid("%s is required", p.flowParameter().DisplayLabel())
		}
	}

	return sel, nil
}
