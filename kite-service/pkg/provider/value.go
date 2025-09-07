package provider

import (
	"context"
	"errors"

	"github.com/kitecloud/kite/kite-service/pkg/thing"
)

type ValueWithMetadata struct {
	Value    thing.Thing
	Metadata map[string]string
}

type MetadataOption func(metadata map[string]string)

func WithMetadata(key, value string) MetadataOption {
	return func(metadata map[string]string) {
		metadata[key] = value
	}
}

// ValueProvider provides access to arbitrary key-value pairs.
// Values are usually scoped to a specific plugin or other entity.
type ValueProvider interface {
	UpdateValue(
		ctx context.Context,
		key string,
		op VariableOperation,
		value thing.Thing,
		opts ...MetadataOption,
	) (thing.Thing, error)
	// GetValue returns the value for the given key.
	// If the value is not found, it returns thing.Null and no error.
	GetValue(ctx context.Context, key string) (thing.Thing, error)
	DeleteValue(ctx context.Context, key string) error
	SearchValues(ctx context.Context, opts ...MetadataOption) ([]thing.Thing, error)
}

type MockValueProvider struct {
	Values map[string]ValueWithMetadata
}

func (p *MockValueProvider) UpdateValue(
	ctx context.Context,
	key string,
	op VariableOperation,
	value thing.Thing,
	opts ...MetadataOption,
) (thing.Thing, error) {
	metadata := make(map[string]string, len(opts))
	for _, opt := range opts {
		opt(metadata)
	}

	currentValue, err := p.GetValue(ctx, key)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			p.Values[key] = ValueWithMetadata{
				Value:    value,
				Metadata: metadata,
			}
			return value, nil
		}
		return thing.Null, err
	}

	newValue := currentValue

	switch op {
	case VariableOperationOverwrite:
		newValue = value
	case VariableOperationAppend:
		newValue = currentValue.Append(value)
	case VariableOperationPrepend:
		newValue = value.Append(currentValue)
	case VariableOperationIncrement:
		newValue = currentValue.Add(value)
	case VariableOperationDecrement:
		newValue = currentValue.Sub(value)
	}

	p.Values[key] = ValueWithMetadata{
		Value:    newValue,
		Metadata: metadata,
	}
	return newValue, nil
}

func (p *MockValueProvider) GetValue(ctx context.Context, key string) (thing.Thing, error) {
	v, ok := p.Values[key]
	if !ok {
		return thing.Null, nil
	}
	return v.Value, nil
}

func (p *MockValueProvider) DeleteValue(ctx context.Context, key string) error {
	delete(p.Values, key)

	return nil
}

func (p *MockValueProvider) SearchValues(ctx context.Context, opts ...MetadataOption) ([]thing.Thing, error) {
	metadata := make(map[string]string, len(opts))
	for _, opt := range opts {
		opt(metadata)
	}

	values := make([]thing.Thing, 0, len(p.Values))
	for _, value := range p.Values {
		if len(metadata) > 0 {
			for k, v := range metadata {
				if v != value.Metadata[k] {
					continue
				}
			}
		}
		values = append(values, value.Value)
	}
	return values, nil
}
