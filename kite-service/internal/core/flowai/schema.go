package flowai

// outputSchema is the shape of the model's answer. Strict structured outputs
// require every field and no extra ones, so unused fields are null.
var outputSchema = map[string]any{
	"type":                 "object",
	"additionalProperties": false,
	"required":             []string{"message", "edits"},
	"properties": map[string]any{
		"message": map[string]any{
			"type":        "string",
			"description": "Short answer to the user.",
		},
		"edits": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"required": []string{
					"op", "ref", "type", "id", "after", "before", "handle",
					"source", "target", "data_json", "items_json", "reconnect",
				},
				"properties": map[string]any{
					"op": map[string]any{
						"type": "string",
						"enum": []string{"add_node", "update_node", "remove_node", "connect", "disconnect"},
					},
					"ref":        nullable("string", "add_node: name of the new block, like $ban."),
					"type":       nullable("string", "add_node: block type from the catalog."),
					"id":         nullable("string", "update_node and remove_node: the block to change."),
					"after":      nullable("string", "add_node: the block the new one runs after."),
					"before":     nullable("string", "add_node: the block the new one runs before."),
					"handle":     nullable("string", "add_node, connect and disconnect: output of the earlier block, if it has several."),
					"source":     nullable("string", "connect and disconnect: the earlier block."),
					"target":     nullable("string", "connect and disconnect: the later block."),
					"data_json":  nullable("string", "add_node and update_node: settings as a JSON object."),
					"items_json": nullable("string", "add_node of a condition: settings of each branch as a JSON array of objects."),
					"reconnect":  nullable("boolean", "remove_node: whether to connect the blocks after it to the block before it."),
				},
			},
		},
	},
}

func nullable(typ string, description string) map[string]any {
	return map[string]any{
		"type":        []string{typ, "null"},
		"description": description,
	}
}
