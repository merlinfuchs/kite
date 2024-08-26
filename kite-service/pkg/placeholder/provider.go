package placeholder

import (
	"context"
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("placeholder not found")

type Provider interface {
	GetPlaceholder(ctx context.Context, key string) (Provider, error)
	ResolvePlaceholder(ctx context.Context) (string, error)
}

type StringProvider struct {
	value string
}

func NewStringProvider(value string) StringProvider {
	return StringProvider{
		value: value,
	}
}

func (s StringProvider) GetPlaceholder(ctx context.Context, key string) (Provider, error) {
	return nil, ErrNotFound
}

func (s StringProvider) ResolvePlaceholder(ctx context.Context) (string, error) {
	return s.value, nil
}

type StringerProvider[T fmt.Stringer] struct {
	value T
}

func NewStringerProvider[T fmt.Stringer](value T) StringerProvider[T] {
	return StringerProvider[T]{
		value: value,
	}
}

func (s StringerProvider[T]) GetPlaceholder(ctx context.Context, key string) (Provider, error) {
	return nil, ErrNotFound
}

func (s StringerProvider[T]) ResolvePlaceholder(ctx context.Context) (string, error) {
	return s.value.String(), nil
}

type MapProvider[T Provider] struct {
	data map[string]T
}

func NewMapProvider[T Provider](data map[string]T) MapProvider[T] {
	return MapProvider[T]{
		data: data,
	}
}

func (m MapProvider[T]) Set(key string, value T) {
	m.data[key] = value
}

func (m MapProvider[T]) Delete(key string) {
	delete(m.data, key)
}

func (m MapProvider[T]) GetPlaceholder(ctx context.Context, key string) (Provider, error) {
	provider, ok := m.data[key]
	if !ok {
		return nil, ErrNotFound
	}
	return provider, nil
}

func (m MapProvider[T]) ResolvePlaceholder(ctx context.Context) (string, error) {
	return "", nil
}
