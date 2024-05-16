use std::{cell::RefCell, collections::HashMap};

pub use crate::Event;
use crate::{manifest::Manifest, EventResponse, ModuleError};

#[link(wasm_import_module = "env")]
extern "C" {
    pub fn kite_set_manifest(offset: u32, length: u32) -> u32;
    pub fn kite_get_config_size() -> u32;
    pub fn kite_get_config(offset: u32) -> u32;
    pub fn kite_log(level: u32, offset: u32, length: u32) -> u32;
    pub fn kite_call(offset: u32, length: u32) -> u32;
    pub fn kite_get_event(offset: u32) -> u32;
    pub fn kite_set_event_response(offset: u32, length: u32) -> u32;
    pub fn kite_get_call_response(offset: u32) -> u32;
}

#[no_mangle]
pub extern "C" fn kite_get_api_version() -> u32 {
    0
}

#[no_mangle]
pub extern "C" fn kite_get_api_encoding() -> u32 {
    0
}

pub type EventHandlerFn = dyn Fn(&Event) -> Result<(), ModuleError>;

thread_local! {
    pub static MANIFEST: RefCell<Manifest> = RefCell::new(Manifest::default());
    pub static EVENT_HANDLERS: RefCell<HashMap<String, Vec<Box<EventHandlerFn>>>> =  RefCell::new(HashMap::new());
}

#[no_mangle]
pub extern "C" fn kite_describe() {
    MANIFEST.with(|manifest| {
        let json = serde_json::to_string(manifest).expect("Failed to serialize manifest");

        let ok = unsafe { kite_set_manifest(json.as_ptr() as u32, json.len() as u32) };
        if ok != 0 {
            panic!("Failed to set manifest");
        }
    });
}

#[no_mangle]
pub extern "C" fn kite_handle(length: u32) {
    let buf = vec![0; length as usize];

    let ok = unsafe { kite_get_event(buf.as_ptr() as u32) };
    if ok != 0 {
        return;
    }

    let event: Event = match serde_json::from_slice(buf.as_slice()) {
        Ok(e) => e,
        Err(err) => {
            set_event_error_message(format!("Failed to parse event: {}", err));
            return;
        }
    };

    EVENT_HANDLERS.with_borrow(|handlers| match handlers.get(event.kind()) {
        Some(handlers) => {
            if handlers.len() == 0 {
                set_event_error_message(format!("No handlers for event kind: {}", event.kind()));
                return;
            }

            for handler in handlers {
                match handler(&event) {
                    Ok(_) => set_event_response(EventResponse {
                        success: true,
                        error: None,
                    }),
                    Err(err) => set_event_response(EventResponse {
                        success: false,
                        error: Some(err),
                    }),
                }
            }
        }
        None => set_event_error_message(format!("No handlers for event kind: {}", event.kind())),
    });
}

fn set_event_error_message(message: String) {
    let resp = EventResponse {
        success: false,
        error: Some(ModuleError {
            code: "unknown".to_string(),
            message,
        }),
    };

    set_event_response(resp);
}

fn set_event_response(resp: EventResponse) {
    let json = serde_json::to_string(&resp).expect("Failed to serialize event response");

    let ok = unsafe { kite_set_event_response(json.as_ptr() as u32, json.len() as u32) };
    if ok != 0 {
        panic!("Failed to set event response");
    }
}
