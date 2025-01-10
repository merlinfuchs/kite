package flow

import (
	"fmt"

	"github.com/expr-lang/expr/ast"
	"github.com/kitecloud/kite/kite-service/pkg/eval"
	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

type nodeEvalEnv struct {
	state *FlowContextState
}

func (e *nodeEvalEnv) GetNode(rawID any) (any, error) {
	var id string
	switch raw := rawID.(type) {
	case string:
		id = raw
	case int:
		id = fmt.Sprintf("%d", raw)
	default:
		return nil, fmt.Errorf("invalid node id type: %T", rawID)
	}

	state := e.state.GetNodeState(id)
	if state == nil {
		return nil, nil
	}

	return map[string]any{
		"result": eval.NewAnyEnv(state.Result),
	}, nil
}

func (ctx *FlowContext) EvalTemplate(template string) (thing.Any, error) {
	res, err := eval.EvalTemplate(ctx, template, ctx.EvalCtx)
	if err != nil {
		return thing.Null, fmt.Errorf("failed to evaluate template: %w", err)
	}
	return thing.New(res), nil
}

type nodeEvalPatcher struct{}

func (p *nodeEvalPatcher) Visit(node *ast.Node) {
	accessor, ok := (*node).(*ast.MemberNode)
	if !ok {
		return
	}

	parent, ok := (accessor.Node).(*ast.IdentifierNode)
	if !ok {
		return
	}

	if parent.Value != "nodes" {
		return
	}

	ast.Patch(node, &ast.CallNode{
		Callee:    &ast.IdentifierNode{Value: "node"},
		Arguments: []ast.Node{accessor.Property},
	})
}
