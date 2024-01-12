mod context;
mod sys;

use once_cell::sync::OnceCell;
use quickjs_wasm_rs::JSContextRef;
use std::io::{self, Read};

static mut JS_CONTEXT: OnceCell<JSContextRef> = OnceCell::new();
static mut JS_BYTECODE: OnceCell<Vec<u8>> = OnceCell::new();

static SCRIPT_NAME: &str = "script.js";

/// init() is executed by wizer to create a snapshot after the quickjs context has been initialized.
#[export_name = "wizer.initialize"]
pub extern "C" fn init() {
    let ctx = JSContextRef::default();

    context::set_quickjs_globals(&ctx).unwrap();

    let mut script = String::new();
    io::stdin().read_to_string(&mut script).unwrap();
    let bytecode = ctx
        .compile_global(SCRIPT_NAME, &script)
        .unwrap();

    unsafe {
        JS_BYTECODE.set(bytecode).unwrap();
        JS_CONTEXT.set(ctx).unwrap();
    }
}

#[export_name = "kite_handle"]
pub extern "C" fn handle(length: u32) {
    let context = unsafe { JS_CONTEXT.get().unwrap() };

    let event = sys::get_event(context, length as usize).unwrap();

    let globals = context.global_object().unwrap();

    globals.set_property("__event", event).unwrap();

    context.eval_global(SCRIPT_NAME, "handle()").unwrap();

    let response = globals.get_property("__response").unwrap();

    sys::set_event_response(response).unwrap();
}

fn main() {
    sys::log(0, "Plugin loaded!".to_string());

    let context = unsafe { JS_CONTEXT.get().unwrap() };
    let bytecode = unsafe { JS_BYTECODE.get().unwrap() };

    context.eval_binary(bytecode).unwrap();
}