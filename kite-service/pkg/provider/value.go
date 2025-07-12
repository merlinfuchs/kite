package provider

import (
	"encoding/json"
	"strconv"
)

// ValueProvider provides access to arbitrary key-value pairs.
type ValueProvider interface {
	SetKey(key string, value json.RawMessage) error
	GetKey(key string) (json.RawMessage, error)
	IncreaseKey(key string, amount int) (int, error)
	DecreaseKey(key string, amount int) (int, error)
	DeleteKey(key string) (json.RawMessage, error)
}

type MockValueProvider struct {
	Values map[string]json.RawMessage
}

func (p *MockValueProvider) SetKey(key string, value json.RawMessage) error {
	p.Values[key] = value
	return nil
}

func (p *MockValueProvider) GetKey(key string) (json.RawMessage, error) {
	v, ok := p.Values[key]
	if !ok {
		return nil, ErrNotFound
	}
	return v, nil
}

func (p *MockValueProvider) IncreaseKey(key string, amount int) (int, error) {
	v, ok := p.Values[key]
	if !ok {
		p.Values[key] = json.RawMessage(strconv.Itoa(amount))
		return amount, nil
	}

	var value int
	if err := json.Unmarshal(v, &value); err != nil {
		return 0, err
	}

	value += amount

	p.Values[key] = json.RawMessage(strconv.Itoa(value))

	return value, nil
}

func (p *MockValueProvider) DecreaseKey(key string, amount int) (int, error) {
	v, ok := p.Values[key]
	if !ok {
		return 0, ErrNotFound
	}

	var value int
	if err := json.Unmarshal(v, &value); err != nil {
		return 0, err
	}

	value -= amount

	p.Values[key] = json.RawMessage(strconv.Itoa(value))

	return value, nil
}

func (p *MockValueProvider) DeleteKey(key string) (json.RawMessage, error) {
	v, ok := p.Values[key]
	if !ok {
		return nil, ErrNotFound
	}

	delete(p.Values, key)

	return v, nil
}
