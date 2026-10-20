package message

import "github.com/diamondburned/arikawa/v3/discord"

// FindComponent returns the first component, at any depth, that match accepts.
func (m *MessageData) FindComponent(match func(c *ComponentData) bool) *ComponentData {
	var res *ComponentData
	m.EachComponent(func(c *ComponentData) error {
		if res == nil && match(c) {
			res = c
		}
		return nil
	})
	return res
}

// PickedOption returns the first option of a string select that one of the
// picked values belongs to and that accept allows, or nil. A value is matched
// against the option values, and otherwise by its position in the menu that was
// sent, because a value can hold a placeholder that was filled in when sending.
func (c *ComponentData) PickedOption(
	values []string,
	sent *discord.StringSelectComponent,
	accept func(option *ComponentSelectOptionData) bool,
) *ComponentSelectOptionData {
	for _, value := range values {
		option := c.optionForValue(value, sent)
		if option != nil && accept(option) {
			return option
		}
	}
	return nil
}

func (c *ComponentData) optionForValue(value string, sent *discord.StringSelectComponent) *ComponentSelectOptionData {
	for i := range c.Options {
		if c.Options[i].Value == value {
			return &c.Options[i]
		}
	}

	if sent == nil || len(sent.Options) != len(c.Options) {
		return nil
	}

	for i, option := range sent.Options {
		if option.Value == value {
			return &c.Options[i]
		}
	}
	return nil
}

// FindSentStringSelect returns the string select with the custom ID in the
// components of a message that was sent, or nil.
func FindSentStringSelect(msg *discord.Message, customID discord.ComponentID) *discord.StringSelectComponent {
	if msg == nil {
		return nil
	}

	for _, component := range msg.Components {
		if res := findSentStringSelect(component, customID); res != nil {
			return res
		}
	}
	return nil
}

func findSentStringSelect(component discord.Component, customID discord.ComponentID) *discord.StringSelectComponent {
	switch c := component.(type) {
	case *discord.StringSelectComponent:
		if c.CustomID == customID {
			return c
		}
	case *discord.ActionRowComponent:
		for _, child := range *c {
			if res := findSentStringSelect(child, customID); res != nil {
				return res
			}
		}
	case *discord.SectionComponent:
		if c.Accessory != nil {
			return findSentStringSelect(c.Accessory, customID)
		}
	case *discord.ContainerComponent:
		for _, child := range c.Components {
			if res := findSentStringSelect(child, customID); res != nil {
				return res
			}
		}
	}
	return nil
}
