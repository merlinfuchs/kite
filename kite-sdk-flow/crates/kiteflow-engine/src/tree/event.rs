use std::collections::{HashMap, HashSet};

use super::{FlowNode, FlowTree};

#[derive(Clone, Default)]
pub struct EventContext {
    pub depth: usize,
    pub variables: HashMap<String, String>,
}

impl FlowTree {
    pub fn events(&self) -> HashSet<String> {
        self.entries
            .iter()
            .filter_map(|entry| match entry.borrow().to_owned() {
                FlowNode::EntryEvent { event_type, .. } => Some(event_type.clone()),
                FlowNode::EntryCommand { .. } => Some("DISCORD_INTERACTION_CREATE".to_string()),
                _ => None,
            })
            .collect()
    }

    pub fn handle_event(&self, ctx: &mut EventContext, event: &kiters_sys::Event) {
        for entry in &self.entries {
            entry.borrow().handle_event(ctx, event);
        }
    }
}

impl FlowNode {
    pub fn handle_event(&self, ctx: &mut EventContext, event: &kiters_sys::Event) {
        if ctx.depth > 100 {
            panic!("Recursion limit reached")
        }

        ctx.depth += 1;

        match self {
            FlowNode::NoOp => {}
            FlowNode::EntryCommand {
                name,
                description,
                options,
                next,
            } => {
                if event.kind() == "DISCORD_INTERACTION_CREATE" {
                    println!("EntryCommand: {} - {}", name, description);

                    for option in options {
                        option.borrow().handle_event(ctx, event);
                    }
                    for node in next {
                        node.borrow().handle_event(ctx, event);
                    }
                }
            }
            FlowNode::CommandOptionText {
                name,
                description,
                required,
            } => {
                println!(
                    "CommandOptionText: {} - {} - {}",
                    name, description, required
                );
            }
            FlowNode::EntryEvent { event_type, next } => {
                if event_type == event.kind() {
                    println!("EntryEvent: {}", event_type);

                    for node in next {
                        node.borrow().handle_event(ctx, event);
                    }
                }
            }
            FlowNode::EntryError {} => {
                println!("EntryError");
            }
            FlowNode::ActionResponseText { text, next } => {
                println!("ActionResponseText: {}", text);
                for node in next {
                    node.borrow().handle_event(ctx, event);
                }
            }
            FlowNode::ActionLog {
                log_level,
                log_message,
                next,
            } => {
                println!("ActionLog: {} - {}", log_level, log_message);

                for node in next {
                    node.borrow().handle_event(ctx, event);
                }
            }
            FlowNode::ConditionCompare { .. } => {}
        };

        ctx.depth -= 1;
    }
}
