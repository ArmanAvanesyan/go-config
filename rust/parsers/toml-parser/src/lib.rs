// WASM parser for TOML using the toml crate (toml-rs).
//
// ABI — Go uses these exports:
//   wasm_alloc(size)   → ptr   allocate input buffer in WASM linear memory
//   wasm_dealloc(ptr, size)    free the input buffer
//   parse(ptr, len)    → i32   parse bytes; 0 = success, -1 = error
//   output_ptr()       → ptr   pointer to result JSON bytes
//   output_len()       → u32   length of result JSON bytes
//
// Output is msgpack bytes with transport prefix: b"GCFGMP1" + msgpack payload.
// On error, output bytes contain the error string (not JSON).

use std::sync::{Mutex, OnceLock};

static OUTPUT: OnceLock<Mutex<Vec<u8>>> = OnceLock::new();

fn output_buf() -> &'static Mutex<Vec<u8>> {
    OUTPUT.get_or_init(|| Mutex::new(Vec::new()))
}

/// Allocate `size` bytes in WASM linear memory and return the pointer.
/// The caller (Go) is responsible for writing `size` bytes at this address
/// before calling `parse`.
#[no_mangle]
pub unsafe extern "C" fn wasm_alloc(size: u32) -> *mut u8 {
    let mut buf: Vec<u8> = Vec::with_capacity(size as usize);
    let ptr = buf.as_mut_ptr();
    std::mem::forget(buf);
    ptr
}

/// Free memory previously allocated by `wasm_alloc`.
#[no_mangle]
pub unsafe extern "C" fn wasm_dealloc(ptr: *mut u8, size: u32) {
    drop(Vec::from_raw_parts(ptr, 0, size as usize));
}

/// Parse TOML bytes written by Go at `ptr`/`len`.
/// Stores JSON-serialised output in the static OUTPUT buffer.
/// Returns 0 on success or -1 on error (output contains the error message).
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

/// Return a pointer to the output buffer populated by the last `parse` call.
#[no_mangle]
pub extern "C" fn output_ptr() -> *const u8 {
    let out = output_buf().lock().expect("output mutex poisoned");
    out.as_ptr()
}

/// Return the byte length of the output buffer.
#[no_mangle]
pub extern "C" fn output_len() -> u32 {
    let out = output_buf().lock().expect("output mutex poisoned");
    out.len() as u32
}

fn do_parse(input: &[u8]) -> Result<Vec<u8>, Box<dyn std::error::Error>> {
    const PREFIX: &[u8] = b"GCFGMP1";
    let text = std::str::from_utf8(input)?;
    // toml v1 parser expects document parsing via `from_str` for full TOML docs.
    let value: toml::Value = toml::from_str(text)?;
    let payload = rmp_serde::to_vec_named(&value)?;
    let mut out = Vec::with_capacity(PREFIX.len() + payload.len());
    out.extend_from_slice(PREFIX);
    out.extend_from_slice(&payload);
    Ok(out)
}
