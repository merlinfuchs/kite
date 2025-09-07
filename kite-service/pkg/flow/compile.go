package flow

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"github.com/diamondburned/arikawa/v3/utils/json/option"
)

type FlowCompiler struct{}

func CompileCommand(data FlowData) (*CompiledFlowNode, error) {
	return compile(data, FlowNodeTypeEntryCommand)
}

func CompileComponentButton(data FlowData) (*CompiledFlowNode, error) {
	return compile(data, FlowNodeTypeEntryComponentButton)
}

func CompileEventListener(data FlowData) (*CompiledFlowNode, error) {
	return compile(data, FlowNodeTypeEntryEvent)
}

func compile(data FlowData, entryType FlowNodeType) (*CompiledFlowNode, error) {
	var entryNode *CompiledFlowNode
	nodeMap := make(map[string]*CompiledFlowNode)
	for _, node := range data.Nodes {
		compiledNode := &CompiledFlowNode{
			ID:   node.ID,
			Type: node.Type,
			Data: node.Data,
			Parents: ConnectedFlowNodes{
				Handles: make(map[string][]*CompiledFlowNode),
			},
			Children: ConnectedFlowNodes{
				Handles: make(map[string][]*CompiledFlowNode),
			},
		}
		nodeMap[node.ID] = compiledNode

		if node.Type == entryType {
			entryNode = compiledNode
		}
	}

	if entryNode == nil {
		return nil, fmt.Errorf("entry node not found")
	}

	for _, edge := range data.Edges {
		parent, ok := nodeMap[edge.Source]
		if !ok {
			continue
		}

		child, ok := nodeMap[edge.Target]
		if !ok {
			continue
		}

		if !edge.SourceHandle.Valid || edge.SourceHandle.String == "" || edge.SourceHandle.String == "default" {
			parent.Children.Default = append(parent.Children.Default, child)
		} else {
			if _, ok := parent.Children.Handles[edge.SourceHandle.String]; !ok {
				parent.Children.Handles[edge.SourceHandle.String] = []*CompiledFlowNode{
					child,
				}
			} else {
				parent.Children.Handles[edge.SourceHandle.String] = append(parent.Children.Handles[edge.SourceHandle.String], child)
			}
		}

		if !edge.TargetHandle.Valid {
			child.Parents.Default = append(child.Parents.Default, parent)
		} else {
			if _, ok := child.Parents.Handles[edge.TargetHandle.String]; !ok {
				child.Parents.Handles[edge.TargetHandle.String] = []*CompiledFlowNode{
					parent,
				}
			} else {
				child.Parents.Handles[edge.TargetHandle.String] = append(child.Parents.Handles[edge.TargetHandle.String], parent)
			}
		}
	}

	return entryNode, nil
}

type CompiledFlowNode struct {
	ID       string
	Type     FlowNodeType
	Data     FlowNodeData
	Parents  ConnectedFlowNodes
	Children ConnectedFlowNodes
}

type ConnectedFlowNodes struct {
	Default []*CompiledFlowNode
	Handles map[string][]*CompiledFlowNode
}

func (n *CompiledFlowNode) IsEntry() bool {
	return n.Type == FlowNodeTypeEntryCommand ||
		n.Type == FlowNodeTypeEntryComponentButton ||
		n.Type == FlowNodeTypeEntryEvent
}

func (n *CompiledFlowNode) IsComponentButtonEntry() bool {
	return n.Type == FlowNodeTypeEntryComponentButton
}

func (n *CompiledFlowNode) IsEventListenerEntry() bool {
	return n.Type == FlowNodeTypeEntryEvent
}

func (n *CompiledFlowNode) IsCommandEntry() bool {
	return n.Type == FlowNodeTypeEntryCommand
}

func (n *CompiledFlowNode) IsCommandArgument() bool {
	return n.Type == FlowNodeTypeOptionCommandArgument
}

func (n *CompiledFlowNode) IsCommandPermissions() bool {
	return n.Type == FlowNodeTypeOptionCommandPermissions
}

func (n *CompiledFlowNode) IsCommandContexts() bool {
	return n.Type == FlowNodeTypeOptionCommandContexts
}

func (n *CompiledFlowNode) IsEventFilter() bool {
	return n.Type == FlowNodeTypeOptionEventFilter
}

