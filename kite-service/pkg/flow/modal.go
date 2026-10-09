package flow

import (
	"encoding/json"
	"fmt"

	"github.com/diamondburned/arikawa/v3/discord"
)

// Discord's limits for modals.
const (
	modalMaxComponents     = 5
	modalMaxSelectOptions  = 25
	modalMaxGroupOptions   = 10
	modalMinRadioOptions   = 2
	modalLabelMaxLength    = 45
	modalDescriptionMaxLen = 100

	modalTextInputPlaceholderMaxLen = 100
	modalSelectPlaceholderMaxLen    = 150
)

// modalComponent is a modal component sent as the given JSON. It's built by
// hand because the arikawa types drop required: false, which Discord defaults
// to true, so an optional select menu couldn't be sent. The embedded
// component only gives it the component interfaces.
type modalComponent struct {
	discord.TopLevelComponent
	json map[string]any
}

func (c *modalComponent) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.json)
}

// normalizeModalComponents turns modals saved before labels existed into the
// current shape. Their components are rows of text inputs that carry their
// own label, which become one label per input.
func normalizeModalComponents(components []ModalComponentData) []ModalComponentData {
	res := make([]ModalComponentData, 0, len(components))
	for _, c := range components {
		if c.Type != "" {
			res = append(res, c)
			continue
		}

		for _, input := range c.Components {
			label := input.Label
			input.Label = ""
			if input.Type == "" {
				input.Type = ModalComponentTypeTextInput
			}

			res = append(res, ModalComponentData{
				Type:       ModalComponentTypeLabel,
				Label:      label,
				Components: []ModalComponentData{input},
			})
		}
	}
	return res
}

// buildModalComponents evaluates the modal's components and returns them in
// the shape Discord expects. Custom IDs aren't evaluated, as input() looks
// values up by them.
func buildModalComponents(ctx *FlowContext, components []ModalComponentData) (discord.TopLevelComponents, error) {
	components = normalizeModalComponents(components)
	if len(components) == 0 {
		return nil, fmt.Errorf("modal must have at least one component")
	}
	if len(components) > modalMaxComponents {
		return nil, fmt.Errorf("modal has %d components, the maximum is %d", len(components), modalMaxComponents)
	}

	hasInput := false
	customIDs := make(map[string]bool, len(components))
	res := make(discord.TopLevelComponents, 0, len(components))
	for _, c := range components {
		switch c.Type {
		case ModalComponentTypeTextDisplay:
			content, err := ctx.EvalTemplateKeepSpace(c.Content)
			if err != nil {
				return nil, err
			}
			if content.String() == "" {
				return nil, fmt.Errorf("modal text display is empty")
			}

			res = append(res, &modalComponent{
				TopLevelComponent: &discord.TextDisplayComponent{},
				json: map[string]any{
					"type":    discord.TextDisplayComponentType,
					"content": content.String(),
				},
			})
		case ModalComponentTypeLabel:
			if len(c.Components) != 1 {
				return nil, fmt.Errorf("modal label must hold exactly one input, got %d", len(c.Components))
			}

			label, err := ctx.EvalTemplate(c.Label)
			if err != nil {
				return nil, err
			}
			if label.String() == "" {
				return nil, fmt.Errorf("modal label is empty")
			}
			if len([]rune(label.String())) > modalLabelMaxLength {
				return nil, fmt.Errorf("modal label is longer than %d characters", modalLabelMaxLength)
			}

			description, err := ctx.EvalTemplate(c.Description)
			if err != nil {
				return nil, err
			}
			if len([]rune(description.String())) > modalDescriptionMaxLen {
				return nil, fmt.Errorf("modal label description is longer than %d characters", modalDescriptionMaxLen)
			}

			input, err := buildModalInput(ctx, c.Components[0])
			if err != nil {
				return nil, err
			}

			customID := c.Components[0].CustomID
			if customIDs[customID] {
				return nil, fmt.Errorf("modal has more than one input with the identifier %q", customID)
			}
			customIDs[customID] = true

			l := map[string]any{
				"type":      discord.LabelComponentType,
				"label":     label.String(),
				"component": input,
			}
			if description.String() != "" {
				l["description"] = description.String()
			}

			hasInput = true
			res = append(res, &modalComponent{
				TopLevelComponent: &discord.LabelComponent{},
				json:              l,
			})
		default:
			return nil, fmt.Errorf("unknown modal component type %q", c.Type)
		}
	}

	if !hasInput {
		return nil, fmt.Errorf("modal must have at least one input")
	}

	return res, nil
}

