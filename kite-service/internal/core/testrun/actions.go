package testrun

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/kitecloud/kite/kite-service/internal/model"
	"github.com/kitecloud/kite/kite-service/pkg/flow"
	"github.com/kitecloud/kite/kite-service/pkg/message"
)

// act does what the user does after the flow ran, see Action.
func (s *testSession) act(ctx context.Context, action Action) error {
	userID := s.userID
	if action.UserID != "" {
		id, err := discord.ParseSnowflake(action.UserID)
		if err != nil {
			return testErrorf("user_id must be the ID of a Discord user")
		}
		userID = discord.UserID(id)
	}

	switch {
	case action.Submit != nil:
		return s.submitModal(ctx, userID, action.Submit)
	case action.Click != "":
		return s.useComponent(ctx, userID, action.Click, false, nil)
	case action.Select != "":
		if len(action.Values) == 0 {
			return testErrorf("values are required to use select menu %q", action.Select)
		}
		return s.useComponent(ctx, userID, action.Select, true, action.Values)
	}
	return testErrorf("a step needs click, select or submit")
}

func (s *testSession) useComponent(ctx context.Context, userID discord.UserID, target string, isSelect bool, values []string) error {
	msg, component, ok := s.findComponent(target, isSelect)
	if !ok {
		kind := "button"
		if isSelect {
			kind = "select menu"
		}
		available := s.availableComponents()
		if len(available) == 0 {
			return testErrorf("no %s %q, the flow sent no message with buttons or select menus", kind, target)
		}
		return testErrorf("no %s %q, the messages the flow sent have: %s", kind, target, strings.Join(available, ", "))
	}

	data := map[string]any{
		"custom_id":      component.CustomID,
		"component_type": component.Type,
	}
	// Named like users see it, the target can be a custom ID, which the
	// editor shows to users.
	name := "the button"
	if isSelect {
		name = "the select menu"
	}
	if n := cmp.Or(component.Label, component.Placeholder, component.Emoji); n != "" {
		name = strconv.Quote(n)
	}
	label := "Click " + name
	if isSelect {
		label = fmt.Sprintf("Select %s in %s", strings.Join(values, ", "), name)
		data["values"] = values
		if component.Type != int(discord.StringSelectComponentType) {
			resolved := map[string]any{}
			if err := s.resolveSelect(resolved, component.Type, values); err != nil {
				return testErrorf("%v", err)
			}
			data["resolved"] = resolved
		}
	}

	guildID, channelID := msg.GuildID, msg.ChannelID
	if !channelID.IsValid() {
		guildID, channelID = s.guildID, s.channelID
	}
	event, err := s.interactionEvent(userID, guildID, channelID, discord.ComponentInteractionType, data, &msg, nil)
	if err != nil {
		return err
	}

	if rpID, _, ok := message.DecodeCustomIDMessageComponentResumePoint(component.CustomID); ok {
		return s.resume(ctx, label, rpID, event)
	}

	tmpl, err := s.messageTemplate(ctx, msg.ID)
	if err != nil {
		return err
	}
	if tmpl == nil {
		return testErrorf("%q is on a message no flow handles", target)
	}
	source, ok := tmpl.FlowSources[component.CustomID]
	if !ok {
		return testErrorf("%q has no flow in message %s", target, tmpl.ID)
	}
	root, err := flow.CompileComponentButton(source)
	if err != nil {
		return testErrorf("the flow of %q is invalid: %v", target, err)
	}
	if tmpl.ID+"/"+component.CustomID == s.testedFlow {
		s.mu.Lock()
		s.testedRoots[root] = true
		s.mu.Unlock()
	}
	s.execute(ctx, label, root, root, event, nil, nil)
	return nil
}

