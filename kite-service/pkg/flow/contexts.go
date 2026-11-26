package flow

import (
	"fmt"
	"slices"
)

// FlowContextType is the kind of flow a block is in. They're the same as
// flowContextTypes in kite-web/src/lib/flow/context.tsx.
type FlowContextType string

const (
	FlowContextCommand             FlowContextType = "command"
	FlowContextComponentButton     FlowContextType = "component_button"
	FlowContextComponentSelectMenu FlowContextType = "component_select_menu"
	FlowContextEventDiscord        FlowContextType = "event_discord"
	FlowContextEventSchedule       FlowContextType = "event_schedule"
	FlowContextEventWebhook        FlowContextType = "event_webhook"
)

// FlowContextTypes are all kinds of flows.
var FlowContextTypes = []FlowContextType{
	FlowContextCommand,
	FlowContextComponentButton,
	FlowContextComponentSelectMenu,
	FlowContextEventDiscord,
	FlowContextEventSchedule,
	FlowContextEventWebhook,
}

var (
	componentContexts = []FlowContextType{FlowContextComponentButton, FlowContextComponentSelectMenu}
	eventContexts     = []FlowContextType{FlowContextEventDiscord, FlowContextEventSchedule, FlowContextEventWebhook}
)

// blockRule says where a block can be used.
type blockRule struct {
	// Contexts are the kinds of flows the block can be in, all if nil.
	Contexts []FlowContextType
	// NeedsInteraction is true for blocks that respond to an interaction.
	// Event listeners have none until a button or select menu on a message
	// they sent is used, so they can only use them in component branches.
	NeedsInteraction bool
}

// blockRules says where the blocks can be used that can't be used anywhere.
// The editor offers blocks by the same rules: catalog.json, which is
// generated from it, is tested against this in TestBlockRulesMatchCatalog.
//
// Saving a flow checks them, see CheckBlocks.
var blockRules = map[FlowNodeType]blockRule{
	FlowNodeTypeEntryCommand:         {Contexts: []FlowContextType{FlowContextCommand}},
	FlowNodeTypeEntryComponentButton: {Contexts: componentContexts},
	FlowNodeTypeEntryEvent:           {Contexts: eventContexts},

	FlowNodeTypeOptionCommandArgument:    {Contexts: []FlowContextType{FlowContextCommand}},
	FlowNodeTypeOptionCommandPermissions: {Contexts: []FlowContextType{FlowContextCommand}},
	FlowNodeTypeOptionCommandContexts:    {Contexts: []FlowContextType{FlowContextCommand}},
	FlowNodeTypeOptionCommandCooldown:    {Contexts: []FlowContextType{FlowContextCommand}},
	// Only events from Discord have something to filter on.
	FlowNodeTypeOptionEventFilter: {Contexts: []FlowContextType{FlowContextEventDiscord}},

	FlowNodeTypeActionResponseCreate: {NeedsInteraction: true},
	FlowNodeTypeActionResponseEdit:   {NeedsInteraction: true},
	FlowNodeTypeActionResponseDelete: {NeedsInteraction: true},
	FlowNodeTypeActionResponseDefer:  {NeedsInteraction: true},
	FlowNodeTypeSuspendResponseModal: {NeedsInteraction: true},

	FlowNodeTypeActionWebhookResponseCreate: {Contexts: []FlowContextType{FlowContextEventWebhook}},
	FlowNodeTypeControlWebhookVerify:        {Contexts: []FlowContextType{FlowContextEventWebhook}},

	// Scheduled and webhook runs aren't in a server.
	FlowNodeTypeActionServerLeave: {Contexts: []FlowContextType{
		FlowContextCommand,
		FlowContextComponentButton,
		FlowContextComponentSelectMenu,
		FlowContextEventDiscord,
	}},
}

// BlockAllowedIn reports whether blocks of type t can be used in flows of the
// given kind.
func BlockAllowedIn(t FlowNodeType, context FlowContextType) bool {
	rule, ok := blockRules[t]
	return !ok || rule.Contexts == nil || slices.Contains(rule.Contexts, context)
}

// BlockNeedsInteraction reports whether blocks of type t respond to an
// interaction.
func BlockNeedsInteraction(t FlowNodeType) bool {
	return blockRules[t].NeedsInteraction
}

// FlowContext returns the kind of flow n is the entry of. Flows of buttons and
// select menus can use the same blocks, so they are both component_button.
func (n *CompiledFlowNode) FlowContext() FlowContextType {
	switch {
	case n.IsCommandEntry():
		return FlowContextCommand
	case n.IsScheduleEntry():
		return FlowContextEventSchedule
	case n.IsWebhookEntry():
		return FlowContextEventWebhook
	case n.IsEventListenerEntry():
		return FlowContextEventDiscord
	default:
		return FlowContextComponentButton
	}
}

func flowContextName(context FlowContextType) string {
	switch context {
	case FlowContextCommand:
		return "commands"
	case FlowContextComponentButton, FlowContextComponentSelectMenu:
		return "button and select menu flows"
	case FlowContextEventDiscord:
		return "Discord event listeners"
	case FlowContextEventSchedule:
		return "scheduled event listeners"
	case FlowContextEventWebhook:
		return "webhook event listeners"
	}
	return string(context) + " flows"
}

// CheckBlocks reports blocks of data that can't be used in the flow entry was
// compiled from: blocks that are for other kinds of flows, like event filters
// in commands, and blocks that respond to an interaction before a button sent
// by an event listener is used.
//
// Saving checks it, not compiling, so flows saved before keep running.
func CheckBlocks(data FlowData, entry *CompiledFlowNode) error {
	context := entry.FlowContext()
	for _, node := range data.Nodes {
		if !BlockAllowedIn(node.Type, context) {
			return fmt.Errorf("block %s can't be used in %s", node.Type, flowContextName(context))
		}
	}

	if entry.IsEventListenerEntry() {
		if node := firstWithoutInteraction(entry, BlockNeedsInteraction); node != nil {
			return fmt.Errorf("block %s can't be used in %s outside of button branches", node.Type, flowContextName(context))
		}
	}
	return nil
}
