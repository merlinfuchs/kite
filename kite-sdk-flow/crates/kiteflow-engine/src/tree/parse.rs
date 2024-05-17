use std::cell::RefCell;
use std::collections::HashMap;
use std::rc::Rc;

use crate::data::Edge;
use crate::data::Node;

use super::tree::FlowTree;
use super::{FlowNode, SharedFlowNode};

impl FlowTree {
    pub fn new(nodes: &[Node], edges: &[Edge]) -> FlowTree {
        let node_map: HashMap<&str, Node> =
            nodes.iter().map(|node| (node.id(), node.clone())).collect();

        let edge_source_map: HashMap<&str, Vec<&str>> =
            edges.iter().fold(HashMap::new(), |mut map, edge| {
                map.entry(&edge.source).or_insert(vec![]).push(&edge.target);
                map
            });

        let edge_target_map: HashMap<&str, Vec<&str>> =
            edges.iter().fold(HashMap::new(), |mut map, edge| {
                map.entry(&edge.target).or_insert(vec![]).push(&edge.source);
                map
            });

        let mut entries: Vec<SharedFlowNode> = vec![];
        let mut entry_map: HashMap<String, SharedFlowNode> = HashMap::new();

        for node in nodes {
            if matches!(
                node,
                Node::EntryCommand { .. } | Node::EntryEvent { .. } | Node::EntryError { .. }
            ) {
                let entry = transform_node(
                    node,
                    &node_map,
                    &edge_source_map,
                    &edge_target_map,
                    &mut entry_map,
                );
                entries.push(entry);
            }
        }

        FlowTree { entries }
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
            let shared_node = SharedFlowNode::default();
            entry_map.insert(node.id().to_string(), shared_node.clone());

            let options: Vec<SharedFlowNode> = edge_target_map
                .get(id.as_str())
                .map(|targets| {
                    targets
                        .iter()
                        .filter_map(|target| {
                            node_map
                                .get(target)
                                .filter(|node| node.is_option())
                                .map(|node| {
                                    transform_node(
                                        node,
                                        node_map,
                                        edge_source_map,
                                        edge_target_map,
                                        entry_map,
                                    )
                                })
                        })
                        .collect()
                })
                .unwrap_or(vec![]);

            let next: Vec<SharedFlowNode> = edge_source_map
                .get(id.as_str())
                .map(|targets| {
                    targets
                        .iter()
                        .filter_map(|target| {
                            node_map.get(target).map(|node| {
                                transform_node(
                                    node,
                                    node_map,
                                    edge_source_map,
                                    edge_target_map,
                                    entry_map,
                                )
                            })
                        })
                        .collect()
                })
                .unwrap_or(vec![]);

            *shared_node.borrow_mut() = FlowNode::EntryCommand {
                name: data.name.clone(),
                description: data.description.clone(),
                options,
                next,
            };
            shared_node
        }
        Node::EntryEvent { id, data } => {
            let shared_node = SharedFlowNode::default();
            entry_map.insert(node.id().to_string(), shared_node.clone());

            let next: Vec<SharedFlowNode> = edge_source_map
                .get(id.as_str())
                .map(|targets| {
                    targets
                        .iter()
                        .filter_map(|target| {
                            node_map.get(target).map(|node| {
                                transform_node(
                                    node,
                                    node_map,
                                    edge_source_map,
                                    edge_target_map,
                                    entry_map,
                                )
                            })
                        })
                        .collect()
                })
                .unwrap_or(vec![]);

            *shared_node.borrow_mut() = FlowNode::EntryEvent {
                event_type: data.event_type.clone(),
                next,
            };
            shared_node
        }
        Node::OptionText { data, .. } => Rc::new(RefCell::new(FlowNode::CommandOptionText {
            name: data.name.clone(),
            description: data.description.clone(),
            required: data.required,
        })),
        Node::ActionResponseText { id, data } => {
            let shared_node = SharedFlowNode::default();
            entry_map.insert(node.id().to_string(), shared_node.clone());

            let next: Vec<SharedFlowNode> = edge_source_map
                .get(id.as_str())
                .map(|targets| {
                    targets
                        .iter()
                        .filter_map(|target| {
                            node_map.get(target).map(|node| {
                                transform_node(
                                    node,
                                    node_map,
                                    edge_source_map,
                                    edge_target_map,
                                    entry_map,
                                )
                            })
                        })
                        .collect()
                })
                .unwrap_or(vec![]);

            *shared_node.borrow_mut() = FlowNode::ActionResponseText {
                text: data.text.clone(),
                next,
            };
            shared_node
        }
        Node::ActionLog { id, data } => {
            let shared_node = SharedFlowNode::default();
            entry_map.insert(node.id().to_string(), shared_node.clone());

            let next: Vec<SharedFlowNode> = edge_source_map
                .get(id.as_str())
                .map(|targets| {
                    targets
                        .iter()
                        .filter_map(|target| {
                            node_map.get(target).map(|node| {
                                transform_node(
                                    node,
                                    node_map,
                                    edge_source_map,
                                    edge_target_map,
                                    entry_map,
                                )
                            })
                        })
                        .collect()
                })
                .unwrap_or(vec![]);

            *shared_node.borrow_mut() = FlowNode::ActionLog {
                log_level: data.log_level.clone(),
                log_message: data.log_message.clone(),
                next,
            };
            shared_node
        }
        Node::EntryError { .. } => Rc::new(RefCell::new(FlowNode::EntryError {})),
        Node::ConditionCompare { id, data } => {
            let shared_node = SharedFlowNode::default();
            entry_map.insert(node.id().to_string(), shared_node.clone());

            let items: Vec<SharedFlowNode> = edge_target_map
                .get(id.as_str())
                .map(|targets| {
                    targets
                        .iter()
                        .filter_map(|target| {
                            node_map.get(target).map(|node| {
                                transform_node(
                                    node,
                                    node_map,
                                    edge_source_map,
                                    edge_target_map,
                                    entry_map,
                                )
                            })
                        })
                        .collect()
                })
                .unwrap_or(vec![]);

            *shared_node.borrow_mut() = FlowNode::ConditionCompare {
                items,
                base_value: data.condition_base_value.clone(),
                allow_multiple: data.condition_allow_multiple,
            };
            shared_node
        }
        Node::ConditionItemCompare { id, data } => {
            let shared_node = SharedFlowNode::default();
            entry_map.insert(node.id().to_string(), shared_node.clone());

            let next: Vec<SharedFlowNode> = edge_source_map
                .get(id.as_str())
                .map(|targets| {
                    targets
                        .iter()
                        .filter_map(|target| {
                            node_map.get(target).map(|node| {
                                transform_node(
                                    node,
                                    node_map,
                                    edge_source_map,
                                    edge_target_map,
                                    entry_map,
                                )
                            })
                        })
                        .collect()
                })
                .unwrap_or(vec![]);

            *shared_node.borrow_mut() = FlowNode::ConditionItemCompare {
                mode: data.condition_item_mode.clone(),
                value: data.condition_item_value.clone(),
                next,
            };
            shared_node
        }
        Node::ConditionItemElse { id } => {
            let shared_node = SharedFlowNode::default();
            entry_map.insert(node.id().to_string(), shared_node.clone());

            let next: Vec<SharedFlowNode> = edge_source_map
                .get(id.as_str())
                .map(|targets| {
                    targets
                        .iter()
                        .filter_map(|target| {
                            node_map.get(target).map(|node| {
                                transform_node(
                                    node,
                                    node_map,
                                    edge_source_map,
                                    edge_target_map,
                                    entry_map,
                                )
                            })
                        })
                        .collect()
                })
                .unwrap_or(vec![]);

            *shared_node.borrow_mut() = FlowNode::ConditionItemElse { next };
            shared_node
        }
    }
}