func (n *CompiledFlowNode) CommandData() discord.Command {
	res := discord.Command{
		Name:                     n.CommandName(),
		Options:                  n.CommandArguments(),
		Description:              n.CommandDescription(),
		DefaultMemberPermissions: n.CommandPermissions(),
		Contexts:                 n.CommandContexts(),
	}

	namesParts := strings.Split(n.Data.Name, " ")
	if len(namesParts) == 1 {
		return res
	} else {
		res.Name = namesParts[0]

		var args []discord.CommandOptionValue
		for _, o := range res.Options {
			args = append(args, o.(discord.CommandOptionValue))
		}

		if len(namesParts) == 2 {
			res.Options = discord.CommandOptions{
				discord.NewSubcommandOption(
					namesParts[1],
					res.Description,
					args...,
				),
			}
		} else {
			res.Options = discord.CommandOptions{
				discord.NewSubcommandGroupOption(
					namesParts[1],
					res.Description,
					discord.NewSubcommandOption(
						namesParts[2],
						res.Description,
						args...,
					),
				),
			}
		}
	}

	return res
}

func (n *CompiledFlowNode) CommandName() string {
	if !n.IsCommandEntry() {
		return ""
	}
	return n.Data.Name
}

func (n *CompiledFlowNode) CommandDescription() string {
	if !n.IsCommandEntry() {
		return ""
	}
	return n.Data.Description
}

func (n *CompiledFlowNode) CommandArguments() discord.CommandOptions {
	res := make(discord.CommandOptions, 0)
	for _, node := range n.Parents.Default {
		if node.IsCommandArgument() {
			var o discord.CommandOption

			switch node.Data.CommandArgumentType {
			case CommandArgumentTypeString:
				var maxLength option.Int
				if node.Data.CommandArgumentMaxLength != 0 {
					maxLength = option.NewInt(node.Data.CommandArgumentMaxLength)
				}

				var choices []discord.StringChoice
				for _, choice := range node.Data.CommandArgumentChoices {
					if choice.Name == "" || choice.Value == "" {
						continue
					}

					choices = append(choices, discord.StringChoice{
						Name:  choice.Name,
						Value: choice.Value,
					})
				}

				o = &discord.StringOption{
					OptionName:  node.Data.Name,
					Description: node.Data.Description,
					Required:    node.Data.CommandArgumentRequired,
					MaxLength:   maxLength,
					Choices:     choices,
				}
			case CommandArgumentTypeInteger:
				var minValue option.Int
				if node.Data.CommandArgumentMinValue != 0 {
					minValue = option.NewInt(int(node.Data.CommandArgumentMinValue))
				}

				var maxValue option.Int
				if node.Data.CommandArgumentMaxValue != 0 {
					maxValue = option.NewInt(int(node.Data.CommandArgumentMaxValue))
				}

				var choices []discord.IntegerChoice
				for _, choice := range node.Data.CommandArgumentChoices {
					if choice.Name == "" || choice.Value == "" {
						continue
					}

					value, _ := strconv.ParseInt(choice.Value, 10, 64)

					choices = append(choices, discord.IntegerChoice{
						Name:  choice.Name,
						Value: int(value),
					})
				}

				o = &discord.IntegerOption{
					OptionName:  node.Data.Name,
					Description: node.Data.Description,
					Required:    node.Data.CommandArgumentRequired,
					Min:         minValue,
					Max:         maxValue,
					Choices:     choices,
				}
			case CommandArgumentTypeBoolean:
				o = &discord.BooleanOption{
					OptionName:  node.Data.Name,
					Description: node.Data.Description,
					Required:    node.Data.CommandArgumentRequired,
				}
			case CommandArgumentTypeUser:
				o = &discord.UserOption{
					OptionName:  node.Data.Name,
					Description: node.Data.Description,
					Required:    node.Data.CommandArgumentRequired,
				}
			case CommandArgumentTypeChannel:
				o = &discord.ChannelOption{
					OptionName:  node.Data.Name,
					Description: node.Data.Description,
					Required:    node.Data.CommandArgumentRequired,
				}
			case CommandArgumentTypeRole:
				o = &discord.RoleOption{
					OptionName:  node.Data.Name,
					Description: node.Data.Description,
					Required:    node.Data.CommandArgumentRequired,
				}
			case CommandArgumentTypeMentionable:
				o = &discord.MentionableOption{
					OptionName:  node.Data.Name,
					Description: node.Data.Description,
					Required:    node.Data.CommandArgumentRequired,
				}
			case CommandArgumentTypeNumber:
				var minValue option.Float
				if node.Data.CommandArgumentMinValue != 0 {
					minValue = option.NewFloat(node.Data.CommandArgumentMinValue)
				}

				var maxValue option.Float
				if node.Data.CommandArgumentMaxValue != 0 {
					maxValue = option.NewFloat(node.Data.CommandArgumentMaxValue)
				}

				var choices []discord.NumberChoice
				for _, choice := range node.Data.CommandArgumentChoices {
					if choice.Name == "" || choice.Value == "" {
						continue
					}

					value, _ := strconv.ParseFloat(choice.Value, 64)

					choices = append(choices, discord.NumberChoice{
						Name:  choice.Name,
						Value: value,
					})
				}

				o = &discord.NumberOption{
					OptionName:  node.Data.Name,
					Description: node.Data.Description,
					Required:    node.Data.CommandArgumentRequired,
					Min:         minValue,
					Max:         maxValue,
					Choices:     choices,
				}
			case CommandArgumentTypeAttachment:
				o = &discord.AttachmentOption{
					OptionName:  node.Data.Name,
					Description: node.Data.Description,
					Required:    node.Data.CommandArgumentRequired,
				}
			}

			if o != nil {
				// Required arguments should be the first in the list
				if node.Data.CommandArgumentRequired {
					res = append([]discord.CommandOption{o}, res...)
				} else {
					res = append(res, o)
				}
			}
		}
	}

	return res
}

