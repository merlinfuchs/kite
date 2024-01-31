mod context;
mod sys;

use once_cell::sync::OnceCell;
use quickjs_wasm_rs::JSContextRef;
use std::io::{self, Read};

static mut JS_CONTEXT: OnceCell<JSContextRef> = OnceCell::new();

static SCRIPT_NAME: &str = "script.js";

fn main() {
    let ctx = JSContextRef::default();

    context::set_quickjs_globals(&ctx)
        .expect("Failed to set quickjs globals");

    let mut script = String::new();
    io::stdin().read_to_string(&mut script)
        .expect("Failed to read script from stdin");

    // We compile and eval the script ahead of time to improve performance and be able to extract the manifest.
    // This does significantly increase the size of the WASM module because the bytecode can be pretty big.
    // If the script was compiled on the host instead we would still end up with the same memory usage, so it's still better to do it here.
    let bytecode = ctx
        .compile_global(SCRIPT_NAME, &script)
        .expect("Failed to compile script");

    ctx.eval_binary(&bytecode)
        .expect("Failed to execute script");

    unsafe {
        JS_CONTEXT.set(ctx).unwrap();
    }
}

#[export_name = "kite_describe"]
pub extern "C" fn describe() {
    let context = unsafe { JS_CONTEXT.get().unwrap() };

    let globals = context.global_object().unwrap();
    let kite_object = globals.get_property("Kite").unwrap();

    let handle_func = kite_object.get_property("describe")
        .expect("Failed to get script Kite.describe() function");

    let response = handle_func.call(&kite_object, &[])
        .expect("Failed to execute script Kite.describe() function");

    sys::set_manifest(response)
        .expect("Failed to set manifest on host");
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

#[export_name = "kite_get_api_encoding"]
pub extern "C" fn get_api_encoding() -> u32 {
    return 0;
}
