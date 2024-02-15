use std::io::{self, Read};

use anyhow::Result;
use engine::{FlowData, FlowTree};

mod engine;

fn main() -> Result<()> {
    let mut flow_data = String::new();
    io::stdin().read_to_string(&mut flow_data)
        .expect("Failed to read flow data from stdin");

    let flow: FlowData = serde_json::from_str(&flow_data)
        .expect("Failed to parse flow data");

    let tree = FlowTree::new(&flow.nodes, &flow.edges);
    
    println!("Parsed: {}", tree.entries.len());

    for entry in tree.entries {
        entry.borrow().walk();
    }

    Ok(())
}