// Package config provides a typed, source-agnostic, format-agnostic
// configuration loading pipeline for Go applications.
//
// Core package responsibilities:
//   - source orchestration
//   - merge pipeline
//   - decode pipeline
//   - validation pipeline
//
// Optional capabilities such as YAML, TOML, remote sources, and WASM-based
// parsing/validation live in separate packages.
package config
