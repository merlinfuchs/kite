package flow

import (
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/diamondburned/arikawa/v3/discord"
)

type FlowCompiler struct{}

func CompileCommand(data FlowData) (*CompiledFlowNode, error) {
	return compile(data, FlowNodeTypeEntryCommand)
}

func CompileComponentButton(data FlowData) (*CompiledFlowNode, error) {
	return compile(data, FlowNodeTypeEntryComponentButton)
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
			continue
		}

		child, ok := nodeMap[edge.Target]
		if !ok {
			continue
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

func (n *CompiledFlowNode) IsCommandPermissions() bool {
	return n.Type == FlowNodeTypeOptionCommandPermissions
}

func (n *CompiledFlowNode) IsCommandContexts() bool {
	return n.Type == FlowNodeTypeOptionCommandContexts
}

func (n *CompiledFlowNode) CommandData() discord.Command {
	res := discord.Command{
		Name:                     n.CommandName(),
		Options:                  n.CommandArguments(),
		Description:              n.CommandDescription(),
		DefaultMemberPermissions: n.CommandPermissions(),
		NoDMPermission:           slices.Contains(n.CommandDisabledContexts(), CommandContextTypeBotDM),
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
	for _, node := range n.Parents {
		if node.IsCommandArgument() {
			var o discord.CommandOption

			switch node.Data.CommandArgumentType {
			case CommandArgumentTypeString:
				o = &discord.StringOption{
					OptionName:  node.Data.Name,
					Description: node.Data.Description,
					Required:    node.Data.CommandArgumentRequired,
				}
			case CommandArgumentTypeInteger:
				o = &discord.IntegerOption{
					OptionName:  node.Data.Name,
					Description: node.Data.Description,
					Required:    node.Data.CommandArgumentRequired,
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
				o = &discord.NumberOption{
					OptionName:  node.Data.Name,
					Description: node.Data.Description,
					Required:    node.Data.CommandArgumentRequired,
				}
			case CommandArgumentTypeAttachment:
				o = &discord.AttachmentOption{
					OptionName:  node.Data.Name,
					Description: node.Data.Description,
					Required:    node.Data.CommandArgumentRequired,
				}
			}

			if o != nil {
				res = append(res, o)
			}
		}
	}

	return res
}

func (n *CompiledFlowNode) CommandPermissions() *discord.Permissions {
	for _, node := range n.Parents {
		if node.IsCommandPermissions() {
			raw, _ := strconv.ParseInt(node.Data.CommandPermissions, 10, 64)
			res := discord.Permissions(raw)
			return &res

		}
	}

	return nil
}

func (n *CompiledFlowNode) CommandDisabledContexts() []CommandContextType {
	for _, node := range n.Parents {
		if node.IsCommandContexts() {
			return node.Data.CommandDisabledContexts
		}
	}

	return nil
}

func (n *CompiledFlowNode) IsAction() bool {
	return n.Type == FlowNodeTypeActionResponseCreate ||
		n.Type == FlowNodeTypeActionMessageCreate ||
		n.Type == FlowNodeTypeActionLog
}

func (n *CompiledFlowNode) FindDirectParentWithType(types ...FlowNodeType) *CompiledFlowNode {
	for _, t := range types {
		for _, node := range n.Parents {
			if node.Type == t {
				return node
			}
		}
	}

	return nil
}

func (n *CompiledFlowNode) FindAllParentsWithType(t FlowNodeType) []*CompiledFlowNode {
	res := make([]*CompiledFlowNode, 0)

	for _, node := range n.Parents {
		if node.Type == t {
			res = append(res, node)
		}

		parents := node.FindAllParentsWithType(t)
		res = append(res, parents...)
	}

	return res
}

func (n *CompiledFlowNode) FindDirectChildWithType(types ...FlowNodeType) *CompiledFlowNode {
	for _, t := range types {
		for _, node := range n.Children {
			if node.Type == t {
				return node
			}
		}
	}

	return nil
}

func (n *CompiledFlowNode) FindParentWithID(id string) *CompiledFlowNode {
	for _, node := range n.Parents {
		if node.ID == id {
			return node
		}

		parent := node.FindParentWithID(id)
		if parent != nil {
			return parent
		}
	}

	return nil
}
