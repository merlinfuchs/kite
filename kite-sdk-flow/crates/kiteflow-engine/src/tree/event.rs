use std::collections::{HashMap, HashSet};

use crate::data::ConditionItemCompareMode;

use super::{FlowNode, FlowTree};

#[derive(Clone)]
pub struct EventContext<'a> {
    pub event: &'a kiters_sys::Event,
    pub depth: usize,
    pub variables: HashMap<String, String>,
}

impl<'a> EventContext<'a> {
    pub fn new(event: &'a kiters_sys::Event) -> Self {
        Self {
            event,
            depth: 0,
            variables: HashMap::new(),
        }
    }
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

    pub fn handle_event(&self, ctx: &mut EventContext) {
        for entry in &self.entries {
            entry.borrow().handle_event(ctx);
        }
    }
}

impl FlowNode {
    pub fn handle_event(&self, ctx: &mut EventContext) {
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
                if ctx.event.kind() == "DISCORD_INTERACTION_CREATE" {
                    println!("EntryCommand: {} - {}", name, description);

                    for option in options {
                        option.borrow().handle_event(ctx);
                    }
                    for node in next {
                        node.borrow().handle_event(ctx);
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
                if event_type == ctx.event.kind() {
                    println!("EntryEvent: {}", event_type);

                    for node in next {
                        node.borrow().handle_event(ctx);
                    }
                }
            }
            FlowNode::EntryError {} => {
                println!("EntryError");
            }
            FlowNode::ActionResponseText { text, next } => {
                println!("ActionResponseText: {}", text);
                for node in next {
                    node.borrow().handle_event(ctx);
                }
            }
            FlowNode::ActionLog {
                log_level,
                log_message,
                next,
            } => {
                println!("ActionLog: {} - {}", log_level, log_message);

                for node in next {
                    node.borrow().handle_event(ctx);
                }
            }
            FlowNode::ConditionCompare {
                items,
                base_value,
                allow_multiple,
            } => {
                let mut else_node = None;
                let mut handled = false;

                for item in items {
                    let i = item.borrow();

                    if i.is_condition_item() {
                        if i.evalute_condition(ctx, base_value) {
                            handled = true;
                            i.handle_event(ctx);
                            if !allow_multiple {
                                break;
                            }
                        }
                    } else if i.is_condition_else() {
                        else_node = Some(i);
                    }
                }

                if !handled {
                    if let Some(else_node) = else_node {
                        else_node.handle_event(ctx);
                    }
                }
            }
            FlowNode::ConditionItemCompare { next, .. } => {
                for node in next {
                    node.borrow().handle_event(ctx);
                }
            }
            FlowNode::ConditionItemElse { next, .. } => {
                for node in next {
                    node.borrow().handle_event(ctx);
                }
            }
        };

        ctx.depth -= 1;
    }

    pub fn evalute_condition(&self, ctx: &mut EventContext, base_value: &str) -> bool {
        // TODO: access variables

        match self {
            FlowNode::ConditionItemCompare { mode, value, next } => match mode {
                ConditionItemCompareMode::Equal => value == base_value,
                ConditionItemCompareMode::NotEqual => value != base_value,
                _ => todo!(),
            },
            _ => true,
        }
    }
}
