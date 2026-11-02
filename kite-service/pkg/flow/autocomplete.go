package flow

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

// AutocompleteHandle is the output of an argument block whose blocks run while
// the user types in the argument, to show the choices for what was typed.
const AutocompleteHandle = "autocomplete"

// maxAutocompleteChoices is the most choices an argument with autocomplete or
// a Show Autocomplete Choices block can have. Discord only shows 25 at once,
// the ones matching what was typed.
const maxAutocompleteChoices = 500

const (
	// discordMaxAutocompleteChoices is the most choices Discord shows.
	discordMaxAutocompleteChoices = 25
	// discordMaxChoiceLength is the longest name and text value a choice can have.
	discordMaxChoiceLength = 100
)

type autocompleteChoice struct {
	Name  string
	Value string
}

// argumentTypeHasChoices reports whether arguments of the type can have
// choices, and so autocomplete.
func argumentTypeHasChoices(t CommandArgumentType) bool {
	switch t {
	case CommandArgumentTypeString, CommandArgumentTypeInteger, CommandArgumentTypeNumber:
		return true
	}
	return false
}

// usesAutocomplete reports whether the argument block asks for choices while
// the user types instead of sending its choices to Discord.
func (n *CompiledFlowNode) usesAutocomplete() bool {
	return n.IsCommandArgument() &&
		n.Data.CommandArgumentAutocomplete &&
		argumentTypeHasChoices(n.Data.CommandArgumentType)
}

// CommandArgument returns the argument block of the command named name, if
// it has one.
func (n *CompiledFlowNode) CommandArgument(name string) *CompiledFlowNode {
	for _, node := range n.commandArgumentNodes() {
		if node.Data.Name == name {
			return node
		}
	}
	return nil
}

// HasAutocompleteFlow reports whether blocks are connected to the autocomplete
// output of the argument block. Without any, its own choices are shown.
func (n *CompiledFlowNode) HasAutocompleteFlow() bool {
	return n.usesAutocomplete() && len(n.Children.Handles[AutocompleteHandle]) > 0
}

// checkAutocomplete reports blocks that can't run where they are: blocks that
// respond to the interaction in an autocomplete output, which can only show
// choices, and Show Autocomplete Choices blocks outside of one.
func (n *CompiledFlowNode) checkAutocomplete() error {
	for _, arg := range n.commandArgumentNodes() {
		children := arg.Children.Handles[AutocompleteHandle]
		if node := firstInExecution(children, func(c *CompiledFlowNode) bool {
			return isInteractionOnly(c.Type) && c.Type != FlowNodeTypeActionResponseAutocomplete
		}); node != nil {
			return fmt.Errorf("block %s can't be used after the autocomplete output of argument %s, it can only show choices", node.Type, arg.Data.Name)
		}
	}

	isAutocompleteResponse := func(c *CompiledFlowNode) bool {
		return c.Type == FlowNodeTypeActionResponseAutocomplete
	}
	if isAutocompleteResponse(n) || n.FirstChildMatching(isAutocompleteResponse) != nil {
		return fmt.Errorf("block %s only works after the autocomplete output of a command argument", FlowNodeTypeActionResponseAutocomplete)
	}
	return nil
}

// StaticAutocompleteResponse shows the choices of the argument block that
// match what the user typed. It shows none for a nil block, e.g. an argument
// the command doesn't have anymore, so Discord doesn't show an error.
func (n *CompiledFlowNode) StaticAutocompleteResponse(focused discord.AutocompleteOption) api.InteractionResponse {
	var choices []autocompleteChoice
	if n != nil && n.usesAutocomplete() {
		for _, c := range n.Data.CommandArgumentChoices {
			choices = append(choices, autocompleteChoice{Name: c.Name, Value: c.Value})
		}
	}

	return autocompleteResponse(filterAutocompleteChoices(choices, focused.String()), focused.Type)
}

// ExecuteAutocomplete runs the blocks connected to the autocomplete output of
// the argument block. If none of them shows choices, it shows none, which
// Discord would show as an error otherwise.
func (n *CompiledFlowNode) ExecuteAutocomplete(ctx *FlowContext) error {
	interaction := ctx.Data.Interaction()
	if interaction == nil {
		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "interaction is nil",
		}
	}

	runErr := n.ExecuteChildrenByHandle(ctx, AutocompleteHandle)

	responded, err := ctx.Discord.HasCreatedInteractionResponse(ctx, interaction.ID)
	if err == nil && !responded {
		var optionType discord.CommandOptionType
		if data, ok := interaction.Data.(*discord.AutocompleteInteraction); ok {
			optionType = eval.FocusedOption(data.Options).Type
		}
		_, err = ctx.Discord.CreateInteractionResponse(ctx, interaction.ID, interaction.Token, autocompleteResponse(nil, optionType))
	}

	if runErr != nil {
		return runErr
	}
	return err
}