func buildModalInput(ctx *FlowContext, c ModalComponentData) (map[string]any, error) {
	if c.CustomID == "" {
		return nil, fmt.Errorf("modal input has no identifier")
	}

	res := map[string]any{
		"custom_id": c.CustomID,
	}

	evalPlaceholder := func(maxLen int) error {
		placeholder, err := ctx.EvalTemplate(c.Placeholder)
		if err != nil {
			return err
		}
		if len([]rune(placeholder.String())) > maxLen {
			return fmt.Errorf("modal input placeholder is longer than %d characters", maxLen)
		}
		if placeholder.String() != "" {
			res["placeholder"] = placeholder.String()
		}
		return nil
	}

	// setValueLimits sets min_values and max_values. defaultMax is what
	// Discord uses when max_values isn't set.
	setValueLimits := func(max, defaultMax int) error {
		if c.MinValues < 0 || c.MinValues > max {
			return fmt.Errorf("minimum values must be between 0 and %d, got %d", max, c.MinValues)
		}
		if c.MaxValues < 0 || c.MaxValues > max {
			return fmt.Errorf("maximum values must be between 1 and %d, got %d", max, c.MaxValues)
		}
		if maxValues := maxValuesOrDefault(c.MaxValues, defaultMax); c.MinValues > maxValues {
			return fmt.Errorf("minimum values %d is more than the maximum %d", c.MinValues, maxValues)
		}
		if c.MinValues != 0 {
			res["min_values"] = c.MinValues
		}
		if c.MaxValues != 0 {
			res["max_values"] = c.MaxValues
		}
		return nil
	}

	checkDefaults := func(maxValues int) error {
		defaults := 0
		for _, o := range c.Options {
			if o.Default {
				defaults++
			}
		}
		if defaults > maxValues {
			return fmt.Errorf("modal input has %d options picked by default, but at most %d can be picked", defaults, maxValues)
		}
		return nil
	}

	// checkOptionLimits checks the value limits against the options. The
	// minimum is checked against the maximum by setValueLimits.
	checkOptionLimits := func(defaultMax int) error {
		if c.MaxValues > len(c.Options) {
			return fmt.Errorf("maximum values %d is more than the %d options", c.MaxValues, len(c.Options))
		}
		return checkDefaults(maxValuesOrDefault(c.MaxValues, defaultMax))
	}

	switch c.Type {
	case "", ModalComponentTypeTextInput:
		value, err := ctx.EvalTemplateKeepSpace(c.Value)
		if err != nil {
			return nil, err
		}

		style := c.Style
		if style == 0 {
			style = int(discord.TextInputShortStyle)
		}

		res["type"] = discord.TextInputComponentType
		res["style"] = style
		res["required"] = c.Required
		if c.MinLength > 0 {
			res["min_length"] = c.MinLength
		}
		if c.MaxLength > 0 {
			res["max_length"] = c.MaxLength
		}
		if value.String() != "" {
			res["value"] = value.String()
		}
		if err := evalPlaceholder(modalTextInputPlaceholderMaxLen); err != nil {
			return nil, err
		}
	case ModalComponentTypeStringSelect:
		options, err := buildModalOptions(ctx, c.Options, 1, modalMaxSelectOptions)
		if err != nil {
			return nil, err
		}

		res["type"] = discord.StringSelectComponentType
		res["options"] = options
		res["required"] = c.Required
		if err := setValueLimits(modalMaxSelectOptions, 1); err != nil {
			return nil, err
		}
		if err := checkOptionLimits(1); err != nil {
			return nil, err
		}
		if err := evalPlaceholder(modalSelectPlaceholderMaxLen); err != nil {
			return nil, err
		}
	case ModalComponentTypeUserSelect,
		ModalComponentTypeRoleSelect,
		ModalComponentTypeMentionableSelect,
		ModalComponentTypeChannelSelect:
		switch c.Type {
		case ModalComponentTypeUserSelect:
			res["type"] = discord.UserSelectComponentType
		case ModalComponentTypeRoleSelect:
			res["type"] = discord.RoleSelectComponentType
		case ModalComponentTypeMentionableSelect:
			res["type"] = discord.MentionableSelectComponentType
		case ModalComponentTypeChannelSelect:
			res["type"] = discord.ChannelSelectComponentType
			if len(c.ChannelTypes) > 0 {
				res["channel_types"] = c.ChannelTypes
			}
		}

		res["required"] = c.Required
		if err := setValueLimits(modalMaxSelectOptions, 1); err != nil {
			return nil, err
		}
		if err := evalPlaceholder(modalSelectPlaceholderMaxLen); err != nil {
			return nil, err
		}
	case ModalComponentTypeRadioGroup:
		options, err := buildModalOptions(ctx, c.Options, modalMinRadioOptions, modalMaxGroupOptions)
		if err != nil {
			return nil, err
		}

		res["type"] = discord.RadioGroupComponentType
		res["options"] = options
		res["required"] = c.Required
		if err := checkDefaults(1); err != nil {
			return nil, err
		}
	case ModalComponentTypeCheckboxGroup:
		options, err := buildModalOptions(ctx, c.Options, 1, modalMaxGroupOptions)
		if err != nil {
			return nil, err
		}

		res["type"] = discord.CheckboxGroupComponentType
		res["options"] = options
		res["required"] = c.Required
		if err := setValueLimits(modalMaxGroupOptions, len(c.Options)); err != nil {
			return nil, err
		}
		if err := checkOptionLimits(len(c.Options)); err != nil {
			return nil, err
		}
	case ModalComponentTypeCheckbox:
		res["type"] = discord.CheckboxComponentType
		if c.Default {
			res["default"] = true
		}
	default:
		return nil, fmt.Errorf("unknown modal input type %q", c.Type)
	}

	return res, nil
}

