use quickjs_wasm_rs::{JSContextRef, JSValue, JSValueRef};

use crate::sys;

/// set quickjs globals
pub fn set_quickjs_globals(context: &JSContextRef) -> anyhow::Result<()> {
    set_quickjs_console(context)?;
    set_quickjs_kite(context)?;

    Ok(())
}

pub fn set_quickjs_kite(context: &JSContextRef) -> anyhow::Result<()> {
    let kite_object = context.object_value()?;

    let kite_call_callback = context.wrap_callback(|ctx, this, args| {
        let call = args[0];

        let resp = sys::call(ctx, call)?;

        // TODO: Ok(JSValue::from(resp))
        this.set_property("callResponse", resp)?;
        Ok(JSValue::Undefined)
    })?;
    kite_object.set_property("call", kite_call_callback)?;

    let kite_config_callback = context.wrap_callback(|ctx, this, _args| {
        let config = sys::get_config(ctx)?;

        // TODO: Ok(JSValue::from(resp))
        this.set_property("config", config)?;
        Ok(JSValue::Undefined)
    })?;
    kite_object.set_property("getConfig", kite_config_callback)?;

    let global = context.global_object()?;
    global.set_property("Kite", kite_object)?;

    Ok(())
}

pub fn set_quickjs_console(context: &JSContextRef) -> anyhow::Result<()> {
    let console_object = context.object_value()?;

    let console_debug_callback = context.wrap_callback(|_ctx, _this, args| {
        sys::log(0, args_to_string(args));
        Ok(JSValue::Undefined)
    })?;
    console_object.set_property("debug", console_debug_callback)?;

    let console_log_callback = context.wrap_callback(|_ctx, _this, args| {
        sys::log(1, args_to_string(args));
        Ok(JSValue::Undefined)
    })?;
    console_object.set_property("log", console_log_callback)?;

    let console_info_callback = context.wrap_callback(|_ctx, _this, args| {
        sys::log(1, args_to_string(args));
        Ok(JSValue::Undefined)
    })?;
    console_object.set_property("info", console_info_callback)?;

    let console_warn_callback = context.wrap_callback(|_ctx, _this, args| {
        sys::log(2, args_to_string(args));
        Ok(JSValue::Undefined)
    })?;
    console_object.set_property("warn", console_warn_callback)?;

    let console_error_callback = context.wrap_callback(|_ctx, _this, args| {
        sys::log(3, args_to_string(args));
        Ok(JSValue::Undefined)
    })?;
    console_object.set_property("error", console_error_callback)?;

    let global = context.global_object()?;
    global.set_property("console", console_object)?;

    Ok(())
}

fn args_to_string(args: &[JSValueRef]) -> String {
    let mut log_line = String::new();
    for (i, arg) in args.iter().enumerate() {
        if i != 0 {
            log_line.push(' ');
        }
        let line = arg.to_string();
        log_line.push_str(&line);
    }
    log_line
}
