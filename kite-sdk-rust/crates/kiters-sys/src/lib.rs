mod call;
mod error;
mod event;
pub mod manifest;
mod sys;

use std::collections::HashMap;

pub use call::*;
pub use error::*;
pub use event::*;

pub fn add_event_handler(
    kind: &str,
    handler: impl Fn(&Event) -> Result<(), ModuleError> + 'static,
) {
    sys::EVENT_HANDLERS.with_borrow_mut(|event_handlers| match event_handlers.get_mut(kind) {
        Some(handlers) => {
            handlers.push(Box::new(handler));
        }
        None => {
            let mut handlers: Vec<Box<sys::EventHandlerFn>> = Vec::new();
            handlers.push(Box::new(handler));
            event_handlers.insert(kind.to_string(), handlers);
        }
    });

    sys::MANIFEST.with_borrow_mut(|manifest| {
        manifest.events.push(kind.to_string());
    });
}

pub fn make_call(data: CallData, config: Option<CallConfig>) -> Result<CallResponse, HostError> {
    let call = Call {
        data,
        config: config.unwrap_or_default(),
    };

    let json = serde_json::to_string(&call).expect("Failed to serialize call");

    let length = unsafe { sys::kite_call(json.as_ptr() as u32, json.len() as u32) };
    if length == 0 {
        return Err(HostError::new("no_response", "No call response from host"));
    }

    let buf = vec![0; length as usize];
    let ok = unsafe { sys::kite_get_call_response(buf.as_ptr() as u32) };
    if ok != 0 {
        return Err(HostError::new(
            "no_response",
            "Failed to get call response from host",
        ));
    }

    let resp: CallResponse = match serde_json::from_slice(buf.as_slice()) {
        Ok(r) => r,
        Err(err) => {
            return Err(HostError::new(
                "invalid_response",
                format!("Failed to parse call response: {}", err),
            ));
        }
    };

    if !resp.success {
        return match resp.error {
            Some(err) => Err(err),
            None => Err(HostError::new("no_error", "No error message from host")),
        };
    }

    Ok(resp)
}

pub fn log<T: ToString>(level: u32, msg: T) {
    let msg = msg.to_string();
    unsafe { sys::kite_log(level, msg.as_ptr() as u32, msg.len() as u32) };
}

pub fn get_config() -> HashMap<String, String> {
    let size = unsafe { sys::kite_get_config_size() };
    let buf = vec![0; size as usize];

    let ok = unsafe { sys::kite_get_config(buf.as_ptr() as u32) };
    if ok != 0 {
        panic!("Failed to get config from host");
    }

    match serde_json::from_slice(buf.as_slice()) {
        Ok(config) => config,
        Err(_) => panic!("Failed to parse config from host"),
    }
}
