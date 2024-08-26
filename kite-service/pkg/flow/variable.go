package flow

import "github.com/kitecloud/kite/kite-service/pkg/placeholder"

type FlowContextVariables struct {
	Variables map[string]FlowValue
}

func NewContextVariables() FlowContextVariables {
	return FlowContextVariables{
		Variables: make(map[string]FlowValue),
	}
}

func (v *FlowContextVariables) SetVariable(name string, value FlowValue) {
	v.Variables[name] = value
}

func (v *FlowContextVariables) Variable(name string) (FlowValue, bool) {
	value, ok := v.Variables[name]
	return value, ok
}

func (v *FlowContextVariables) GetPlaceholder(key string) (placeholder.Provider, error) {
	if key == "test" {
		return placeholder.NewStringProvider("yeet"), nil
	}

	value, ok := v.Variable(key)
	if !ok {
		return nil, placeholder.ErrNotFound
	}
	return value, nil
}

func (v *FlowContextVariables) ResolvePlaceholder() (string, error) {
	return "", nil
}
