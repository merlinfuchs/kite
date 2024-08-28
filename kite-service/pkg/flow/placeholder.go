package flow

import (
	"context"

	"github.com/kitecloud/kite/kite-service/pkg/placeholder"
)

type variablePlaceholderProvider struct {
	Variable FlowVariableProvider
}

func (p *variablePlaceholderProvider) GetPlaceholder(ctx context.Context, key string) (placeholder.Provider, error) {
	value, err := p.Variable.GetVariable(ctx, key)
	if err != nil {
		if err == ErrNotFound {
			return nil, placeholder.ErrNotFound
		}
		return nil, err
	}
	return value, nil
}

func (p *variablePlaceholderProvider) ResolvePlaceholder(ctx context.Context) (string, error) {
	return "", nil
}

type nodePlaceholderProvider struct {
	node *CompiledFlowNode
}

func (p *nodePlaceholderProvider) setNode(node *CompiledFlowNode) {
	p.node = node
}

func (p *nodePlaceholderProvider) GetPlaceholder(ctx context.Context, key string) (placeholder.Provider, error) {
	if key == "result" {
		res := p.node.State.Result
		if res.IsNull() {
			return nil, placeholder.ErrNotFound
		}
		return res, nil
	}

	node := p.node.FindParentWithID(key)
	if node == nil {
		return nil, placeholder.ErrNotFound
	}

	return &nodePlaceholderProvider{
		node: node,
	}, nil
}

func (p *nodePlaceholderProvider) ResolvePlaceholder(ctx context.Context) (string, error) {
	return "", nil
}
