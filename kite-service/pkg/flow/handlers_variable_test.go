package flow

import (
	"context"
	"testing"
	"time"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/provider"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/guregu/null.v4"
)

type variableTestProvider struct {
	provider.MockVariableProvider
	scoped bool
	scopes []null.String
}

func (p *variableTestProvider) VariableScoped(ctx context.Context, id string) (bool, error) {
	return p.scoped, nil
}

func (p *variableTestProvider) UpdateVariable(ctx context.Context, id string, scope null.String, operation provider.VariableOperation, value thing.Thing) (thing.Thing, error) {
	p.scopes = append(p.scopes, scope)
	return value, nil
}

func (p *variableTestProvider) Variable(ctx context.Context, id string, scope null.String) (thing.Thing, error) {
	p.scopes = append(p.scopes, scope)
	return thing.Null, nil
}

func (p *variableTestProvider) DeleteVariable(ctx context.Context, id string, scope null.String) error {
	p.scopes = append(p.scopes, scope)
	return nil
}

// executeVariableBlocks runs a Set, Get and Delete Variable block with scope
// and returns the scopes they passed on.
func executeVariableBlocks(t *testing.T, scoped bool, scope string) ([]null.String, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	p := &variableTestProvider{scoped: scoped}
	c := NewContext(
		ctx,
		5*time.Second,
		&TestContextData{},
		FlowProviders{Variable: p},
		FlowContextLimits{MaxStackDepth: 10, MaxOperations: 1000, MaxCredits: 1000},
		eval.NewContext(eval.Env{}),
		nil,
	)
	t.Cleanup(c.Cancel)

	for _, nodeType := range []FlowNodeType{
		FlowNodeTypeActionVariableSet,
		FlowNodeTypeActionVariableGet,
		FlowNodeTypeActionVariableDelete,
	} {
		node := &CompiledFlowNode{ID: "1", Type: nodeType, Data: FlowNodeData{
			VariableID:    "1",
			VariableScope: scope,
			VariableValue: "1",
		}}
		if err := node.Execute(c); err != nil {
			return p.scopes, err
		}
	}
	return p.scopes, nil
}

func TestVariableBlocksUseScope(t *testing.T) {
	scopes, err := executeVariableBlocks(t, true, "{{'user'}}")
	require.NoError(t, err)
	user := null.StringFrom("user")
	assert.Equal(t, []null.String{user, user, user}, scopes)

	// Only an empty scope is missing, 0 is a scope like any other.
	scopes, err = executeVariableBlocks(t, true, "0")
	require.NoError(t, err)
	assert.Equal(t, null.StringFrom("0"), scopes[0])
}

// Blocks keep their scope when their variable is switched to unscoped. It
// must neither pick a scoped value nor fail the block.
func TestVariableBlocksIgnoreScopeOfUnscopedVariable(t *testing.T) {
	for _, scope := range []string{"", "user", "{{ a.b.c }}"} {
		scopes, err := executeVariableBlocks(t, false, scope)
		require.NoError(t, err, scope)
		assert.Equal(t, []null.String{{}, {}, {}}, scopes, scope)
	}
}

// Without a scope, a scoped variable would read and write a value no scope
// can see.
func TestVariableBlocksRequireScopeOfScopedVariable(t *testing.T) {
	for _, scope := range []string{"", "{{''}}"} {
		scopes, err := executeVariableBlocks(t, true, scope)
		assert.ErrorContains(t, err, "the scope is empty", scope)
		assert.Empty(t, scopes, scope)
	}

	_, err := executeVariableBlocks(t, true, "{{ a.b.c }}")
	assert.Error(t, err)
}

type variableListTestProvider struct {
	provider.MockVariableProvider
	scoped bool
	opts   provider.VariableListOptions
}

func (p *variableListTestProvider) VariableScoped(ctx context.Context, id string) (bool, error) {
	return p.scoped, nil
}

func (p *variableListTestProvider) VariableValues(ctx context.Context, id string, opts provider.VariableListOptions) ([]provider.VariableValue, error) {
	p.opts = opts
	return []provider.VariableValue{
		{Scope: null.StringFrom("guild:alex"), Value: thing.NewInt(1240), UpdatedAt: time.Unix(100, 0)},
		{Scope: null.StringFrom("guild:sam"), Value: thing.NewString("980"), UpdatedAt: time.Unix(200, 0)},
	}, nil
}

func executeVariableList(t *testing.T, scoped bool, data FlowNodeData) (*variableListTestProvider, thing.Thing, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	p := &variableListTestProvider{scoped: scoped}
	c := NewContext(
		ctx,
		5*time.Second,
		&TestContextData{},
		FlowProviders{Variable: p},
		FlowContextLimits{MaxStackDepth: 10, MaxOperations: 1000, MaxCredits: 1000},
		eval.NewContext(eval.Env{}),
		nil,
	)
	t.Cleanup(c.Cancel)

	data.VariableID = "1"
	node := &CompiledFlowNode{ID: "list", Type: FlowNodeTypeActionVariableList, Data: data}
	err := node.Execute(c)
	return p, c.GetNodeResult("list"), err
}

func TestVariableListResult(t *testing.T) {
	p, res, err := executeVariableList(t, true, FlowNodeData{
		VariableScopePrefix: "{{'guild'}}:",
		VariableLimit:       "{{2}}",
	})
	require.NoError(t, err)
	assert.Equal(t, provider.VariableListOptions{
		ScopePrefix: "guild:",
		Sort:        provider.VariableListSortValueDesc,
		Limit:       2,
	}, p.opts)

	items := res.Array()
	require.Len(t, items, 2)
	first := items[0].Value.(map[string]thing.Thing)
	assert.Equal(t, "guild:alex", first["scope"].String())
	assert.Equal(t, int64(1240), first["value"].Int())
	assert.Equal(t, int64(100), first["updated_at"].Int())
}

func TestVariableListDefaultsAndUnscoped(t *testing.T) {
	// Unscoped variables have one value, which a prefix can't match.
	p, _, err := executeVariableList(t, false, FlowNodeData{
		VariableScopePrefix: "guild:",
		VariableSort:        provider.VariableListSortUpdatedDesc,
	})
	require.NoError(t, err)
	assert.Equal(t, provider.VariableListOptions{
		Sort:  provider.VariableListSortUpdatedDesc,
		Limit: defaultVariableListLimit,
	}, p.opts)
}

func TestVariableListRejectsBadLimit(t *testing.T) {
	for _, limit := range []string{"0", "101", "ten", "{{''}} "} {
		_, _, err := executeVariableList(t, true, FlowNodeData{VariableLimit: limit})
		assert.ErrorContains(t, err, "from 1 to 100", limit)
	}
}