func (n *CompiledFlowNode) CommandPermissions() *discord.Permissions {
	for _, node := range n.Parents.Default {
		if node.IsCommandPermissions() {
			raw, _ := strconv.ParseInt(node.Data.CommandPermissions, 10, 64)
			res := discord.Permissions(raw)
			return &res

		}
	}

	return nil
}

func (n *CompiledFlowNode) CommandContexts() []discord.InteractionContext {
	// True when disabled
	var guild, botDM, privateChannel bool

	for _, node := range n.Parents.Default {
		if node.IsCommandContexts() {
			for _, ctx := range node.Data.CommandDisabledContexts {
				switch ctx {
				case CommandContextTypeGuild:
					guild = true
				case CommandContextTypeBotDM:
					botDM = true
				case CommandContextTypePrivateChannel:
					privateChannel = true
				}
			}
		}
	}

	res := []discord.InteractionContext{}
	if !guild {
		res = append(res, discord.InteractionContextGuild)
	}
	if !botDM {
		res = append(res, discord.InteractionContextBotDM)
	}
	if !privateChannel {
		res = append(res, discord.InteractionContextPrivateChannel)
	}

	return res
}

func (n *CompiledFlowNode) CommandIntegrations() []discord.ApplicationIntegrationType {
	// True when disabled
	var guild, user bool

	for _, node := range n.Parents.Default {
		if node.IsCommandContexts() {
			for _, integration := range node.Data.CommandDisabledIntegrations {
				switch integration {
				case CommandDisabledIntegrationTypeGuildInstall:
					guild = true
				case CommandDisabledIntegrationTypeUserInstall:
					user = true
				}
			}
		}
	}

	res := []discord.ApplicationIntegrationType{}
	if !guild {
		res = append(res, discord.ApplicationIntegrationTypeGuild)
	}
	if !user {
		res = append(res, discord.ApplicationIntegrationTypeUser)
	}

	return res
}

func (n *CompiledFlowNode) EventListenerType() string {
	if !n.IsEventListenerEntry() {
		return ""
	}
	return n.Data.EventType
}

func (n *CompiledFlowNode) FilterEvent(ctx *FlowContext) (bool, error) {
	if len(n.Parents.Default) == 0 {
		return true, nil
	}

	for _, node := range n.Parents.Default {
		if node.IsEventFilter() {
			var target string

			switch node.Data.EventFilterTarget {
			case EventFilterTypeMessageContent:
				if msg, ok := ctx.Data.Event().(*gateway.MessageCreateEvent); ok {
					target = msg.Content
				}
			case EventFilterTypeUserID:
				target = ctx.Data.UserID().String()
			case EventFilterTypeGuildID:
				target = ctx.Data.GuildID().String()
			case EventFilterTypeChannelID:
				target = ctx.Data.ChannelID().String()
			}

			switch node.Data.EventFilterMode {
			case ComparsionModeEqual:
				if target != node.Data.EventFilterValue {
					return false, nil
				}
			case ComparsionModeNotEqual:
				if target == node.Data.EventFilterValue {
					return false, nil
				}
			case ComparsionModeContains:
				if !strings.Contains(target, node.Data.EventFilterValue) {
					return false, nil
				}
			case ComparsionModeStartsWith:
				if !strings.HasPrefix(target, node.Data.EventFilterValue) {
					return false, nil
				}
			case ComparsionModeEndsWith:
				if !strings.HasSuffix(target, node.Data.EventFilterValue) {
					return false, nil
				}
			}
		}
	}

	return true, nil
}

