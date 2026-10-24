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
