package flow

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
)

var (
	// ParameterKeyPattern is what a key of a parameter can be.
	ParameterKeyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

	// JSON escapes < > and &, which would hide a reference like {{1<params.max}}
	// from the pattern.
	htmlUnescaper = strings.NewReplacer(`\u003c`, "<", `\u003e`, ">", `\u0026`, "&")

	// Not after a dot, as webhook.data.params.count isn't a parameter.
	paramReferenceRe = regexp.MustCompile(`(?:^|[^.\w])params\.([A-Za-z_][A-Za-z0-9_]*)`)

	paramPlaceholderRe = regexp.MustCompile(`\{\{\s*params\.([A-Za-z_][A-Za-z0-9_]*)\s*\}\}`)
)

// withParameters replaces every {{params.<key>}} in value with the value of
// the parameter. It's for settings that aren't templates, like the value of an
// event filter, so other placeholders are left as they are.
func withParameters(value string, params map[string]string) string {
	return paramPlaceholderRe.ReplaceAllStringFunc(value, func(m string) string {
		return params[paramPlaceholderRe.FindStringSubmatch(m)[1]]
	})
}

// ReferencedParameters returns the keys of the parameters the blocks use, in
// the order they're first used. Notes only mention them.
func (d FlowData) ReferencedParameters() []string {
	templates := make([]string, 0, len(d.Nodes))
	for _, node := range d.Nodes {
		if node.Type == FlowNodeTypeNote {
			continue
		}
		raw, err := json.Marshal(node.Data)
		if err != nil {
			continue
		}
		templates = append(templates, htmlUnescaper.Replace(string(raw)))
	}
	return placeholderReferences(templates, paramReferenceRe)
}

// UsedParameters returns the parameters the blocks use, the ones the flow
// doesn't define as required text without a value.
func (d FlowData) UsedParameters() []FlowParameter {
	keys := d.ReferencedParameters()
	params := make([]FlowParameter, len(keys))
	for i, key := range keys {
		params[i] = FlowParameter{Key: key, Type: FlowParameterTypeText, Required: true}
		for _, p := range d.Parameters {
			if p.Key == key {
				params[i] = p
				break
			}
		}
	}
	return params
}

// ParameterValues returns the values of the parameters by their key.
func (d FlowData) ParameterValues() map[string]string {
	values := make(map[string]string, len(d.Parameters))
	for _, p := range d.Parameters {
		values[p.Key] = p.Value
	}
	return values
}

// WithParameterValues returns the flow with the values of the parameters it
// has, and whether one of them changed. Other values are left out.
func (d FlowData) WithParameterValues(values map[string]string) (FlowData, bool) {
	params := make([]FlowParameter, len(d.Parameters))
	changed := false
	for i, p := range d.Parameters {
		if value, ok := values[p.Key]; ok && value != p.Value {
			p.Value = value
			changed = true
		}
		params[i] = p
	}
	d.Parameters = params
	return d, changed
}

// DisplayLabel is the label of the parameter, or its key without one.
func (p FlowParameter) DisplayLabel() string {
	if p.Label != "" {
		return p.Label
	}
	return p.Key
}

// CheckValue reports a value that doesn't fit the type of the parameter, like
// a channel that isn't an ID. Empty values fit any type.
func (p FlowParameter) CheckValue(value string) error {
	if value == "" || !p.Type.Picked() {
		return nil
	}
	if id, err := discord.ParseSnowflake(value); err != nil || !id.IsValid() {
		return fmt.Errorf("%s must be the ID of a %s", p.DisplayLabel(), p.Type)
	}
	return nil
}