func maxValuesOrDefault(maxValues, defaultMax int) int {
	if maxValues == 0 {
		return defaultMax
	}
	return maxValues
}

func buildModalOptions(ctx *FlowContext, options []ModalComponentOptionData, min, max int) ([]map[string]any, error) {
	if len(options) < min || len(options) > max {
		return nil, fmt.Errorf("modal input must have between %d and %d options, got %d", min, max, len(options))
	}

	res := make([]map[string]any, len(options))
	values := make(map[string]bool, len(options))
	for i, o := range options {
		label, err := ctx.EvalTemplate(o.Label)
		if err != nil {
			return nil, err
		}
		value, err := ctx.EvalTemplate(o.Value)
		if err != nil {
			return nil, err
		}
		description, err := ctx.EvalTemplate(o.Description)
		if err != nil {
			return nil, err
		}

		if label.String() == "" {
			return nil, fmt.Errorf("option %d has no label", i+1)
		}

		// An option without a value submits its label.
		v := value.String()
		if v == "" {
			v = label.String()
		}
		if values[v] {
			return nil, fmt.Errorf("modal input has more than one option with the value %q", v)
		}
		values[v] = true

		option := map[string]any{
			"label": label.String(),
			"value": v,
		}
		if description.String() != "" {
			option["description"] = description.String()
		}
		if o.Default {
			option["default"] = true
		}

		res[i] = option
	}

	return res, nil
}
