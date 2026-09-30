package knowledge

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"strings"
)

// Catalog is kite-service/pkg/flow/catalog.json, which describes every block
// of the flow editor. It's generated from the editor's schemas.
type Catalog struct {
	Nodes map[string]CatalogNode `json:"nodes"`
}

type CatalogNode struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Contexts     []string `json:"contexts"`
	Outputs      []string `json:"outputs"`
	DataSchema   *Schema  `json:"data_schema"`
	ResultSchema *Schema  `json:"result_schema"`
}

type Schema struct {
	Description string             `json:"description"`
	Properties  map[string]*Schema `json:"properties"`
	Required    []string           `json:"required"`
	Enum        []any              `json:"enum"`
	Items       *Schema            `json:"items"`
	Templated   bool               `json:"x-templated"`
	UserPicked  bool               `json:"x-user-picked"`
}

func ParseCatalog(data []byte) (*Catalog, error) {
	var c Catalog
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse catalog: %w", err)
	}
	return &c, nil
}

var contextNames = map[string]string{
	"command":               "commands",
	"event_discord":         "Discord event listeners",
	"event_schedule":        "scheduled event listeners",
	"component_button":      "buttons",
	"component_select_menu": "select menus",
}

// Describe returns what the flow editor and the docs show about a block: its
// settings and the result later blocks can use.
func (c *Catalog) Describe(nodeType string) (string, bool) {
	n, ok := c.Nodes[nodeType]
	if !ok {
		return "", false
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Block `%s`: %s\n", n.Title, n.Description)

	contexts := make([]string, 0, len(n.Contexts))
	for _, ctx := range n.Contexts {
		if name, ok := contextNames[ctx]; ok {
			contexts = append(contexts, name)
		}
	}
	if len(contexts) > 0 {
		fmt.Fprintf(&b, "Available in: %s\n", strings.Join(contexts, ", "))
	}
	if len(n.Outputs) > 1 {
		fmt.Fprintf(&b, "Outputs: %s\n", strings.Join(n.Outputs, ", "))
	}

	if n.DataSchema != nil && len(n.DataSchema.Properties) > 0 {
		var settings strings.Builder
		for _, key := range slices.Sorted(maps.Keys(n.DataSchema.Properties)) {
			// Every block has it and it changes nothing.
			if key == "custom_label" {
				continue
			}
			// Keys are internal, the editor shows its own labels for them.
			s := n.DataSchema.Properties[key]
			if s.Description != "" {
				fmt.Fprintf(&settings, "- %s", s.Description)
			} else {
				fmt.Fprintf(&settings, "- %s", key)
			}
			var notes []string
			if slices.Contains(n.DataSchema.Required, key) {
				notes = append(notes, "required")
			}
			if s.Templated {
				notes = append(notes, "placeholders allowed")
			}
			if s.UserPicked {
				notes = append(notes, "picked from the app")
			}
			if len(notes) > 0 {
				fmt.Fprintf(&settings, " (%s)", strings.Join(notes, ", "))
			}
			if len(s.Enum) > 0 {
				fmt.Fprintf(&settings, " One of: %s.", joinAny(s.Enum))
			}
			settings.WriteByte('\n')
		}
		if settings.Len() > 0 {
			b.WriteString("Settings:\n")
			b.WriteString(settings.String())
		}
	}

	if n.ResultSchema != nil {
		var result strings.Builder
		writeResult(&result, "", n.ResultSchema, 0)
		if result.Len() > 0 {
			fmt.Fprintf(&b, "Result, used as {{result('<block id>').<field>}}:\n%s", result.String())
		}
	}

	return strings.TrimSpace(b.String()), true
}

func writeResult(b *strings.Builder, prefix string, s *Schema, depth int) {
	if s.Items != nil {
		s = s.Items
	}
	for _, key := range slices.Sorted(maps.Keys(s.Properties)) {
		p := s.Properties[key]
		path := prefix + key
		fmt.Fprintf(b, "- %s", path)
		if p.Description != "" {
			fmt.Fprintf(b, ": %s", p.Description)
		}
		b.WriteByte('\n')
		if depth < 1 {
			writeResult(b, path+".", p, depth+1)
		}
	}
}

func joinAny(values []any) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = fmt.Sprint(v)
	}
	return strings.Join(parts, ", ")
}
