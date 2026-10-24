package flow

import (
	"encoding/json"
	"regexp"
	"slices"
)

var (
	// ParameterKeyPattern is what a key of a parameter can be.
	ParameterKeyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

	templatePattern = regexp.MustCompile(`{{(.*?)}}`)
	// Not after a dot, as webhook.data.params.count isn't a parameter.
	paramReferencePattern = regexp.MustCompile(`(?:^|[^.\w])params\.([A-Za-z_][A-Za-z0-9_]*)`)
)

// ReferencedParameters returns the keys of the parameters the blocks use, in
// the order they're first used.
func (d FlowData) ReferencedParameters() []string {
	var keys []string
	for _, node := range d.Nodes {
		raw, err := json.Marshal(node.Data)
		if err != nil {
			continue
		}
		for _, template := range templatePattern.FindAllStringSubmatch(string(raw), -1) {
			for _, ref := range paramReferencePattern.FindAllStringSubmatch(template[1], -1) {
				if !slices.Contains(keys, ref[1]) {
					keys = append(keys, ref[1])
				}
			}
		}
	}
	return keys
}

// ParameterValues returns the values of the parameters by their key.
func (d FlowData) ParameterValues() map[string]string {
	values := make(map[string]string, len(d.Parameters))
	for _, p := range d.Parameters {
		values[p.Key] = p.Value
	}
	return values
}
