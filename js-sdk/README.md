# Kite JavaScript SDK

This directory contains all the necessary things to write Kite plugins in JavaScript. To get started you first have to install the compiler.

## Install the Compiler

### Install and setup Rust

```sh
# Install rust from https://rustup.rs
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Add the WASM (WASI) target
rustup target add wasm32-wasi
```

### Build the engine

Go to the [kitejs-engine](kitejs-engine) directory.

```sh
cargo build --release --target wasm32-wasi
```

### Install the compiler

Go to the [kitejs-compiler](kitejs-compiler) directory.

```sh
# This can take a while!
cargo install --path .
```