func (s *testSession) submitModal(ctx context.Context, userID discord.UserID, values map[string]json.RawMessage) error {
	s.mu.Lock()
	modal := s.modal
	s.modal = nil
	s.mu.Unlock()
	if modal == nil || modal.CustomID == nil {
		return testErrorf("no modal is open to submit")
	}

	components, resolved, err := s.modalSubmission(modal, values)
	if err != nil {
		return err
	}

	data := map[string]any{
		"custom_id":  modal.CustomID.Val,
		"components": components,
	}
	if len(resolved) > 0 {
		data["resolved"] = resolved
	}
	event, err := s.interactionEvent(userID, s.guildID, s.channelID, discord.ModalInteractionType, data, nil, nil)
	if err != nil {
		return err
	}

	rpID, _ := message.DecodeCustomIDModalResumePoint(modal.CustomID.Val)
	return s.resume(ctx, "Submit the modal", rpID, event)
}

// resume continues the flow that waits at a resume point, like clicks and
// submits do.
func (s *testSession) resume(ctx context.Context, label string, rpID string, event gateway.Event) error {
	s.mu.Lock()
	rp, ok := s.resumePoints[rpID]
	s.mu.Unlock()
	if !ok {
		return testErrorf("%s: no flow is waiting for it", label)
	}
	node := rp.root.FindChildWithID(rp.rp.NodeID, true)
	if node == nil {
		return testErrorf("%s: the block that waits for it doesn't exist", label)
	}

	// A click or submit starts a new execution, like in resumeFlow.
	state := rp.rp.State.Copy()
	state.DurableSleeps = 0
	s.execute(ctx, label, rp.root, node, event, &state, nil)
	return nil
}

// messageTemplate returns the template a message was sent from, nil if none.
func (s *testSession) messageTemplate(ctx context.Context, messageID discord.MessageID) (*model.Message, error) {
	s.mu.Lock()
	id := s.templates[messageID]
	tmpl := s.templateCache[id]
	s.mu.Unlock()
	if id == "" || tmpl != nil {
		return tmpl, nil
	}

	tmpl, err := s.env.MessageStore.Message(ctx, s.appID, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get message template: %w", err)
	}
	s.mu.Lock()
	s.templateCache[id] = tmpl
	s.mu.Unlock()
	return tmpl, nil
}

// modalSubmission fills in the inputs of the modal with values, in the format
// Discord sends a submitted modal in.
func (s *testSession) modalSubmission(modal *api.InteractionResponseData, values map[string]json.RawMessage) ([]any, map[string]any, error) {
	var sent []map[string]any
	if modal.Components != nil {
		b, err := json.Marshal(modal.Components)
		if err != nil {
			return nil, nil, err
		}
		if err := json.Unmarshal(b, &sent); err != nil {
			return nil, nil, err
		}
	}

	used := make(map[string]bool)
	resolved := make(map[string]any)
	var inputs []string

	fill := func(input map[string]any, label string) (map[string]any, error) {
		customID, _ := input["custom_id"].(string)
		if label == "" {
			label, _ = input["label"].(string)
		}
		inputs = append(inputs, fmt.Sprintf("%s (%s)", customID, label))

		raw, ok := values[customID]
		key := customID
		if !ok {
			raw, ok = values[label]
			key = label
		}
		if ok {
			used[key] = true
		}

		typ, _ := input["type"].(float64)
		res := map[string]any{"type": typ, "custom_id": customID}
		switch discord.ComponentType(typ) {
		case discord.TextInputComponentType, discord.RadioGroupComponentType:
			value := ""
			if ok {
				if err := json.Unmarshal(raw, &value); err != nil {
					return nil, fmt.Errorf("the value of %s must be text", key)
				}
			}
			res["value"] = value
		case discord.CheckboxComponentType:
			value := false
			if ok {
				if err := json.Unmarshal(raw, &value); err != nil {
					return nil, fmt.Errorf("the value of %s must be true or false", key)
				}
			}
			res["value"] = value
		default:
			picked := []string{}
			if ok {
				if err := json.Unmarshal(raw, &picked); err != nil {
					var one string
					if json.Unmarshal(raw, &one) != nil {
						return nil, fmt.Errorf("the value of %s must be a list of values", key)
					}
					picked = []string{one}
				}
			}
			res["values"] = picked
			if discord.ComponentType(typ) != discord.StringSelectComponentType && discord.ComponentType(typ) != discord.CheckboxGroupComponentType && len(picked) > 0 {
				if err := s.resolveSelect(resolved, int(typ), picked); err != nil {
					return nil, err
				}
			}
		}
		return res, nil
	}

	var res []any
	for _, c := range sent {
		switch typ, _ := c["type"].(float64); discord.ComponentType(typ) {
		case discord.LabelComponentType:
			input, _ := c["component"].(map[string]any)
			if input == nil {
				continue
			}
			label, _ := c["label"].(string)
			filled, err := fill(input, label)
			if err != nil {
				return nil, nil, testErrorf("%v", err)
			}
			res = append(res, map[string]any{"type": typ, "component": filled})
		case discord.ActionRowComponentType:
			items, _ := c["components"].([]any)
			var row []any
			for _, item := range items {
				input, _ := item.(map[string]any)
				filled, err := fill(input, "")
				if err != nil {
					return nil, nil, testErrorf("%v", err)
				}
				row = append(row, filled)
			}
			res = append(res, map[string]any{"type": typ, "components": row})
		}
	}

	for key := range values {
		if !used[key] {
			return nil, nil, testErrorf("the modal has no input %q, it has: %s", key, strings.Join(inputs, ", "))
		}
	}
	return res, resolved, nil
}

