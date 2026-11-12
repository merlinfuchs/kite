package provider

import (
	"context"
	"errors"

	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

// ValueProvider provides access to arbitrary key-value pairs.
// Values are usually scoped to a specific plugin or other entity.
type ValueProvider interface {
	UpdateValue(ctx context.Context, key string, op VariableOperation, value thing.Thing) (thing.Thing, error)
	// GetValue returns the value for the given key.
	// If the value is not found, it returns thing.Null and no error.
	GetValue(ctx context.Context, key string) (thing.Thing, error)
	DeleteValue(ctx context.Context, key string) error
}

type MockValueProvider struct {
	Values map[string]thing.Thing
}

func (p *MockValueProvider) UpdateValue(ctx context.Context, key string, op VariableOperation, value thing.Thing) (thing.Thing, error) {
	currentValue, err := p.GetValue(ctx, key)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return thing.Null, err
	}
	newValue := op.Apply(currentValue, err == nil, value)

	p.Values[key] = newValue
	return newValue, nil
}

func (p *MockValueProvider) GetValue(ctx context.Context, key string) (thing.Thing, error) {
	v, ok := p.Values[key]
	if !ok {
		return thing.Null, nil
	}
	return v, nil
}

func (p *MockValueProvider) DeleteValue(ctx context.Context, key string) (thing.Thing, error) {
	v, ok := p.Values[key]
	if !ok {
		return thing.Null, ErrNotFound
	}

	delete(p.Values, key)

	return v, nil
}
