cargo build --release --target wasm32-wasi
# wasm-opt -Os -o ./target/wasm32-wasi/release/kitejs-engine.wasm ./target/wasm32-wasi/release/kitejs-engine.wasm
# wizer --wasm-bulk-memory true --allow-wasi ./target/wasm32-wasi/release/kitejs-engine.wasm -o ../../../examples/automod/plugin.wasm
