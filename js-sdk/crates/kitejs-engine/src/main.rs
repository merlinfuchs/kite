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

    context::set_quickjs_globals(&ctx)
        .expect("Failed to set quickjs globals");

    let mut script = String::new();
    io::stdin().read_to_string(&mut script)
        .expect("Failed to read script from stdin");

    let bytecode = ctx
        .compile_global(SCRIPT_NAME, &script)
        .expect("Failed to compile script");

    unsafe {
        JS_BYTECODE.set(bytecode).unwrap();
        JS_CONTEXT.set(ctx).unwrap();
    }
}

#[export_name = "kite_handle"]
pub extern "C" fn handle(length: u32) {
    let context = unsafe { JS_CONTEXT.get().unwrap() };

    let event = sys::get_event(context, length as usize)
        .expect("Failed to get event from host");

    let globals = context.global_object().unwrap();
    let kite_object = globals.get_property("Kite").unwrap();

    let handle_func = kite_object.get_property("handle")
        .expect("Failed to get script Kite.handle() function");

    let response = handle_func.call(&kite_object, &[event])
        .expect("Failed to execute script Kite.handle() function");

    sys::set_event_response(response)
        .expect("Failed to set event response on host");
}

#[export_name = "kite_get_api_version"]
pub extern "C" fn get_api_version() -> u32 {
    return 0;
}


fn main() {
    sys::log(0, "Plugin loaded!".to_string());

    let context = unsafe { JS_CONTEXT.get().unwrap() };
    let bytecode = unsafe { JS_BYTECODE.get().unwrap() };

    context.eval_binary(bytecode)
        .expect("Failed to execute script");
}