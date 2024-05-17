use std::cell::RefCell;
use std::rc::Rc;

use crate::data::ConditionItemCompareMode;

#[derive(Clone)]
pub struct FlowTree {
    pub entries: Vec<SharedFlowNode>,
}

#[derive(Clone)]
pub enum FlowNode {
    NoOp,
    EntryCommand {
        name: String,
        description: String,
        options: Vec<SharedFlowNode>,
        next: Vec<SharedFlowNode>,
    },
    CommandOptionText {
        name: String,
        description: String,
        required: bool,
    },
    EntryEvent {
        event_type: String,
        next: Vec<SharedFlowNode>,
    },
    EntryError {},
    ActionResponseText {
        text: String,
        next: Vec<SharedFlowNode>,
    },
    ActionLog {
        log_level: String,
        log_message: String,
        next: Vec<SharedFlowNode>,
    },
    ConditionCompare {
        items: Vec<SharedFlowNode>,
        base_value: String,
        allow_multiple: bool,
    },
    ConditionItemCompare {
        mode: ConditionItemCompareMode,
        value: String,
        next: Vec<SharedFlowNode>,
    },
    ConditionItemElse {
        next: Vec<SharedFlowNode>,
    },
}

impl FlowNode {
    pub fn is_condition_item(&self) -> bool {
        match self {
            FlowNode::ConditionItemCompare { .. } => true,
            _ => false,
        }
    }

    pub fn is_condition_else(&self) -> bool {
        match self {
            FlowNode::ConditionItemElse { .. } => true,
            _ => false,
        }
    }
}

impl Default for FlowNode {
    fn default() -> Self {
        FlowNode::NoOp
    }
}

pub type SharedFlowNode = Rc<RefCell<FlowNode>>;
