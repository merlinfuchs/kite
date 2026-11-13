package bundle

import (
	"encoding/json"
	"fmt"
	"strings"
)

// CardContent is the content without what the cards of the marketplace don't
// show, for lists of many bundles: the blocks of its flows other than their
// entry, which has the name, and the content of its messages. Other fields
// stay as they are.
func CardContent(content json.RawMessage) (json.RawMessage, error) {
	var c map[string]json.RawMessage
	if err := json.Unmarshal(content, &c); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bundle content: %w", err)
	}

	for _, key := range []string{"commands", "event_listeners"} {
		err := mapItems(c, key, func(item map[string]json.RawMessage) error {
			flowSource, err := entryOnly(item["flow_source"])
			item["flow_source"] = flowSource
			return err
		})
		if err != nil {
			return nil, err
		}
	}
	err := mapItems(c, "messages", func(item map[string]json.RawMessage) error {
		item["data"] = json.RawMessage("{}")
		item["flow_sources"] = json.RawMessage("{}")
		return nil
	})
	if err != nil {
		return nil, err
	}
	return json.Marshal(c)
}

func mapItems(c map[string]json.RawMessage, key string, f func(map[string]json.RawMessage) error) error {
	raw, ok := c[key]
	if !ok {
		return nil
	}
	var items []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &items); err != nil {
		return fmt.Errorf("failed to unmarshal bundle %s: %w", key, err)
	}
	for _, item := range items {
		if err := f(item); err != nil {
			return err
		}
	}
	var err error
	c[key], err = json.Marshal(items)
	return err
}

// entryOnly is the flow with only its entry blocks and no edges.
func entryOnly(raw json.RawMessage) (json.RawMessage, error) {
	var f struct {
		Nodes []json.RawMessage `json:"nodes"`
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		return nil, fmt.Errorf("failed to unmarshal bundle flow: %w", err)
	}

	nodes := []json.RawMessage{}
	for _, n := range f.Nodes {
		var node struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(n, &node); err != nil {
			return nil, fmt.Errorf("failed to unmarshal bundle flow node: %w", err)
		}
		if strings.HasPrefix(node.Type, "entry_") {
			nodes = append(nodes, n)
		}
	}
	return json.Marshal(map[string]any{"nodes": nodes, "edges": []any{}})
}
