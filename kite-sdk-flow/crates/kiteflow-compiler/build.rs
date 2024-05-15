use anyhow::Result;
use std::env;
use std::fs;
use std::path::{Path, PathBuf};

// https://github.com/bytecodealliance/javy/blob/main/crates/cli/build.rs

fn main() -> Result<()> {
    let cargo_manifest_dir = env::var("CARGO_MANIFEST_DIR")?;
    let engine_path = PathBuf::from(&cargo_manifest_dir)
        .parent()
        .unwrap()
        .join("kiteflow-engine/target/wasm32-wasi/release/kiteflow-engine.wasm");

    println!("cargo:rerun-if-changed={}", engine_path.to_str().unwrap());
    println!("cargo:rerun-if-changed=build.rs");

    if engine_path.exists() {
        let out_dir = env::var("OUT_DIR")?;
        let copied_engine_path = Path::new(&out_dir).join("engine.wasm");

        fs::copy(&engine_path, copied_engine_path)?;
    } else {
        return Err(anyhow::anyhow!(
            "engine release build not found, build kiteflow-engine first"
        ));
    }

    Ok(())
}