func (n *CompiledFlowNode) EventDescription() string {
	if !n.IsEventListenerEntry() {
		return ""
	}
	return n.Data.Description
}

func (n *CompiledFlowNode) IsAction() bool {
	return strings.HasPrefix(string(n.Type), "action_")
}

func (n *CompiledFlowNode) FindDirectParentWithType(types ...FlowNodeType) *CompiledFlowNode {
	for _, t := range types {
		for _, node := range n.Parents.Default {
			if node.Type == t {
				return node
			}
		}
	}

	return nil
}

func (n *CompiledFlowNode) FindAllParentsWithType(t FlowNodeType) []*CompiledFlowNode {
	return n.findAllParentsWithType(t, make(map[string]bool))
}

func (n *CompiledFlowNode) findAllParentsWithType(t FlowNodeType, visited map[string]bool) []*CompiledFlowNode {
	res := make([]*CompiledFlowNode, 0)

	// Mark this node as visited to prevent cycles
	visited[n.ID] = true

	for _, node := range n.Parents.Default {
		if node.Type == t {
			res = append(res, node)
		}

		// Only recurse if we haven't visited this node before
		if !visited[node.ID] {
			parents := node.findAllParentsWithType(t, visited)
			res = append(res, parents...)
		}
	}

	return res
}

func (n *CompiledFlowNode) FindDirectChildWithType(types ...FlowNodeType) *CompiledFlowNode {
	for _, t := range types {
		for _, node := range n.Children.Default {
			if node.Type == t {
				return node
			}
		}
	}

	return nil
}

func (n *CompiledFlowNode) FindChildWithType(types ...FlowNodeType) *CompiledFlowNode {
	return n.findChildWithType(make(map[string]bool), types...)
}

func (n *CompiledFlowNode) findChildWithType(visited map[string]bool, types ...FlowNodeType) *CompiledFlowNode {
	// Mark this node as visited to prevent cycles
	visited[n.ID] = true

	// We first want to check all direct children
	for _, node := range n.Children.Default {
		for _, t := range types {
			if node.Type == t {
				return node
			}
		}
	}

	// If no direct children are found, we want to check all children recursively
	for _, node := range n.Children.Default {
		// Only recurse if we haven't visited this node before
		if !visited[node.ID] {
			child := node.findChildWithType(visited, types...)
			if child != nil {
				return child
			}
		}
	}

	return nil
}

func (n *CompiledFlowNode) FindParentWithID(id string) *CompiledFlowNode {
	return n.findParentWithID(id, make(map[string]bool))
}

func (n *CompiledFlowNode) findParentWithID(id string, visited map[string]bool) *CompiledFlowNode {
	// Mark this node as visited to prevent cycles
	visited[n.ID] = true

	for _, node := range n.Parents.Default {
		if node.ID == id {
			return node
		}

		// Only recurse if we haven't visited this node before
		if !visited[node.ID] {
			parent := node.findParentWithID(id, visited)
			if parent != nil {
				return parent
			}
		}
	}

	return nil
}

func (n *CompiledFlowNode) FindChildWithID(nodeID string, includeSubFlows bool) *CompiledFlowNode {
	return n.findChildWithID(nodeID, includeSubFlows, make(map[string]bool))
}

func (n *CompiledFlowNode) findChildWithID(nodeID string, includeSubFlows bool, visited map[string]bool) *CompiledFlowNode {
	if n.ID == nodeID {
		return n
	}

	// Mark this node as visited to prevent cycles
	visited[n.ID] = true

	for _, child := range n.Children.Default {
		// Only recurse if we haven't visited this node before
		if !visited[child.ID] {
			if node := child.findChildWithID(nodeID, includeSubFlows, visited); node != nil {
				return node
			}
		}
	}

	if includeSubFlows {
		for _, children := range n.Children.Handles {
			for _, child := range children {
				// Only recurse if we haven't visited this node before
				if !visited[child.ID] {
					if node := child.findChildWithID(nodeID, includeSubFlows, visited); node != nil {
						return node
					}
				}
			}
		}
	}

	return nil
}
