use std::cell::RefCell;
use std::rc::Rc;

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
        conditions: Vec<SharedFlowNode>,
        otherwise: Option<SharedFlowNode>,
    },
}

impl Default for FlowNode {
    fn default() -> Self {
        FlowNode::NoOp
    }
}

pub type SharedFlowNode = Rc<RefCell<FlowNode>>;
