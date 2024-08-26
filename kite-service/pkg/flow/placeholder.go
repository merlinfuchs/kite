package flow

import (
	"context"

	"github.com/kitecloud/kite/kite-service/pkg/placeholder"
)

type VariablePlaceholderProvider struct {
	Variable FlowVariableProvider
}

func (p *VariablePlaceholderProvider) GetPlaceholder(ctx context.Context, key string) (placeholder.Provider, error) {
	value, err := p.Variable.GetVariable(ctx, key)
	if err != nil {
		if err == ErrNotFound {
			return nil, placeholder.ErrNotFound
		}
		return nil, err
	}
	return value, nil
}

func (p *VariablePlaceholderProvider) ResolvePlaceholder(ctx context.Context) (string, error) {
	return "", nil
}
