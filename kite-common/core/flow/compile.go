package flow

import (
	"fmt"

	"github.com/diamondburned/arikawa/v3/discord"
)

type FlowCompiler struct{}

func CompileCommand(data FlowData) (*CompiledFlowNode, error) {
	return compile(data, FlowNodeTypeEntryCommand)
}

func compile(data FlowData, entryType FlowNodeType) (*CompiledFlowNode, error) {
	var entryNode *CompiledFlowNode
	nodeMap := make(map[string]*CompiledFlowNode)
	for _, node := range data.Nodes {
		compiledNode := &CompiledFlowNode{
			ID:   node.ID,
			Type: node.Type,
			Data: node.Data,
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
			return nil, nil
		}

		child, ok := nodeMap[edge.Target]
		if !ok {
			return nil, nil
		}

		parent.Children = append(parent.Children, child)
		child.Parents = append(child.Parents, parent)
	}

	return entryNode, nil
}

type CompiledFlowNode struct {
	ID       string
	Type     FlowNodeType
	Data     FlowNodeData
	Parents  []*CompiledFlowNode
	Children []*CompiledFlowNode
}

func (n *CompiledFlowNode) IsEntry() bool {
	return n.Type == FlowNodeTypeEntryCommand || n.Type == FlowNodeTypeEntryEvent

}

func (n *CompiledFlowNode) IsCommandEntry() bool {
	return n.Type == FlowNodeTypeEntryCommand
}

func (n *CompiledFlowNode) IsCommandArgument() bool {
	return n.Type == FlowNodeTypeOptionCommandArgument
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
	for _, node := range n.Parents {
		if node.IsCommandArgument() {
			var o discord.CommandOption

			switch node.Data.CommandArgumentType {
			case CommandArgumentTypeString:
				res = append(res, &discord.StringOption{
					OptionName:  node.Data.Name,
					Description: node.Data.Description,
					Required:    node.Data.CommandArgumentRequired,
				})
			case CommandArgumentTypeInteger:
				res = append(res, &discord.IntegerOption{
					OptionName:  node.Data.Name,
					Description: node.Data.Description,
					Required:    node.Data.CommandArgumentRequired,
				})
			case CommandArgumentTypeBoolean:
				res = append(res, &discord.BooleanOption{
					OptionName:  node.Data.Name,
					Description: node.Data.Description,
					Required:    node.Data.CommandArgumentRequired,
				})
			case CommandArgumentTypeUser:
				res = append(res, &discord.UserOption{
					OptionName:  node.Data.Name,
					Description: node.Data.Description,
					Required:    node.Data.CommandArgumentRequired,
				})
			case CommandArgumentTypeChannel:
				res = append(res, &discord.ChannelOption{
					OptionName:  node.Data.Name,
					Description: node.Data.Description,
					Required:    node.Data.CommandArgumentRequired,
				})
			case CommandArgumentTypeRole:
				res = append(res, &discord.RoleOption{
					OptionName:  node.Data.Name,
					Description: node.Data.Description,
					Required:    node.Data.CommandArgumentRequired,
				})
			case CommandArgumentTypeMentionable:
				res = append(res, &discord.MentionableOption{
					OptionName:  node.Data.Name,
					Description: node.Data.Description,
					Required:    node.Data.CommandArgumentRequired,
				})
			case CommandArgumentTypeNumber:
				res = append(res, &discord.NumberOption{
					OptionName:  node.Data.Name,
					Description: node.Data.Description,
					Required:    node.Data.CommandArgumentRequired,
				})
			case CommandArgumentTypeAttachment:
				res = append(res, &discord.AttachmentOption{
					OptionName:  node.Data.Name,
					Description: node.Data.Description,
					Required:    node.Data.CommandArgumentRequired,
				})
			}

			if o != nil {
				res = append(res, o)
			}
		}
	}

	return res
}

func (n *CompiledFlowNode) IsAction() bool {
	return n.Type == FlowNodeTypeActionResponseCreate ||
		n.Type == FlowNodeTypeActionMessageCreate ||
		n.Type == FlowNodeTypeActionLog
}

func (n *CompiledFlowNode) ParentWithType(t FlowNodeType) *CompiledFlowNode {
	for _, parent := range n.Parents {
		if parent.Type == t {
			return parent
		}
	}
	return nil
}
