package flow

import (
	"context"
	"encoding/json"
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// executeRandom runs a Generate Random Value block with data and env, after
// storing results, and returns its result.
func executeRandom(t *testing.T, data FlowNodeData, env eval.Env, results map[string]thing.Thing) (thing.Thing, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	c := NewContext(
		ctx,
		5*time.Second,
		&TestContextData{},
		FlowProviders{},
		FlowContextLimits{MaxStackDepth: 10, MaxOperations: 1000, MaxCredits: 1000},
		eval.NewContext(env),
		nil,
	)
	t.Cleanup(c.Cancel)

	for id, result := range results {
		c.StoreNodeResult(&CompiledFlowNode{ID: id}, result)
	}

	node := &CompiledFlowNode{ID: "1", Type: FlowNodeTypeActionRandomGenerate, Data: data}
	err := node.Execute(c)
	return c.GetNodeResult("1"), err
}

// pickAll runs the item mode until every item was likely picked, and returns
// the picked items.
func pickAll(t *testing.T, items string, env eval.Env, results map[string]thing.Thing) map[string]thing.Thing {
	picked := map[string]thing.Thing{}
	for range 200 {
		item, err := executeRandom(t, FlowNodeData{RandomMode: RandomModeItem, RandomItems: items}, env, results)
		require.NoError(t, err)
		picked[item.String()] = item
	}
	return picked
}

func TestRandomNumberIsDefault(t *testing.T) {
	for _, mode := range []RandomMode{"", RandomModeNumber} {
		n, err := executeRandom(t, FlowNodeData{RandomMode: mode, RandomMin: "3", RandomMax: "5"}, eval.Env{}, nil)
		require.NoError(t, err)
		assert.Equal(t, thing.TypeInt, n.Type)
		assert.Contains(t, []int64{3, 4}, n.Int())
	}
}

func TestRandomUnknownMode(t *testing.T) {
	_, err := executeRandom(t, FlowNodeData{RandomMode: "dice", RandomItems: "a"}, eval.Env{}, nil)
	assert.ErrorContains(t, err, `unknown random mode "dice"`)
}

func TestRandomItemPicksLine(t *testing.T) {
	picked := pickAll(t, "Germany\nFrance\nJapan", eval.Env{}, nil)
	assert.ElementsMatch(t, []string{"Germany", "France", "Japan"}, slices.Collect(maps.Keys(picked)))
	assert.Equal(t, thing.TypeString, picked["Germany"].Type)
}

func TestRandomItemIgnoresBlankLines(t *testing.T) {
	item, err := executeRandom(t, FlowNodeData{
		RandomMode:  RandomModeItem,
		RandomItems: "\n  \r\n  Germany is a country in Europe.  \r\n\n\t\n",
	}, eval.Env{}, nil)
	require.NoError(t, err)
	assert.Equal(t, "Germany is a country in Europe.", item.String())
}

func TestRandomItemErrorsWithoutItems(t *testing.T) {
	for _, items := range []string{"", "\n \n\t", "{{ empty }}\n", "{{ [] }}"} {
		_, err := executeRandom(t, FlowNodeData{RandomMode: RandomModeItem, RandomItems: items}, eval.Env{"empty": " "}, nil)
		assert.ErrorContains(t, err, "no items to pick from", "%q", items)
	}
}

// The lines of a placeholder are separate items.
func TestRandomItemEvaluatesTemplates(t *testing.T) {
	picked := pickAll(t, "{{ items }}\n\n{{ 'Ja' + 'pan' }}", eval.Env{"items": "Germany\nFrance"}, nil)
	assert.ElementsMatch(t, []string{"Germany", "France", "Japan"}, slices.Collect(maps.Keys(picked)))
}

func TestRandomItemPicksFromList(t *testing.T) {
	for _, items := range []string{"{{ ['a', 'b'] }}", "{{ prizes }}"} {
		picked := pickAll(t, items, eval.Env{"prizes": []string{"a", "b"}}, nil)
		assert.ElementsMatch(t, []string{"a", "b"}, slices.Collect(maps.Keys(picked)), items)
	}
}

// A picked member is still a member, so {{result('id').mention}} works.
func TestRandomItemKeepsTypeOfListItems(t *testing.T) {
	members := thing.NewArray([]thing.Thing{
		thing.NewDiscordMember(discord.Member{User: discord.User{ID: 1}}),
		thing.NewDiscordMember(discord.Member{User: discord.User{ID: 2}}),
	})
	picked := pickAll(t, "{{ result('members') }}", eval.Env{}, map[string]thing.Thing{"members": members})
	require.Len(t, picked, 2)
	for _, item := range picked {
		assert.Equal(t, thing.TypeDiscordMember, item.Type)
	}
}

// Stored variables come back from JSON, as an array or, for values stored
// before arrays had their own type, as any.
func TestListOrLinesStoredLists(t *testing.T) {
	for _, stored := range []string{
		`{"t":"array","v":[{"t":"string","v":"a"},{"t":"int","v":5}]}`,
		`{"t":"any","v":["a",5]}`,
	} {
		var value thing.Thing
		require.NoError(t, json.Unmarshal([]byte(stored), &value))

		var items []string
		for _, item := range listOrLines(value) {
			items = append(items, item.String())
		}
		assert.Equal(t, []string{"a", "5"}, items, stored)
	}
}

func TestListOrLinesSplitsText(t *testing.T) {
	var items []string
	for _, item := range listOrLines(thing.NewString(" a \n\n b\r\n")) {
		items = append(items, item.String())
	}
	assert.Equal(t, []string{"a", "b"}, items)
}

// executeValueLookup runs a lookup block with data and returns its result.
func executeValueLookup(t *testing.T, env eval.Env, data FlowNodeData) thing.Thing {
	t.Helper()

	c := NewContext(
		context.Background(),
		5*time.Second,
		&TestContextData{},
		FlowProviders{Log: &provider.MockLogProvider{}},
		FlowContextLimits{MaxStackDepth: 10, MaxOperations: 1000, MaxCredits: 1000},
		eval.NewContext(env),
		nil,
	)
	t.Cleanup(c.Cancel)

	node := &CompiledFlowNode{ID: "lookup", Type: FlowNodeTypeActionValueLookup, Data: data}
	require.NoError(t, node.Execute(c))
	return c.GetNodeResult("lookup")
}

var lookupTestRows = []LookupRow{
	{Key: "aura", Value: "aura reply"},
	{Key: "gtg", Value: "bye"},
	{Key: "Hello", Value: "hi there"},
}

func TestValueLookupEqual(t *testing.T) {
	data := FlowNodeData{LookupValue: "gtg", LookupRows: lookupTestRows}
	assert.Equal(t, "bye", executeValueLookup(t, eval.Env{}, data).String())

	// Equal doesn't match a part of the value.
	data.LookupValue = "gtg now"
	assert.Equal(t, "", executeValueLookup(t, eval.Env{}, data).String())
}

func TestValueLookupContains(t *testing.T) {
	data := FlowNodeData{LookupValue: "ok gtg now", LookupMode: ComparsionModeContains, LookupRows: lookupTestRows}
	assert.Equal(t, "bye", executeValueLookup(t, eval.Env{}, data).String())
}

func TestValueLookupStartsWith(t *testing.T) {
	data := FlowNodeData{LookupValue: "gtg now", LookupMode: ComparsionModeStartsWith, LookupRows: lookupTestRows}
	assert.Equal(t, "bye", executeValueLookup(t, eval.Env{}, data).String())

	data.LookupValue = "ok gtg"
	assert.Equal(t, "", executeValueLookup(t, eval.Env{}, data).String())
}

func TestValueLookupIgnoreCase(t *testing.T) {
	data := FlowNodeData{LookupValue: "hello", LookupRows: lookupTestRows}
	assert.Equal(t, "", executeValueLookup(t, eval.Env{}, data).String())

	data.LookupIgnoreCase = true
	assert.Equal(t, "hi there", executeValueLookup(t, eval.Env{}, data).String())

	data.LookupValue = "AURA"
	assert.Equal(t, "aura reply", executeValueLookup(t, eval.Env{}, data).String())
}

func TestValueLookupFirstMatchWins(t *testing.T) {
	data := FlowNodeData{
		LookupValue: "aura gtg",
		LookupMode:  ComparsionModeContains,
		LookupRows: []LookupRow{
			{Key: "gtg", Value: "first"},
			{Key: "aura", Value: "second"},
			{Key: "gtg", Value: "third"},
		},
	}
	assert.Equal(t, "first", executeValueLookup(t, eval.Env{}, data).String())
}

func TestValueLookupDefault(t *testing.T) {
	data := FlowNodeData{
		LookupValue:   "unknown",
		LookupRows:    lookupTestRows,
		LookupDefault: "no idea about {{word}}",
	}
	env := eval.Env{"word": "unknown"}
	assert.Equal(t, "no idea about unknown", executeValueLookup(t, env, data).String())
}

func TestValueLookupNoMatch(t *testing.T) {
	data := FlowNodeData{LookupValue: "unknown", LookupRows: lookupTestRows}
	res := executeValueLookup(t, eval.Env{}, data)
	assert.Equal(t, thing.TypeString, res.Type)
	assert.Equal(t, "", res.String())
}

func TestValueLookupTemplatedKeys(t *testing.T) {
	env := eval.Env{"owner": "123", "content": "123"}
	data := FlowNodeData{
		LookupValue: "{{content}}",
		LookupRows: []LookupRow{
			{Key: "456", Value: "someone else"},
			{Key: "{{owner}}", Value: "the owner is {{owner}}"},
		},
	}
	assert.Equal(t, "the owner is 123", executeValueLookup(t, env, data).String())
}

// Keys that are empty once placeholders are filled in don't match everything.
func TestValueLookupSkipsEmptyKeys(t *testing.T) {
	data := FlowNodeData{
		LookupValue: "anything",
		LookupMode:  ComparsionModeContains,
		LookupRows: []LookupRow{
			{Key: "{{missing}}", Value: "empty key"},
			{Key: "thing", Value: "thing"},
		},
	}
	assert.Equal(t, "thing", executeValueLookup(t, eval.Env{}, data).String())
}

// Only the value of the row that matches is evaluated, so a broken value of
// another row doesn't fail the block.
func TestValueLookupOnlyEvaluatesMatchedValue(t *testing.T) {
	data := FlowNodeData{
		LookupValue: "b",
		LookupRows: []LookupRow{
			{Key: "a", Value: "{{ nope( }}"},
			{Key: "b", Value: "fine"},
			{Key: "c", Value: "{{ nope( }}"},
		},
	}
	assert.Equal(t, "fine", executeValueLookup(t, eval.Env{}, data).String())
}

func TestValueLookupValidation(t *testing.T) {
	rows := make([]LookupRow, maxLookupRows+1)
	for i := range rows {
		rows[i] = LookupRow{Key: "k", Value: "v"}
	}
	assert.Error(t, FlowNodeData{LookupRows: rows}.Validate(FlowNodeTypeActionValueLookup))
	assert.NoError(t, FlowNodeData{LookupRows: rows[:maxLookupRows]}.Validate(FlowNodeTypeActionValueLookup))

	assert.Error(t, FlowNodeData{LookupRows: []LookupRow{{Key: "", Value: "v"}}}.Validate(FlowNodeTypeActionValueLookup))
	assert.Error(t, FlowNodeData{LookupMode: ComparsionModeEndsWith}.Validate(FlowNodeTypeActionValueLookup))
	assert.NoError(t, FlowNodeData{LookupMode: ComparsionModeContains}.Validate(FlowNodeTypeActionValueLookup))
}
