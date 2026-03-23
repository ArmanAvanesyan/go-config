// WASM parser for JSON using serde_json.
// Rust's serde_json outperforms Go's encoding/json on large payloads.
//
// ABI — identical to toml_parser and yaml_parser:
//   wasm_alloc(size)   -> ptr
//   wasm_dealloc(ptr, size)
//   parse(ptr, len)    -> i32   0 = success, -1 = error
//   output_ptr()       -> ptr   JSON bytes
//   output_len()       -> u32

use std::sync::{Mutex, OnceLock};

static OUTPUT: OnceLock<Mutex<Vec<u8>>> = OnceLock::new();

fn output_buf() -> &'static Mutex<Vec<u8>> {
    OUTPUT.get_or_init(|| Mutex::new(Vec::new()))
}

#[no_mangle]
pub unsafe extern "C" fn wasm_alloc(size: u32) -> *mut u8 {
    let mut buf: Vec<u8> = Vec::with_capacity(size as usize);
    let ptr = buf.as_mut_ptr();
    std::mem::forget(buf);
    ptr
}

#[no_mangle]
pub unsafe extern "C" fn wasm_dealloc(ptr: *mut u8, size: u32) {
    drop(Vec::from_raw_parts(ptr, 0, size as usize));
}

#[no_mangle]
pub unsafe extern "C" fn parse(ptr: *const u8, len: u32) -> i32 {
    let input = std::slice::from_raw_parts(ptr, len as usize);
    let mut out = output_buf().lock().expect("output mutex poisoned");
    match do_parse(input) {
        Ok(json) => {
            out.clear();
            out.extend_from_slice(&json);
            0
        }
        Err(e) => {
            out.clear();
            out.extend_from_slice(e.to_string().as_bytes());
            -1
        }
    }
}

#[no_mangle]
pub extern "C" fn output_ptr() -> *const u8 {
    let out = output_buf().lock().expect("output mutex poisoned");
    out.as_ptr()
}

#[no_mangle]
pub extern "C" fn output_len() -> u32 {
    let out = output_buf().lock().expect("output mutex poisoned");
    out.len() as u32
}

fn do_parse(input: &[u8]) -> Result<Vec<u8>, Box<dyn std::error::Error>> {
    let value: serde_json::Value = serde_json::from_slice(input)?;
    Ok(serde_json::to_vec(&value)?)
}
