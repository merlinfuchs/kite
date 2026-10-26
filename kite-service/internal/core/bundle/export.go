package bundle

import (
	"encoding/json"

	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
)

// ExportItems are items of an app, with the message templates and variables
// they use.
type ExportItems struct {
	Commands       []*model.Command
	EventListeners []*model.EventListener
	Messages       []*model.Message
	Variables      []*model.Variable
}

// Export builds a bundle of the items as they are in the app, to share them
// as one code. Their IDs become their keys, which the blocks already use for
// the message templates and variables. The parameters of the flows become the
// bundle's, without their values, so whoever adds it fills in their own.
func Export(name string, items ExportItems) (*Bundle, error) {
	b := &Bundle{
		Version:        1,
		Name:           name,
		Parameters:     []Parameter{},
		Commands:       []Command{},
		EventListeners: []EventListener{},
		Messages:       []Message{},
		Variables:      []Variable{},
	}

	// exportFlow leaves out the values of the parameters, and adds the ones
	// the blocks use without defining them.
	exportFlow := func(data flow.FlowData) (json.RawMessage, error) {
		data.Parameters = data.UsedParameters()
		for i := range data.Parameters {
			data.Parameters[i].Value = ""
			b.addParameter(data.Parameters[i])
		}
		return json.Marshal(data)
	}

	for _, cmd := range items.Commands {
		raw, err := exportFlow(cmd.FlowSource)
		if err != nil {
			return nil, err
		}
		b.Commands = append(b.Commands, Command{ID: cmd.ID, FlowSource: raw})
	}
	for _, listener := range items.EventListeners {
		raw, err := exportFlow(listener.FlowSource)
		if err != nil {
			return nil, err
		}
		b.EventListeners = append(b.EventListeners, EventListener{ID: listener.ID, Source: string(listener.Source), FlowSource: raw})
	}
	for _, msg := range items.Messages {
		// Attachments are assets of the app, which other apps can't use.
		msgData := msg.Data
		msgData.Attachments = nil
		data, err := json.Marshal(msgData)
		if err != nil {
			return nil, err
		}
		flowSources := make(map[string]json.RawMessage, len(msg.FlowSources))
		for key, flowSource := range msg.FlowSources {
			if flowSources[key], err = exportFlow(flowSource); err != nil {
				return nil, err
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

	return b, nil
}

// addParameter adds the parameter of a flow to the bundle, unless one with the
// key is there already. Flows that use the same key share the value.
func (b *Bundle) addParameter(p flow.FlowParameter) {
	for _, existing := range b.Parameters {
		if existing.Key == p.Key {
			return
		}
	}
	b.Parameters = append(b.Parameters, Parameter{
		Key:         p.Key,
		Label:       p.DisplayLabel(),
		Description: p.Description,
		Type:        p.Type,
		Required:    p.Required,
	})
}
