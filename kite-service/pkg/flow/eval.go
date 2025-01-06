package flow

import (
	"context"
	"fmt"
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
		"result": state.Result.EvalEnv(),
	}, nil
}
