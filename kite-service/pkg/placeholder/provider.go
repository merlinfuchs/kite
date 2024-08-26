package placeholder

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("placeholder not found")

type Provider interface {
	GetPlaceholder(key string) (Provider, error)
	ResolvePlaceholder() (string, error)
}

type StringProvider struct {
	value string
}

func NewStringProvider(value string) StringProvider {
	return StringProvider{
		value: value,
	}
}

func (s StringProvider) GetPlaceholder(key string) (Provider, error) {
	return nil, ErrNotFound
}

func (s StringProvider) ResolvePlaceholder() (string, error) {
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

func (s StringerProvider[T]) GetPlaceholder(key string) (Provider, error) {
	return nil, ErrNotFound
}

func (s StringerProvider[T]) ResolvePlaceholder() (string, error) {
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

func (m MapProvider[T]) GetPlaceholder(key string) (Provider, error) {
	provider, ok := m.data[key]
	if !ok {
		return nil, ErrNotFound
	}
	return provider, nil
}

func (m MapProvider[T]) ResolvePlaceholder() (string, error) {
	return "", nil
}
