use anyhow::Result;
use quickjs_wasm_rs::{Deserializer, JSContextRef, JSValueRef, Serializer};

#[link(wasm_import_module = "env")]
extern "C" {
    fn kite_set_manifest(offset: u32, length: u32) -> u32;
    fn kite_get_config_size() -> u32;
    fn kite_get_config(offset: u32) -> u32;
    fn kite_log(level: u32, offset: u32, length: u32) -> u32;
    fn kite_call(offset: u32, length: u32) -> u32;
    fn kite_get_event(offset: u32) -> u32;
    fn kite_set_event_response(offset: u32, length: u32) -> u32;
    fn kite_get_call_response(offset: u32) -> u32;
}

/// Transcodes a byte slice containing a JSON encoded payload into a [`JSValueRef`].
///
/// Arguments:
/// * `context` - A reference to the [`JSContextRef`] that will contain the
///   returned [`JSValueRef`].
/// * `bytes` - A byte slice containing a JSON encoded payload.
pub fn transcode_input<'a>(context: &'a JSContextRef, bytes: &[u8]) -> Result<JSValueRef<'a>> {
    let mut deserializer = serde_json::Deserializer::from_slice(bytes);
    let mut serializer = Serializer::from_context(context)?;
    serde_transcode::transcode(&mut deserializer, &mut serializer)?;
    Ok(serializer.value)
}

/// Transcodes a [`JSValueRef`] into a JSON encoded byte vector.
pub fn transcode_output(val: JSValueRef) -> Result<Vec<u8>> {
    let mut output = Vec::new();
    let mut deserializer = Deserializer::from(val);
    let mut serializer = serde_json::Serializer::new(&mut output);
    serde_transcode::transcode(&mut deserializer, &mut serializer)?;
    Ok(output)
}

/// gets the config data from the host as a JSValueRef
pub fn get_config(context: &JSContextRef) -> Result<JSValueRef> {
    let cfg_size = unsafe { kite_get_config_size() } as usize;

    let mut buf: Vec<u8> = Vec::with_capacity(cfg_size);
    let ptr = buf.as_mut_ptr();
    unsafe { kite_get_config(ptr as u32) };

    let cfg_buf = unsafe { Vec::from_raw_parts(ptr, cfg_size, cfg_size) };

    Ok(transcode_input(context, &cfg_buf)?)
}

/// sets the event response on the host
pub fn set_manifest(output: JSValueRef) -> Result<()> {
    let output = transcode_output(output)?;

    let size = output.len();
    let ptr = output.as_ptr();

    unsafe {
        kite_set_manifest(ptr as u32, size as u32);
    }
    Ok(())
}

/// sets the event response on the host
pub fn log(level: u32, message: String) {
    let ptr = message.as_ptr();

    unsafe {
        kite_log(level, ptr as u32, message.len() as u32);
    }
}

/// gets the event data from the host as a JSValueRef
pub fn get_event(context: &JSContextRef, size: usize) -> Result<JSValueRef> {
    let mut buf: Vec<u8> = Vec::with_capacity(size);
    let ptr = buf.as_mut_ptr();
    unsafe { kite_get_event(ptr as u32) };

    let event_buf = unsafe { Vec::from_raw_parts(ptr, size, size) };

    Ok(transcode_input(context, &event_buf)?)
}

/// sets the event response on the host
pub fn set_event_response(output: JSValueRef) -> Result<()> {
    let output = transcode_output(output)?;

    let size = output.len();
    let ptr = output.as_ptr();

    unsafe {
        kite_set_event_response(ptr as u32, size as u32);
    }
    Ok(())
}

/// sets the event response on the host
pub fn call<'a>(context: &'a JSContextRef, call: JSValueRef) -> Result<JSValueRef<'a>> {
    let raw_call = transcode_output(call)?;

    let size = raw_call.len();
    let ptr = raw_call.as_ptr();

    let resp_size = unsafe {
        kite_call(ptr as u32, size as u32)
    } as usize;

    let mut buf: Vec<u8> = Vec::with_capacity(resp_size);
    let ptr = buf.as_mut_ptr();

    unsafe { kite_get_call_response(ptr as u32) };

    let resp_buf = unsafe { Vec::from_raw_parts(ptr, resp_size, resp_size) };

    Ok(transcode_input(context, &resp_buf)?)
}
