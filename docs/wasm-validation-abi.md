# Validation WASM ABI

Policy and validation modules used by go-config's WASM validator must export the following ABI. The host (Go) passes config as JSON bytes in guest memory and calls `validate`; the guest returns a status and, on failure, an error message.

| Export         | Signature                     | Purpose                                                                                            |
| -------------- | ----------------------------- | -------------------------------------------------------------------------------------------------- |
| `wasm_alloc`   | `(size: u32) -> u32`           | Allocate buffer in guest memory; host writes config JSON here.                                     |
| `wasm_dealloc` | `(ptr: u32, size: u32)`        | Free buffer.                                                                                       |
| `validate`     | `(ptr: u32, len: u32) -> i32`  | Run policy/validation on JSON at ptr/len. Return 0 = pass, non-zero = fail.                        |
| `error_ptr`    | `() -> u32`                    | After failed validate, pointer to UTF-8 error message (valid until next validate or module close). |
| `error_len`    | `() -> u32`                    | Length of error message in bytes.                                                                 |

- **Input**: Config as JSON bytes in guest memory. The host allocates with `wasm_alloc`, writes the JSON, then calls `validate(ptr, len)`.
- **Output**: `validate` returns 0 for success. On failure it returns non-zero; the host then reads the error message via `error_ptr` and `error_len`.
- **Memory**: No WASI is required for this ABI. The default policy crate is built with `wasm32-wasip1` for consistency with the parser crates.

Custom policy WASM binaries that export this ABI can be loaded with `rustpolicy.NewFromBytes(ctx, wasmBytes)`.
