package flow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

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

	return map[string]any{
		"result": eval.NewThingEnv(e.state.GetNodeResult(id)),
	}, nil
}

func (ctx *FlowContext) EvalTemplate(template string) (thing.Thing, error) {
	res, err := eval.EvalTemplate(ctx, template, ctx.EvalCtx)
	if err != nil {
		return thing.Null, fmt.Errorf("failed to evaluate template: %w", err)
	}
	return res, nil
}

func (ctx *FlowContext) EvalTemplateKeepSpace(template string) (thing.Thing, error) {
	res, err := eval.EvalTemplateKeepSpace(ctx, template, ctx.EvalCtx)
	if err != nil {
		return thing.Null, fmt.Errorf("failed to evaluate template: %w", err)
	}
	return res, nil
}

// evalKeyValues evaluates the values of key value pairs, like the parameters
// of a request.
func (ctx *FlowContext) evalKeyValues(pairs []HTTPRequestDataKeyValue) (map[string]thing.Thing, error) {
	res := make(map[string]thing.Thing, len(pairs))
	for _, pair := range pairs {
		value, err := ctx.EvalTemplate(pair.Value)
		if err != nil {
			return nil, err
		}
		res[pair.Key] = value
	}
	return res, nil
}

// EvalJSONTemplate evaluates the placeholders in the string values of a JSON
// document. Unlike evaluating the whole document as one template, results
// can't break its syntax. A value that is a single placeholder keeps the type
// of its result, so numbers and lists can be filled in too.
func (ctx *FlowContext) EvalJSONTemplate(raw json.RawMessage) ([]byte, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	// Keeps IDs written as numbers exact.
	dec.UseNumber()

	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	v, err := ctx.evalJSONValue(v)
	if err != nil {
		return nil, err
	}
	return json.Marshal(v)
}

func (ctx *FlowContext) evalJSONValue(v any) (any, error) {
	switch v := v.(type) {
	case map[string]any:
		for key, value := range v {
			res, err := ctx.evalJSONValue(value)
			if err != nil {
				return nil, err
			}
			v[key] = res
		}
		return v, nil
	case []any:
		for i, value := range v {
			res, err := ctx.evalJSONValue(value)
			if err != nil {
				return nil, err
			}
			v[i] = res
		}
		return v, nil
	case string:
		// An empty template evaluates to null, which isn't what a plain string
		// should become.
		if !strings.Contains(v, "{{") {
			return v, nil
		}

		res, err := ctx.EvalTemplateKeepSpace(v)
		if err != nil {
			return nil, err
		}
		return res.JSONValue(), nil
	default:
		return v, nil
	}
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
