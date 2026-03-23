WASM_TARGET = wasm32-wasip1
RELEASE_DIR  = target/$(WASM_TARGET)/release

GO_ROOT      = ..
RUSTTOML_DST = $(GO_ROOT)/extensions/wasm/parser/rusttoml
RUSTYAML_DST = $(GO_ROOT)/extensions/wasm/parser/rustyaml
RUSTJSON_DST = $(GO_ROOT)/extensions/wasm/parser/rustjson
POLICY_DST   = $(GO_ROOT)/extensions/wasm/validator/rustpolicy

.PHONY: all build-toml build-yaml build-json build-policy clean

all: build-toml build-yaml build-json build-policy

build-toml:
	cargo build --manifest-path parsers/toml-parser/Cargo.toml \
		--target $(WASM_TARGET) --release --locked
	cp parsers/toml-parser/$(RELEASE_DIR)/toml_parser.wasm $(RUSTTOML_DST)/toml_parser.wasm

build-yaml:
	cargo build --manifest-path parsers/yaml-parser/Cargo.toml \
		--target $(WASM_TARGET) --release --locked
	cp parsers/yaml-parser/$(RELEASE_DIR)/yaml_parser.wasm $(RUSTYAML_DST)/yaml_parser.wasm

build-json:
	cargo build --manifest-path parsers/json-parser/Cargo.toml \
		--target $(WASM_TARGET) --release --locked
	cp parsers/json-parser/$(RELEASE_DIR)/json_parser.wasm $(RUSTJSON_DST)/json_parser.wasm

build-policy:
	cargo build --manifest-path validators/config-policy/Cargo.toml \
		--target $(WASM_TARGET) --release --locked
	cp validators/config-policy/$(RELEASE_DIR)/go_config_policy_validator.wasm $(POLICY_DST)/policy.wasm

clean:
	cargo clean --manifest-path parsers/toml-parser/Cargo.toml
	cargo clean --manifest-path parsers/yaml-parser/Cargo.toml
	cargo clean --manifest-path parsers/json-parser/Cargo.toml
	cargo clean --manifest-path validators/config-policy/Cargo.toml
