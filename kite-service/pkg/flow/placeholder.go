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

type flowStatePlaceholderProvider struct {
	state *FlowContextState
}

func (p *flowStatePlaceholderProvider) GetPlaceholder(ctx context.Context, key string) (placeholder.Provider, error) {
	state := p.state.GetNodeState(key)
	if state == nil {
		return nil, placeholder.ErrNotFound
	}

	return &flowNodeStatePlaceholderProvider{
		state: state,
	}, nil
}

func (p *flowStatePlaceholderProvider) ResolvePlaceholder(ctx context.Context) (string, error) {
	return "", nil
}

type flowNodeStatePlaceholderProvider struct {
	state *FlowContextNodeState
}

func (p *flowNodeStatePlaceholderProvider) GetPlaceholder(ctx context.Context, key string) (placeholder.Provider, error) {
	if key == "result" {
		res := p.state.Result
		if res.IsNull() {
			return nil, placeholder.ErrNotFound
		}
		return res, nil
	}

	return nil, placeholder.ErrNotFound
}

func (p *flowNodeStatePlaceholderProvider) ResolvePlaceholder(ctx context.Context) (string, error) {
	return "", nil
}
