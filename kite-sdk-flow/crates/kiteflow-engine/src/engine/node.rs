use serde::{Serialize, Deserialize};

#[derive(Clone, Serialize, Deserialize)]
#[serde(tag = "type", rename_all = "snake_case")]
pub enum Node {
    EntryCommand { id: String, data: NodeDataEntryCommand },
    EntryEvent { id: String, data: NodeDataEntryEvent },
    EntryError { id: String, data: NodeDataEntryError },
    OptionText { id: String, data: NodeDataOptionText },
}

impl Node {
    pub fn id(&self) -> &str {
        match self {
            Node::EntryCommand { id, .. } => id,
            Node::EntryEvent { id, .. } => id,
            Node::EntryError { id, .. } => id,
            Node::OptionText { id, .. } => id,
        }
    }

    pub fn is_entry(&self) -> bool {
        matches!(self, Node::EntryCommand { .. } | Node::EntryEvent { .. } | Node::EntryError { .. })
    }

    pub fn is_action(&self) -> bool {
        return false;
    }

    pub fn is_condition(&self) -> bool {
        return false;
    }

    pub fn is_option(&self) -> bool {
        return false;
    }
}

#[derive(Clone, Serialize, Deserialize)]
pub struct NodeDataEntryCommand {
    pub name: String,
    pub description: String
}

#[derive(Clone, Serialize, Deserialize)]
pub struct NodeDataEntryEvent {
    pub event_type: String
}

#[derive(Clone, Serialize, Deserialize)]
pub struct NodeDataEntryError {
    pub log_level: String,
    pub log_message: String
}

#[derive(Clone, Serialize, Deserialize)]
pub struct NodeDataOptionText {
    pub name: String,
    pub description: String,
    pub required: bool
}