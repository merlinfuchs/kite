package bundle

import (
	"cmp"
	"encoding/json"
	"fmt"
	"regexp"
	"slices"
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

	// The setting of each ID, or "" for one that stays as it is.
	keys := map[string]string{}
	var unmatched []string
	err := b.eachString(func(s string) string {
		return snowflakePattern.ReplaceAllStringFunc(s, func(id string) string {
			key, seen := keys[id]
			if !seen {
				if target, ok := resolve(id); ok {
					key = settingKey(target, b.Settings)
					b.Settings = append(b.Settings, settingFor(key, target))
				} else {
					unmatched = append(unmatched, id)
				}
				keys[id] = key
			}
			if key == "" {
				return id
			}
			return "$setting." + key
		})
	})
	if err != nil {
		return nil, nil, err
	}

	// The strings are found in no particular order.
	slices.SortFunc(b.Settings, func(a, b Setting) int {
		return cmp.Or(cmp.Compare(a.Type, b.Type), cmp.Compare(a.Label, b.Label))
	})
	slices.Sort(unmatched)

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
	for i := 2; slices.ContainsFunc(settings, func(s Setting) bool { return s.Key == key }); i++ {
		key = fmt.Sprintf("%s_%d", base, i)
	}
	return key
}

// eachString replaces every string in the blocks and messages with what fn
// returns for it.
func (b *Bundle) eachString(fn func(string) string) error {
	keep := func(s string) (string, bool) { return fn(s), true }
	var err error
	for i := range b.Commands {
		if b.Commands[i].FlowSource, err = mapStrings(b.Commands[i].FlowSource, keep); err != nil {
			return err
		}
	}
	for i := range b.EventListeners {
		if b.EventListeners[i].FlowSource, err = mapStrings(b.EventListeners[i].FlowSource, keep); err != nil {
			return err
		}
	}
	for i := range b.Messages {
		msg := &b.Messages[i]
		if msg.Data, err = mapStrings(msg.Data, keep); err != nil {
			return err
		}
		for key, flowSource := range msg.FlowSources {
			if msg.FlowSources[key], err = mapStrings(flowSource, keep); err != nil {
				return err
			}
		}
	}
	return nil
}