func executeActionResponseAutocomplete(n *CompiledFlowNode, ctx *FlowContext) error {
	interaction := ctx.Data.Interaction()
	var data *discord.AutocompleteInteraction
	if interaction != nil {
		data, _ = interaction.Data.(*discord.AutocompleteInteraction)
	}
	if data == nil {
		return &FlowError{
			Code:    FlowNodeErrorUnknown,
			Message: "Show Autocomplete Choices only works after the autocomplete output of a command argument",
		}
	}
	focused := eval.FocusedOption(data.Options)

	var choices []autocompleteChoice
	for _, c := range n.Data.AutocompleteChoices {
		name, err := ctx.EvalTemplate(c.Name)
		if err != nil {
			return err
		}
		value, err := ctx.EvalTemplate(c.Value)
		if err != nil {
			return err
		}
		choices = append(choices, newAutocompleteChoice(name.String(), value.String()))
	}

	if n.Data.AutocompleteChoicesList != "" {
		list, err := ctx.EvalTemplate(n.Data.AutocompleteChoicesList)
		if err != nil {
			return err
		}
		choices = append(choices, autocompleteChoicesFromList(list)...)
	}

	if len(choices) > maxAutocompleteChoices {
		choices = choices[:maxAutocompleteChoices]
	}
	if n.Data.AutocompleteFilter {
		choices = filterAutocompleteChoices(choices, focused.String())
	}

	resp := autocompleteResponse(choices, focused.Type)
	if _, err := ctx.Discord.CreateInteractionResponse(ctx, interaction.ID, interaction.Token, resp); err != nil {
		return err
	}

	return n.ExecuteChildren(ctx)
}

// newAutocompleteChoice is a choice whose value is its name, if it has none.
func newAutocompleteChoice(name, value string) autocompleteChoice {
	name = strings.TrimSpace(name)
	value = strings.TrimSpace(value)
	if value == "" {
		value = name
	}
	return autocompleteChoice{Name: name, Value: value}
}

// autocompleteChoicesFromList turns a list into choices: texts and numbers are
// the name and value of a choice, and objects can set both with name and
// value. A text that isn't a JSON list has a choice on each line.
func autocompleteChoicesFromList(list thing.Thing) []autocompleteChoice {
	items, ok := list.JSONValue().([]any)
	if !ok {
		text := strings.TrimSpace(list.String())
		if text == "" {
			return nil
		}
		if err := json.Unmarshal([]byte(text), &items); err != nil {
			items = nil
			for _, line := range strings.Split(text, "\n") {
				items = append(items, line)
			}
		}
	}

	var res []autocompleteChoice
	for _, item := range items {
		if len(res) == maxAutocompleteChoices {
			break
		}

		switch item := item.(type) {
		case map[string]any:
			name := choiceText(item["name"])
			if name == "" {
				name = choiceText(item["label"])
			}
			res = append(res, newAutocompleteChoice(name, choiceText(item["value"])))
		default:
			text := choiceText(item)
			res = append(res, newAutocompleteChoice(text, text))
		}
	}
	return res
}

func choiceText(v any) string {
	switch v := v.(type) {
	case nil:
		return ""
	case string:
		return v
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return thing.NewGuessTypeWithFallback(v).String()
	}
}

// filterAutocompleteChoices returns the choices that contain what the user
// typed, ignoring case, those that start with it first.
func filterAutocompleteChoices(choices []autocompleteChoice, typed string) []autocompleteChoice {
	typed = strings.ToLower(strings.TrimSpace(typed))
	if typed == "" {
		return choices
	}

	var starting, containing []autocompleteChoice
	for _, c := range choices {
		name := strings.ToLower(c.Name)
		value := strings.ToLower(c.Value)
		switch {
		case strings.HasPrefix(name, typed) || strings.HasPrefix(value, typed):
			starting = append(starting, c)
		case strings.Contains(name, typed) || strings.Contains(value, typed):
			containing = append(containing, c)
		}
	}
	return append(starting, containing...)
}

// autocompleteResponse shows the first 25 choices that Discord accepts for an
// argument of the type: ones with a name, and a value of the argument's type.
func autocompleteResponse(choices []autocompleteChoice, optionType discord.CommandOptionType) api.InteractionResponse {
	var res api.AutocompleteChoices
	switch optionType {
	case discord.IntegerOptionType:
		c := api.AutocompleteIntegerChoices{}
		for _, choice := range choices {
			value, err := strconv.ParseInt(choice.Value, 10, 64)
			if choice.Name == "" || err != nil {
				continue
			}
			if c = append(c, discord.IntegerChoice{Name: truncateChoice(choice.Name), Value: int(value)}); len(c) == discordMaxAutocompleteChoices {
				break
			}
		}
		res = c
	case discord.NumberOptionType:
		c := api.AutocompleteNumberChoices{}
		for _, choice := range choices {
			value, err := strconv.ParseFloat(choice.Value, 64)
			if choice.Name == "" || err != nil {
				continue
			}
			if c = append(c, discord.NumberChoice{Name: truncateChoice(choice.Name), Value: value}); len(c) == discordMaxAutocompleteChoices {
				break
			}
		}
		res = c
	default:
		c := api.AutocompleteStringChoices{}
		for _, choice := range choices {
			// A cut value would be a different one.
			if choice.Name == "" || choice.Value == "" || utf8.RuneCountInString(choice.Value) > discordMaxChoiceLength {
				continue
			}
			if c = append(c, discord.StringChoice{Name: truncateChoice(choice.Name), Value: choice.Value}); len(c) == discordMaxAutocompleteChoices {
				break
			}
		}
		res = c
	}

	return api.InteractionResponse{
		Type: api.AutocompleteResult,
		Data: &api.InteractionResponseData{Choices: res},
	}
}

func truncateChoice(name string) string {
	if utf8.RuneCountInString(name) <= discordMaxChoiceLength {
		return name
	}
	runes := []rune(name)
	return string(runes[:discordMaxChoiceLength-1]) + "…"
}
