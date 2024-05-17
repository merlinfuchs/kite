use serde::{Deserialize, Serialize};

mod edge;
mod node;

pub use edge::*;
pub use node::*;

#[derive(Clone, Serialize, Deserialize)]
pub struct FlowData {
    pub nodes: Vec<Node>,
    pub edges: Vec<Edge>,
}
