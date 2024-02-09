use std::collections::HashMap;
use std::rc::Rc;

use super::node::Node;
use super::edge::Edge;

#[derive(Clone)]
pub struct FlowTree {
    entries: Vec<SharedFlowNode>
}

impl FlowTree {
    pub fn new(nodes: &[Node], edges: &[Edge]) -> FlowTree {
        let node_map: HashMap<&str, Node> = nodes.iter()
            .map(|node| (node.id(), node.clone()))
            .collect();

        let edge_source_map: HashMap<&str, Vec<&str>> = edges.iter()
            .fold(HashMap::new(), |mut map, edge| {
                map.entry(&edge.source).or_insert(vec![]).push(&edge.target);
                map
            });

        let edge_target_map: HashMap<&str, Vec<&str>> = edges.iter()
            .fold(HashMap::new(), |mut map, edge| {
                map.entry(&edge.target).or_insert(vec![]).push(&edge.source);
                map
            });

        let mut entries: Vec<SharedFlowNode> = vec![];
        let mut entry_map: HashMap<String, SharedFlowNode> = HashMap::new();

        for node in nodes {
            if matches!(node, Node::EntryCommand { .. } | Node::EntryEvent {.. } | Node::EntryError { .. }) {
                let entry = transform_node(node, &node_map, &edge_source_map, &edge_target_map, &mut entry_map);
                entries.push(entry);
            }
        }

        FlowTree {
            entries
        }
    }
}

fn transform_node(
    node: &Node,
    node_map: &HashMap<&str, Node>, 
    edge_source_map: &HashMap<&str, Vec<&str>>, 
    edge_target_map: &HashMap<&str, Vec<&str>>,
    entry_map: &mut HashMap<String, SharedFlowNode>,
) -> SharedFlowNode {
    // We only want one shared node per node id
    if let Some(entry) = entry_map.get(node.id()) {
        return entry.clone();
    }

    let new_node = match node {
        Node::EntryCommand { id, data } => {
            let options: Vec<SharedFlowNode> = edge_target_map.get(id.as_str())
                .map(|targets| targets
                    .iter()
                    .filter_map(|target| node_map
                        .get(target)
                        .filter(|node| node.is_option())
                        .map(|node| transform_node(node, node_map, edge_source_map, edge_target_map, entry_map)))
                    .collect())
                .unwrap_or(vec![]);

            let next: Vec<SharedFlowNode> = edge_source_map.get(id.as_str())
                .map(|targets| targets 
                    .iter()
                    .filter_map(|target| node_map
                        .get(target)
                        .map(|node| transform_node(node, node_map, edge_source_map, edge_target_map, entry_map)))
                    .collect())
                .unwrap_or(vec![]);

            FlowNode::EntryCommand { 
                name: data.name.clone(),
                description: data.description.clone(), 
                options: options, 
                next: next
            }
        },
        Node::EntryEvent { id, data } => {
            let next: Vec<SharedFlowNode> = edge_source_map.get(id.as_str())
                .map(|targets| targets 
                    .iter()
                    .filter_map(|target| node_map
                        .get(target)
                        .map(|node| transform_node(node, node_map, edge_source_map, edge_target_map, entry_map)))
                    .collect())
                .unwrap_or(vec![]);

            FlowNode::Event { 
                event_type: data.event_type.clone(),
                next: next
            }
        },
        Node::OptionText { data, .. } => {
            FlowNode::CommandOptionText { 
                name: data.name.clone(),
                description: data.description.clone(),
                required: data.required
            }
        }
        Node::EntryError { .. } => unimplemented!()
    };

    let shared_node = Rc::new(new_node);
    entry_map.insert(node.id().to_string(), shared_node.clone());
    return shared_node;
}

#[derive(Clone)]
pub enum FlowNode {
    EntryCommand {
        name: String,
        description: String,
        options: Vec<SharedFlowNode>,
        next: Vec<SharedFlowNode>
    },
    CommandOptionText {
        name: String,
        description: String,
        required: bool
    },
    Event {
        event_type: String,
        next: Vec<SharedFlowNode>
    }
}

pub type SharedFlowNode = Rc<FlowNode>;
