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

func (e *nodeEvalEnv) GetNode(rawKey any) (any, error) {
	var key string
	switch raw := rawKey.(type) {
	case string:
		key = raw
	case int:
		key = fmt.Sprintf("%d", raw)
	default:
		return nil, fmt.Errorf("invalid node key type: %T", rawKey)
	}

	return map[string]any{
		"result": eval.NewAnyEnv(e.state.GetNodeResult(key)),
	}, nil
}

func (ctx *FlowContext) EvalTemplate(template string) (thing.Thing, error) {
	res, err := eval.EvalTemplate(ctx, template, ctx.EvalCtx)
	if err != nil {
		return thing.Null, fmt.Errorf("failed to evaluate template: %w", err)
	}
	return res, nil
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
