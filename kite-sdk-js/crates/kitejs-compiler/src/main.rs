use std::{env, fs, io::{Read, Write}, rc::Rc, sync::OnceLock};

use anyhow::{anyhow, Result};
use wasi_common::{pipe::ReadPipe, WasiCtx};
use wasmtime::Linker;
use wasmtime_wasi::WasiCtxBuilder;
use wizer::Wizer;
use binaryen::{CodegenConfig, Module};
use std::io;

static mut WASI: OnceLock<WasiCtx> = OnceLock::new();

fn main() -> Result<()> {
    let args: Vec<String> = env::args().collect();

    let input_file = args.get(1).ok_or(anyhow!("No input file specified"))?;
    let output_file = args.get(2).ok_or(anyhow!("No output file specified"))?;
    let optimize = args.get(3).map(|s| s == "--optimize").unwrap_or(false);

    let js = {
        if input_file == "-" {
            let mut js = Vec::new();
            io::stdin().read_to_end(&mut js)?;
            js
        } else {
            fs::read(input_file)?
        }
    };

    let wasm = generate_module(js, optimize)?;

    if output_file == "-" {
        io::stdout().write_all(&wasm)?;
    } else {
        fs::write(output_file, wasm)?;
    }
    
    Ok(())
}

fn generate_module(js: Vec<u8>, optimize: bool) -> Result<Vec<u8>> {
    let wasm = include_bytes!(concat!(env!("OUT_DIR"), "/engine.wasm"));

    let wasi = WasiCtxBuilder::new()
        .stdin(Box::new(ReadPipe::from(js)))
        .inherit_stdout()
        .inherit_stderr()
        .build();

    // We can't move the WasiCtx into `make_linker` since WasiCtx doesn't implement the `Copy` trait.
    // So we move the WasiCtx into a mutable static OnceLock instead.
    // Setting the value in the `OnceLock` and getting the reference back from it should be safe given
    // we're never executing this code concurrently. This code will also fail if `generate` is invoked
    // more than once per execution.
    if unsafe { WASI.set(wasi) }.is_err() {
        panic!("Failed to set WASI static variable")
    }

    let wasm = Wizer::new()
        .make_linker(Some(Rc::new(|engine| {
            let mut linker = Linker::new(engine);
            wasmtime_wasi::add_to_linker(&mut linker, |_ctx: &mut Option<WasiCtx>| {
                unsafe { WASI.get_mut() }.unwrap()
            })?;
            Ok(linker)
        })))?
        .init_func("_start")
        .wasm_bulk_memory(true)
        .run(wasm)
        .map_err(|_| anyhow!("JS compilation failed"))?;

    if optimize {
        // This is more or less equivalent to running `wasm-opt -O4 --strip` on the engine module beforehand.

        let codegen_cfg = CodegenConfig {
            optimization_level: 4, // Aggressively optimize for speed.
            shrink_level: 0,       // Don't optimize for size at the expense of performance.
            debug_info: false,
        };
    
        let mut module = Module::read(&wasm)
            .map_err(|_| anyhow!("Unable to read wasm binary for wasm-opt optimizations"))?;
    
        module.optimize(&codegen_cfg);
        module
            .run_optimization_passes(vec!["strip"], &codegen_cfg)
            .map_err(|_| anyhow!("Running wasm-opt optimization passes failed"))?;
    
        Ok(module.write())
    } else {
        Ok(wasm)
    }
}