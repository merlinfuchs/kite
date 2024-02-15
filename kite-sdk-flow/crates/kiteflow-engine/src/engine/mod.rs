use serde::{Serialize, Deserialize};

mod node;
mod tree;
mod edge;

pub use node::Node;
pub use tree::FlowTree;
pub use edge::Edge;

#[derive(Clone, Serialize, Deserialize)]
pub struct FlowData {
    pub nodes: Vec<Node>,
    pub edges: Vec<Edge>,
}