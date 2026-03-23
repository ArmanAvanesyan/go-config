# go_config_policy_validator

WASM policy/validation module for [go-config](https://github.com/ArmanAvanesyan/go-config). It implements the validation WASM ABI used by the Go library's `extensions/wasm/validator` package.

**ABI:** The crate exports `wasm_alloc`, `wasm_dealloc`, `validate`, `error_ptr`, and `error_len`. See the Go repo's [docs/architecture.md#validation-wasm-abi](../../../docs/architecture.md#validation-wasm-abi) for the contract.

**Build:** From the Go repo root, `cd rust && make build-policy` compiles this crate to `wasm32-wasip1` and copies the `.wasm` to `extensions/wasm/validator/rustpolicy/policy.wasm`.

**Current behaviour:** v0.1 parses the config JSON and accepts all valid JSON (allow-all). Future versions can add required keys, value checks, or JSON Schema validation.
