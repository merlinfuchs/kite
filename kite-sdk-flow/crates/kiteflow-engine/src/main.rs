use std::io::{self, Read};

use anyhow::Result;
use engine::{FlowData, FlowTree};

mod engine;
mod sys;

fn main() -> Result<()> {
    let mut flow_data = String::new();
    io::stdin()
        .read_to_string(&mut flow_data)
        .expect("Failed to read flow data from stdin");

    let flow: FlowData = serde_json::from_str(&flow_data).expect("Failed to parse flow data");

    let tree = FlowTree::new(&flow.nodes, &flow.edges);

    println!("Parsed: {}", tree.entries.len());

    for entry in tree.entries {
        entry.borrow().walk();
    }

    Ok(())
}

#[export_name = "kite_describe"]
pub extern "C" fn describe() {}

#[export_name = "kite_handle"]
pub extern "C" fn handle(length: u32) {}

#[export_name = "kite_get_api_version"]
pub extern "C" fn get_api_version() -> u32 {
    return 0;
}

#[export_name = "kite_get_api_encoding"]
pub extern "C" fn get_api_encoding() -> u32 {
    return 0;
}
