// Policy/validation WASM module for go-config.
// Exports the validation ABI: wasm_alloc, wasm_dealloc, validate, error_ptr, error_len.
// See docs/wasm-validation-abi.md in the Go repo.

const MAX_ERROR_LEN: usize = 1024;

use std::sync::{Mutex, OnceLock};

static ERROR_BUF: OnceLock<Mutex<Vec<u8>>> = OnceLock::new();

fn error_buf() -> &'static Mutex<Vec<u8>> {
    ERROR_BUF.get_or_init(|| Mutex::new(Vec::with_capacity(MAX_ERROR_LEN)))
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
pub unsafe extern "C" fn validate(ptr: *const u8, len: u32) -> i32 {
    let input = std::slice::from_raw_parts(ptr, len as usize);
    match do_validate(input) {
        Ok(()) => 0,
        Err(msg) => {
            let bytes = msg.as_bytes();
            let n = bytes.len().min(MAX_ERROR_LEN);
            let mut buf = error_buf().lock().expect("error mutex poisoned");
            buf.clear();
            buf.extend_from_slice(&bytes[..n]);
            -1
        }
    }
}

#[no_mangle]
pub extern "C" fn error_ptr() -> *const u8 {
    let buf = error_buf().lock().expect("error mutex poisoned");
    buf.as_ptr()
}

#[no_mangle]
pub extern "C" fn error_len() -> u32 {
    let buf = error_buf().lock().expect("error mutex poisoned");
    buf.len() as u32
}

fn do_validate(input: &[u8]) -> Result<(), String> {
    let value: serde_json::Value = serde_json::from_slice(input)
        .map_err(|e| format!("invalid JSON: {}", e))?;
    // v1: allow all; optional minimal check — config must be an object or array
    match value {
        serde_json::Value::Object(_) | serde_json::Value::Array(_) | serde_json::Value::Null => {}
        _ => {
            // Allow primitives too so arbitrary decoded config is acceptable
        }
    }
    Ok(())
}
