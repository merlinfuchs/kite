package flow

import (
	"context"
	"fmt"

	"github.com/kitecloud/kite/kite-service/pkg/eval"
)

type nodeEvalEnv struct {
	state *FlowContextState
}

func (e *nodeEvalEnv) GetNode(ctx context.Context, rawID any) (any, error) {
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
		"result": state.Result.Inner,
	}, nil
}

func (ctx *FlowContext) EvalTemplate(template string) (Value, error) {
	res, err := eval.EvalTemplate(ctx, template, ctx.EvalEnv)
	if err != nil {
		return Value{}, fmt.Errorf("failed to evaluate template: %w", err)
	}
	return Value{Inner: res}, nil
}
