use std::cell::RefCell;
use std::io::{self, Read};

use anyhow::Result;
use data::FlowData;
use kiters_sys::{EventData, ModuleError};
use tree::{EventContext, FlowTree};

mod data;
mod tree;

thread_local! {
    pub static TREE: RefCell<FlowTree> = panic!("Flow tree not initialized");
}

fn main() -> Result<()> {
    let mut flow_data = String::new();
    io::stdin()
        .read_to_string(&mut flow_data)
        .expect("Failed to read flow data from stdin");

    let flow: FlowData = serde_json::from_str(&flow_data).expect("Failed to parse flow data");

    let tree = FlowTree::new(&flow.nodes, &flow.edges);

    println!("Parsed entries: {}", tree.entries.len());

    for event_type in tree.events() {
        kiters_sys::add_event_handler(&event_type, handle_event);
    }

    TREE.set(tree);

    handle_event(&kiters_sys::Event {
        app_id: "".to_string(),
        guild_id: "".to_string(),
        data: EventData::Initiate,
    })
    .expect("Failed to handle initiate event");

    Ok(())
}

pub fn handle_event(event: &kiters_sys::Event) -> Result<(), ModuleError> {
    // kiters_sys::make_call(CallData::Sleep { duration: 100 }, None)?;

    TREE.with_borrow(|tree: &FlowTree| {
        let mut ctx = EventContext::default();

        tree.handle_event(&mut ctx, event);
        Ok(())
    })
}
