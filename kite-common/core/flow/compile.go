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

func (n *CompiledFlowNode) IsCommandOption() bool {
	return n.Type == FlowNodeTypeOptionCommandText ||
		n.Type == FlowNodeTypeOptionCommandNumber ||
		n.Type == FlowNodeTypeOptionCommandUser ||
		n.Type == FlowNodeTypeOptionCommandChannel ||
		n.Type == FlowNodeTypeOptionCommandRole ||
		n.Type == FlowNodeTypeOptionCommandAttachment
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

func (n *CompiledFlowNode) CommandOptions() discord.CommandOptions {
	res := make(discord.CommandOptions, 0)
	for _, node := range n.Parents {
		if node.IsCommandOption() {
			// TODO: Implement other option types
			res = append(res, &discord.StringOption{
				OptionName:  node.CommandOptionName(),
				Description: node.CommandOptionDescription(),
				Required:    true,
			})
		}
	}

	return res
}

func (n *CompiledFlowNode) CommandOptionName() string {
	if !n.IsCommandOption() {
		return ""
	}
	return n.Data.Name
}

func (n *CompiledFlowNode) CommandOptionDescription() string {
	if !n.IsCommandOption() {
		return ""
	}
	return n.Data.Description
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
