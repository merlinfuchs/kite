package flow

import (
	"context"
	"github.com/diamondburned/arikawa/v3/utils/json"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

func autocompleteCommand(argData FlowNodeData, autocompleteNodes []FlowNode, mainNodes []FlowNode) FlowData {
	data := FlowData{
		Nodes: []FlowNode{
			{ID: "entry", Type: FlowNodeTypeEntryCommand, Data: FlowNodeData{Name: "fruit", Description: "Pick a fruit"}},
			{ID: "arg", Type: FlowNodeTypeOptionCommandArgument, Data: argData},
		},
		Edges: []FlowEdge{{ID: "e0", Source: "arg", Target: "entry"}},
	}
	for i, n := range autocompleteNodes {
		data.Nodes = append(data.Nodes, n)
		data.Edges = append(data.Edges, FlowEdge{
			ID:           "a" + string(rune('0'+i)),
			Source:       "arg",
			Target:       n.ID,
			SourceHandle: null.StringFrom(AutocompleteHandle),
		})
	}
	for i, n := range mainNodes {
		data.Nodes = append(data.Nodes, n)
		data.Edges = append(data.Edges, FlowEdge{ID: "m" + string(rune('0'+i)), Source: "entry", Target: n.ID})
	}
	return data
}

func fruitArgument(autocomplete bool) FlowNodeData {
	return FlowNodeData{
		Name:                        "fruit",
		Description:                 "The fruit",
		CommandArgumentType:         CommandArgumentTypeString,
		CommandArgumentAutocomplete: autocomplete,
		CommandArgumentChoices: []CommandArgumentChoiceData{
			{Name: "Apple", Value: "apple"},
			{Name: "Banana", Value: "banana"},
			{Name: "Pineapple", Value: "pineapple"},
		},
	}
}

// Discord rejects arguments that have both choices and autocomplete.
func TestAutocompleteArgumentSendsNoChoices(t *testing.T) {
	entry, err := CompileCommand(autocompleteCommand(fruitArgument(true), nil, nil))
	require.NoError(t, err)

	option := entry.CommandArguments()[0].(*discord.StringOption)
	assert.True(t, option.Autocomplete)
	assert.Empty(t, option.Choices)

	entry, err = CompileCommand(autocompleteCommand(fruitArgument(false), nil, nil))
	require.NoError(t, err)

	option = entry.CommandArguments()[0].(*discord.StringOption)
	assert.False(t, option.Autocomplete)
	assert.Len(t, option.Choices, 3)
}

func TestAutocompleteOnlyForArgumentsWithChoices(t *testing.T) {
	data := FlowNodeData{Name: "who", Description: "Someone", CommandArgumentType: CommandArgumentTypeUser, CommandArgumentAutocomplete: true}
	assert.Error(t, data.Validate(FlowNodeTypeOptionCommandArgument))

	entry, err := CompileCommand(autocompleteCommand(data, nil, nil))
	require.NoError(t, err)
	assert.Empty(t, entry.CommandArgument("who").StaticAutocompleteResponse(discord.AutocompleteOption{}).Data.Choices)

	data = fruitArgument(true)
	assert.NoError(t, data.Validate(FlowNodeTypeOptionCommandArgument))
}

func TestStaticAutocompleteShowsMatchingChoices(t *testing.T) {
	entry, err := CompileCommand(autocompleteCommand(fruitArgument(true), nil, nil))
	require.NoError(t, err)

	arg := entry.CommandArgument("fruit")
	require.NotNil(t, arg)
	assert.False(t, arg.HasAutocompleteFlow())

	resp := arg.StaticAutocompleteResponse(discord.AutocompleteOption{
		Type:    discord.StringOptionType,
		Name:    "fruit",
		Value:   json.Raw(`"apple"`),
		Focused: true,
	})
	assert.Equal(t, api.AutocompleteResult, resp.Type)
	assert.Equal(t, api.AutocompleteStringChoices{
		{Name: "Apple", Value: "apple"},
		{Name: "Pineapple", Value: "pineapple"},
	}, resp.Data.Choices)

	// A missing argument shows nothing instead of an error.
	var missing *CompiledFlowNode
	assert.Equal(t, api.AutocompleteStringChoices{}, missing.StaticAutocompleteResponse(discord.AutocompleteOption{}).Data.Choices)
}

func TestAutocompleteFlowChecks(t *testing.T) {
	respond := FlowNode{ID: "show", Type: FlowNodeTypeActionResponseAutocomplete}

	entry, err := CompileCommand(autocompleteCommand(fruitArgument(true), []FlowNode{respond}, nil))
	require.NoError(t, err)
	assert.True(t, entry.CommandArgument("fruit").HasAutocompleteFlow())
	assert.NoError(t, entry.CheckCommandArguments())

	// Autocomplete can only show choices.
	entry, err = CompileCommand(autocompleteCommand(fruitArgument(true), []FlowNode{{ID: "reply", Type: FlowNodeTypeActionResponseCreate}}, nil))
	require.NoError(t, err)
	assert.ErrorContains(t, entry.CheckCommandArguments(), "can only show choices")

	// The command itself has no choices to show.
	entry, err = CompileCommand(autocompleteCommand(fruitArgument(true), nil, []FlowNode{respond}))
	require.NoError(t, err)
	assert.ErrorContains(t, entry.CheckCommandArguments(), "only works after the autocomplete output")

	// Without autocomplete, the output isn't used.
	entry, err = CompileCommand(autocompleteCommand(fruitArgument(false), []FlowNode{respond}, nil))
	require.NoError(t, err)
	assert.False(t, entry.CommandArgument("fruit").HasAutocompleteFlow())
}

func TestFilterAutocompleteChoices(t *testing.T) {
	choices := []autocompleteChoice{
		{Name: "Pineapple", Value: "pineapple"},
		{Name: "Apple", Value: "apple"},
		{Name: "Cherry", Value: "cherry"},
		{Name: "Red", Value: "apple_red"},
	}

	assert.Equal(t, []autocompleteChoice{
		{Name: "Apple", Value: "apple"},
		{Name: "Red", Value: "apple_red"},
		{Name: "Pineapple", Value: "pineapple"},
	}, filterAutocompleteChoices(choices, " APP "))
	assert.Equal(t, choices, filterAutocompleteChoices(choices, ""))
	assert.Empty(t, filterAutocompleteChoices(choices, "kiwi"))
}

func TestAutocompleteResponseFitsDiscord(t *testing.T) {
	var many []autocompleteChoice
	for i := 0; i < 30; i++ {
		many = append(many, autocompleteChoice{Name: "n", Value: "1"})
	}
	assert.Len(t, autocompleteResponse(many, discord.StringOptionType).Data.Choices, 25)

	long := string(make([]rune, 101))
	resp := autocompleteResponse([]autocompleteChoice{
		{Name: "One", Value: "1"},
		{Name: "Half", Value: "1.5"},
		{Name: "", Value: "2"},
		{Name: "Long value", Value: long},
	}, discord.IntegerOptionType)
	assert.Equal(t, api.AutocompleteIntegerChoices{{Name: "One", Value: 1}}, resp.Data.Choices)

	resp = autocompleteResponse([]autocompleteChoice{{Name: "Half", Value: "1.5"}}, discord.NumberOptionType)
	assert.Equal(t, api.AutocompleteNumberChoices{{Name: "Half", Value: 1.5}}, resp.Data.Choices)

	resp = autocompleteResponse([]autocompleteChoice{{Name: "Long value", Value: "ok"}, {Name: "x", Value: "123456789 123456789 123456789 123456789 123456789 123456789 123456789 123456789 123456789 123456789 1"}}, discord.StringOptionType)
	assert.Len(t, resp.Data.Choices, 1)

	name := truncateChoice(string(make([]rune, 150)))
	assert.Equal(t, 100, len([]rune(name)))
}

func TestAutocompleteChoicesFromList(t *testing.T) {
	assert.Equal(t, []autocompleteChoice{
		{Name: "Apple", Value: "Apple"},
		{Name: "Banana", Value: "banana"},
		{Name: "Cherry", Value: "cherry"},
		{Name: "3", Value: "3"},
	}, autocompleteChoicesFromList(thing.NewAny([]any{
		"Apple",
		map[string]any{"name": "Banana", "value": "banana"},
		map[string]any{"label": "Cherry", "value": "cherry"},
		3,
	})))

	assert.Equal(t, []autocompleteChoice{
		{Name: "Red", Value: "Red"},
		{Name: "Green", Value: "Green"},
	}, autocompleteChoicesFromList(thing.NewString("Red\nGreen\n")))

	assert.Equal(t, []autocompleteChoice{
		{Name: "a", Value: "1"},
	}, autocompleteChoicesFromList(thing.NewString(`[{"name": "a", "value": 1}]`)))

	assert.Empty(t, autocompleteChoicesFromList(thing.Null))
}

func runAutocomplete(t *testing.T, node *CompiledFlowNode, options discord.AutocompleteOptions) (*TestDiscordProvider, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	interaction := &discord.InteractionEvent{
		ID:   1,
		Data: &discord.AutocompleteInteraction{Name: "fruit", Options: options},
	}
	discordProvider := &TestDiscordProvider{}
	c := NewContext(
		ctx,
		5*time.Second,
		&TestContextData{interaction: interaction},
		FlowProviders{Discord: discordProvider},
		FlowContextLimits{MaxStackDepth: 10, MaxOperations: 1000, MaxCredits: 1000},
		eval.NewContext(eval.Env{
			"autocomplete": map[string]any{"value": eval.FocusedOption(options).String()},
			"list":         []string{"Kiwi", "Mango"},
		}),
		nil,
	)
	t.Cleanup(c.Cancel)

	return discordProvider, node.Execute(c)
}

func TestShowAutocompleteChoices(t *testing.T) {
	node := &CompiledFlowNode{
		ID:   "show",
		Type: FlowNodeTypeActionResponseAutocomplete,
		Data: FlowNodeData{
			AutocompleteChoices: []CommandArgumentChoiceData{
				{Name: "You typed {{autocomplete.value}}", Value: "{{autocomplete.value}}"},
				{Name: "Melon"},
			},
			AutocompleteChoicesList: "{{list}}",
		},
	}
	focused := discord.AutocompleteOptions{{Type: discord.StringOptionType, Name: "fruit", Value: json.Raw(`"m"`), Focused: true}}

	discordProvider, err := runAutocomplete(t, node, focused)
	require.NoError(t, err)
	assert.Equal(t, api.AutocompleteStringChoices{
		{Name: "You typed m", Value: "m"},
		{Name: "Melon", Value: "Melon"},
		{Name: "Kiwi", Value: "Kiwi"},
		{Name: "Mango", Value: "Mango"},
	}, discordProvider.response.Data.Choices)

	node.Data.AutocompleteFilter = true
	discordProvider, err = runAutocomplete(t, node, focused)
	require.NoError(t, err)
	assert.Equal(t, api.AutocompleteStringChoices{
		{Name: "You typed m", Value: "m"},
		{Name: "Melon", Value: "Melon"},
		{Name: "Mango", Value: "Mango"},
	}, discordProvider.response.Data.Choices)

	// Only the matching ones.
	focused[0].Value = json.Raw(`"an"`)
	discordProvider, err = runAutocomplete(t, node, focused)
	require.NoError(t, err)
	assert.Equal(t, api.AutocompleteStringChoices{
		{Name: "You typed an", Value: "an"},
		{Name: "Mango", Value: "Mango"},
	}, discordProvider.response.Data.Choices)
}

// The blocks after an autocomplete output that show nothing still have to
// answer, or Discord shows an error.
func TestAutocompleteFlowWithoutChoicesShowsNone(t *testing.T) {
	entry, err := CompileCommand(autocompleteCommand(fruitArgument(true), []FlowNode{
		{ID: "log", Type: FlowNodeTypeActionLog, Data: FlowNodeData{LogLevel: "info", LogMessage: "typing"}},
	}, nil))
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	discordProvider := &TestDiscordProvider{}
	c := NewContext(
		ctx,
		5*time.Second,
		&TestContextData{interaction: &discord.InteractionEvent{Data: &discord.AutocompleteInteraction{}}},
		FlowProviders{Discord: discordProvider, Log: &provider.MockLogProvider{}},
		FlowContextLimits{MaxStackDepth: 10, MaxOperations: 1000, MaxCredits: 1000},
		eval.NewContext(eval.Env{}),
		nil,
	)
	defer c.Cancel()

	require.NoError(t, entry.CommandArgument("fruit").ExecuteAutocomplete(c))
	assert.Equal(t, api.AutocompleteResult, discordProvider.response.Type)
	assert.Equal(t, api.AutocompleteStringChoices{}, discordProvider.response.Data.Choices)
}

func TestShowAutocompleteChoicesNeedsAutocomplete(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c := NewContext(
		ctx,
		5*time.Second,
		&TestContextData{interaction: &discord.InteractionEvent{Data: &discord.CommandInteraction{}}},
		FlowProviders{Discord: &TestDiscordProvider{}},
		FlowContextLimits{MaxStackDepth: 10, MaxOperations: 1000, MaxCredits: 1000},
		eval.NewContext(eval.Env{}),
		nil,
	)
	defer c.Cancel()

	node := &CompiledFlowNode{ID: "show", Type: FlowNodeTypeActionResponseAutocomplete}
	assert.ErrorContains(t, node.Execute(c), "only works after the autocomplete output")
}
