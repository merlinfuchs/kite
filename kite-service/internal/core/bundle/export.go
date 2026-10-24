package bundle

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/kitecloud/kite/kite-service/internal/model"
)

// ExportItems are items of an app, with the message templates and variables
// they use.
type ExportItems struct {
	Commands       []*model.Command
	EventListeners []*model.EventListener
	Messages       []*model.Message
	Variables      []*model.Variable
}

// Target is what an ID in a block or message is in the app's servers.
type Target struct {
	Type SettingType
	Name string
}

// Resolver returns what the ID is, or false if it isn't a channel or role of
// the app's servers.
type Resolver func(id string) (Target, bool)

var snowflakePattern = regexp.MustCompile(`\b\d{17,20}\b`)

// Export builds a bundle of the items as they are in the app, to share them
// as one code. Their IDs become their keys, which the blocks already use for
// the message templates and variables. Channels and roles that resolve finds
// become settings, so the app that imports them picks its own. It returns the
// IDs that stay as they are, like users or servers.
func Export(name string, items ExportItems, resolve Resolver) (*Bundle, []string, error) {
	b := &Bundle{
		Version:        1,
		Name:           name,
		Settings:       []Setting{},
		Commands:       []Command{},
		EventListeners: []EventListener{},
		Messages:       []Message{},
		Variables:      []Variable{},
	}

	for _, cmd := range items.Commands {
		raw, err := json.Marshal(cmd.FlowSource)
		if err != nil {
			return nil, nil, err
		}
		b.Commands = append(b.Commands, Command{ID: cmd.ID, FlowSource: raw})
	}
	for _, listener := range items.EventListeners {
		raw, err := json.Marshal(listener.FlowSource)
		if err != nil {
			return nil, nil, err
		}
		b.EventListeners = append(b.EventListeners, EventListener{ID: listener.ID, Source: string(listener.Source), FlowSource: raw})
	}
	for _, msg := range items.Messages {
		data, err := json.Marshal(msg.Data)
		if err != nil {
			return nil, nil, err
		}
		flowSources := make(map[string]json.RawMessage, len(msg.FlowSources))
		for key, flowSource := range msg.FlowSources {
			if flowSources[key], err = json.Marshal(flowSource); err != nil {
				return nil, nil, err
			}
		}
		b.Messages = append(b.Messages, Message{
			ID:          msg.ID,
			Name:        msg.Name,
			Description: msg.Description,
			Data:        data,
			FlowSources: flowSources,
		})
	}
	for _, v := range items.Variables {
		b.Variables = append(b.Variables, Variable{ID: v.ID, Name: v.Name, Scoped: v.Scoped})
	}

	// The IDs in the order they're first used, so the settings are too.
	var found []string
	seen := map[string]bool{}
	err := b.eachString(func(s string) string {
		for _, id := range snowflakePattern.FindAllString(s, -1) {
			if !seen[id] {
				seen[id] = true
				found = append(found, id)
			}
		}
		return s
	})
	if err != nil {
		return nil, nil, err
	}

	keys := map[string]string{}
	var unmatched []string
	for _, id := range found {
		target, ok := resolve(id)
		if !ok {
			unmatched = append(unmatched, id)
			continue
		}
		key := settingKey(target, b.Settings)
		keys[id] = key
		b.Settings = append(b.Settings, settingFor(key, target))
	}

	if len(keys) > 0 {
		err = b.eachString(func(s string) string {
			return snowflakePattern.ReplaceAllStringFunc(s, func(id string) string {
				if key, ok := keys[id]; ok {
					return "$setting." + key
				}
				return id
			})
		})
		if err != nil {
			return nil, nil, err
		}
	}

	return b, unmatched, nil
}

func settingFor(key string, target Target) Setting {
	s := Setting{Key: key, Type: target.Type, Required: true}
	if target.Type == SettingTypeChannel {
		s.Label = "#" + target.Name
		s.Description = fmt.Sprintf("The channel to use where the exported app used #%s.", target.Name)
	} else {
		s.Label = "@" + target.Name
		s.Description = fmt.Sprintf("The role to use where the exported app used @%s.", target.Name)
	}
	return s
}

var nonKeyPattern = regexp.MustCompile(`[^a-z0-9]+`)

// settingKey makes a key from the name that the other settings don't have.
func settingKey(target Target, settings []Setting) string {
	base := strings.Trim(nonKeyPattern.ReplaceAllString(strings.ToLower(target.Name), "_"), "_")
	if base == "" {
		base = string(target.Type)
	} else {
		base = string(target.Type) + "_" + base
	}
	key := base
	for i := 2; ; i++ {
		taken := false
		for _, s := range settings {
			if s.Key == key {
				taken = true
				break
			}
		}
		if !taken {
			return key
		}
		key = fmt.Sprintf("%s_%d", base, i)
	}
}

// eachString replaces every string in the blocks and messages with what fn
// returns for it.
func (b *Bundle) eachString(fn func(string) string) error {
	var err error
	for i := range b.Commands {
		if b.Commands[i].FlowSource, err = mapStrings(b.Commands[i].FlowSource, fn); err != nil {
			return err
		}
	}
	for i := range b.EventListeners {
		if b.EventListeners[i].FlowSource, err = mapStrings(b.EventListeners[i].FlowSource, fn); err != nil {
			return err
		}
	}
	for i := range b.Messages {
		msg := &b.Messages[i]
		if msg.Data, err = mapStrings(msg.Data, fn); err != nil {
			return err
		}
		for key, flowSource := range msg.FlowSources {
			if msg.FlowSources[key], err = mapStrings(flowSource, fn); err != nil {
				return err
			}
		}
	}
	return nil
}

func mapStrings(raw json.RawMessage, fn func(string) string) (json.RawMessage, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	// Keeps large numbers like colors and permissions exact.
	dec.UseNumber()

	var tree any
	if err := dec.Decode(&tree); err != nil {
		return nil, fmt.Errorf("failed to decode bundle item: %w", err)
	}

	var apply func(v any) any
	apply = func(v any) any {
		switch v := v.(type) {
		case string:
			return fn(v)
		case []any:
			for i := range v {
				v[i] = apply(v[i])
			}
			return v
		case map[string]any:
			for key, field := range v {
				v[key] = apply(field)
			}
			return v
		default:
			return v
		}
	}

	return json.Marshal(apply(tree))
}
