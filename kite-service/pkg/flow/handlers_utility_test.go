package flow

import (
	"context"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// executeRandomPick runs a Random Pick block with items and env and returns
// the picked item.
func executeRandomPick(t *testing.T, items string, env eval.Env) (string, error) {
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

	node := &CompiledFlowNode{ID: "1", Type: FlowNodeTypeActionRandomPick, Data: FlowNodeData{RandomItems: items}}
	err := node.Execute(c)
	return c.GetNodeResult("1").String(), err
}

func TestRandomPickPicksItem(t *testing.T) {
	items := []string{"Germany", "France", "Japan"}
	picked := map[string]bool{}
	for range 200 {
		item, err := executeRandomPick(t, "Germany\nFrance\nJapan", eval.Env{})
		require.NoError(t, err)
		require.Contains(t, items, item)
		picked[item] = true
	}
	assert.Len(t, picked, len(items))
}

func TestRandomPickIgnoresBlankLines(t *testing.T) {
	item, err := executeRandomPick(t, "\n  \r\n  Germany is a country in Europe.  \r\n\n\t\n", eval.Env{})
	require.NoError(t, err)
	assert.Equal(t, "Germany is a country in Europe.", item)
}

func TestRandomPickErrorsWithoutItems(t *testing.T) {
	for _, items := range []string{"", "\n \n\t", "{{ empty }}\n"} {
		_, err := executeRandomPick(t, items, eval.Env{"empty": " "})
		assert.ErrorContains(t, err, "no items to pick from", "%q", items)
	}
}

func TestRandomPickEvaluatesTemplates(t *testing.T) {
	// The lines of a placeholder are separate items.
	env := eval.Env{"items": "Germany\nFrance"}
	picked := map[string]bool{}
	for range 200 {
		item, err := executeRandomPick(t, "{{ items }}\n\n{{ 'Ja' + 'pan' }}", env)
		require.NoError(t, err)
		picked[item] = true
	}
	assert.Equal(t, map[string]bool{"Germany": true, "France": true, "Japan": true}, picked)
}
