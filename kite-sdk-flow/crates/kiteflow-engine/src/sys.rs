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