// testComponent is a button or select menu of a sent message.
type testComponent struct {
	CustomID    string
	Type        int
	Label       string
	Placeholder string
	Emoji       string
}

// findComponent finds the button or select menu with the label, placeholder
// or custom ID in the messages the flows sent, newest first.
func (s *testSession) findComponent(target string, isSelect bool) (discord.Message, testComponent, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := len(s.messages) - 1; i >= 0; i-- {
		for _, c := range messageComponents(s.messages[i]) {
			if isSelect != (c.Type != int(discord.ButtonComponentType)) {
				continue
			}
			if c.CustomID == target || (c.Label != "" && c.Label == target) || (c.Placeholder != "" && c.Placeholder == target) {
				return s.messages[i], c, true
			}
		}
	}
	return discord.Message{}, testComponent{}, false
}

func (s *testSession) availableComponents() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	var res []string
	for i := len(s.messages) - 1; i >= 0; i-- {
		for _, c := range messageComponents(s.messages[i]) {
			if c.Type == int(discord.ButtonComponentType) {
				res = append(res, fmt.Sprintf("button %q", c.Label))
			} else {
				res = append(res, fmt.Sprintf("select menu %q (%s)", c.Placeholder, c.CustomID))
			}
		}
	}
	return res
}

// messageComponents returns the buttons and select menus of a message that
// run flows. Link buttons don't.
func messageComponents(msg discord.Message) []testComponent {
	b, err := json.Marshal(msg.Components)
	if err != nil {
		return nil
	}
	var tree any
	if err := json.Unmarshal(b, &tree); err != nil {
		return nil
	}

	var res []testComponent
	var walk func(v any)
	walk = func(v any) {
		switch v := v.(type) {
		case []any:
			for _, item := range v {
				walk(item)
			}
		case map[string]any:
			typ, _ := v["type"].(float64)
			customID, _ := v["custom_id"].(string)
			if customID != "" && discord.ComponentType(typ) != discord.TextInputComponentType {
				label, _ := v["label"].(string)
				placeholder, _ := v["placeholder"].(string)
				emoji, _ := v["emoji"].(map[string]any)
				emojiName, _ := emoji["name"].(string)
				res = append(res, testComponent{CustomID: customID, Type: int(typ), Label: label, Placeholder: placeholder, Emoji: emojiName})
			}
			for _, child := range v {
				walk(child)
			}
		}
	}
	walk(tree)
	return res
}
