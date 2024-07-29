package flow

import "github.com/kitecloud/kite/kite-common/core/template"

type FlowContextVariables struct {
	*template.TemplateContext

	Variables map[string]FlowValue
}

func (v *FlowContextVariables) SetVariable(name string, value FlowValue) {
	v.Variables[name] = value

	var tValue interface{}
	switch value.Type {
	case FlowValueTypeNull:
		v = nil
	case FlowValueTypeString:
		tValue = value.String()
	case FlowValueTypeNumber:
		tValue = value.Number()
	case FlowValueTypeMessage:
		// TODO: replace with custom message struct?
		tValue, _ = value.Message()
	}

	v.SetData("Variables."+name, tValue)
}

func (v *FlowContextVariables) Variable(name string) FlowValue {
	value := v.Variables[name]
	return value
}
