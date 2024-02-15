use std::borrow::BorrowMut;
use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;

use super::node::Node;
use super::edge::Edge;

#[derive(Clone)]
pub struct FlowTree {
    pub entries: Vec<SharedFlowNode>
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

     match node {
        Node::EntryCommand { id, data } => {
            let shared_node = Rc::new(RefCell::new(FlowNode::EntryCommand { 
                name: data.name.clone(),
                description: data.description.clone(), 
                options: vec![],
                next: vec![]
            }));
            entry_map.insert(node.id().to_string(), shared_node.clone());

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

            shared_node.as_ref().borrow_mut().set_next(next);
            shared_node
        },
        Node::EntryEvent { id, data } => {
            let shared_node = Rc::new(RefCell::new(FlowNode::EntryEvent { 
                event_type: data.event_type.clone(),
                next: vec![]
            }));
            entry_map.insert(node.id().to_string(), shared_node.clone());

            let next: Vec<SharedFlowNode> = edge_source_map.get(id.as_str())
                .map(|targets| targets 
                    .iter()
                    .filter_map(|target| node_map
                        .get(target)
                        .map(|node| transform_node(node, node_map, edge_source_map, edge_target_map, entry_map)))
                    .collect())
                .unwrap_or(vec![]);

            shared_node.as_ref().borrow_mut().set_next(next);
            shared_node
        },
        Node::OptionText { data, .. } => {
            Rc::new(RefCell::new(FlowNode::CommandOptionText { 
                name: data.name.clone(),
                description: data.description.clone(),
                required: data.required
            }))
        }
        Node::ActionResponseText { id, data } => {
            let shared_node = Rc::new(RefCell::new(FlowNode::ActionResponseText { 
                text: data.text.clone(),
                next: vec![]
            }));
            entry_map.insert(node.id().to_string(), shared_node.clone());

            let next: Vec<SharedFlowNode> = edge_source_map.get(id.as_str())
                .map(|targets| targets 
                    .iter()
                    .filter_map(|target| node_map
                        .get(target)
                        .map(|node| transform_node(node, node_map, edge_source_map, edge_target_map, entry_map)))
                    .collect())
                .unwrap_or(vec![]);

            shared_node.as_ref().borrow_mut().set_next(next);
            shared_node
        }
        Node::ActionLog { id, data } => {
            let shared_node = Rc::new(RefCell::new(FlowNode::ActionLog { 
                log_level: data.log_level.clone(),
                log_message: data.log_message.clone(),
                next: vec![]
            }));
            entry_map.insert(node.id().to_string(), shared_node.clone());

            let next: Vec<SharedFlowNode> = edge_source_map.get(id.as_str())
                .map(|targets| targets 
                    .iter()
                    .filter_map(|target| node_map
                        .get(target)
                        .map(|node| transform_node(node, node_map, edge_source_map, edge_target_map, entry_map)))
                    .collect())
                .unwrap_or(vec![]);

            // TODO: temp.borrow_mut() = next;

            shared_node.as_ref().borrow_mut().set_next(next);
            shared_node
        }
        Node::EntryError { .. } => Rc::new(RefCell::new(FlowNode::EntryError {}))
    }
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
    EntryEvent {
        event_type: String,
        next: Vec<SharedFlowNode>
    },
    EntryError {},
    ActionResponseText {
        text: String,
        next: Vec<SharedFlowNode>
    },
    ActionLog {
        log_level: String,
        log_message: String,
        next: Vec<SharedFlowNode>
    }
}

impl FlowNode {
    pub fn set_next(&mut self, new_next: Vec<SharedFlowNode>) {
        match self {
            FlowNode::EntryCommand { next, .. } => {
                *next = new_next;
            },
            FlowNode::EntryEvent { next, .. } => {
                *next = new_next;
            },
            FlowNode::ActionResponseText { next, .. } => {
                *next = new_next;
            },
            FlowNode::ActionLog { next, .. } => {
                *next = new_next;
            },
            _ => {}
        }
    }

    pub fn walk(&self) {
        match self {
            FlowNode::EntryCommand { name, description, options, next } => {
                println!("EntryCommand: {} - {}", name, description);
                for option in options {
                    option.borrow().walk();
                }
                for node in next {
                    node.borrow().walk();
                }
            },
            FlowNode::CommandOptionText { name, description, required } => {
                println!("CommandOptionText: {} - {} - {}", name, description, required);
            },
            FlowNode::EntryEvent { event_type, next } => {
                println!("EntryEvent: {}", event_type);
                for node in next {
                    node.borrow().walk();
                }
            },
            FlowNode::EntryError {} => {
                println!("EntryError");
            },
            FlowNode::ActionResponseText { text, next } => {
                println!("ActionResponseText: {}", text);
                for node in next {
                    node.borrow().walk();
                }
            },
            FlowNode::ActionLog { log_level, log_message, next } => {
                println!("ActionLog: {} - {}", log_level, log_message);
                for node in next {
                    node.borrow().walk();
                }
            }
        }
    }
}

pub type SharedFlowNode = Rc<RefCell<FlowNode>>;
