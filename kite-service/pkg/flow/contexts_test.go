package flow

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The editor offers blocks by its own categories and block definitions, which
// catalog.json is generated from. They must agree with blockRules, or the
// editor offers blocks that can't be saved, or hides ones that can.
func TestBlockRulesMatchCatalog(t *testing.T) {
	var catalog struct {
		Nodes map[string]struct {
			Contexts         []FlowContextType `json:"contexts"`
			NeedsInteraction bool              `json:"needs_interaction"`
		} `json:"nodes"`
	}
	require.NoError(t, json.Unmarshal(CatalogJSON, &catalog))

	for nodeType, node := range catalog.Nodes {
		var allowed []FlowContextType
		for _, context := range FlowContextTypes {
			if BlockAllowedIn(FlowNodeType(nodeType), context) {
				allowed = append(allowed, context)
			}
		}
		assert.Equalf(t, node.Contexts, allowed, "contexts of %s", nodeType)
		assert.Equalf(t, node.NeedsInteraction, BlockNeedsInteraction(FlowNodeType(nodeType)), "needs_interaction of %s", nodeType)
	}

	for nodeType := range blockRules {
		assert.Containsf(t, catalog.Nodes, string(nodeType), "%s has a rule, but isn't a block", nodeType)
	}
}

// FlowContext can't tell flows of buttons and select menus apart.
func TestBlockRulesSameForButtonsAndSelectMenus(t *testing.T) {
	for nodeType := range blockRules {
		assert.Equalf(t,
			BlockAllowedIn(nodeType, FlowContextComponentButton),
			BlockAllowedIn(nodeType, FlowContextComponentSelectMenu),
			"%s", nodeType,
		)
	}
}
