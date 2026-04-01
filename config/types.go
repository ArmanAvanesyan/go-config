package config

// Metadata holds optional key-value metadata attached to a document (e.g. source name, format).
type Metadata map[string]string

// MissingPolicy controls behavior when source read fails due to missing input.
type MissingPolicy int

const (
	// MissingPolicyDefault preserves backward-compatible behavior via Required.
	MissingPolicyDefault MissingPolicy = iota
	// MissingPolicyIgnore treats missing source as empty tree.
	MissingPolicyIgnore
	// MissingPolicyFail fails loading when source is missing.
	MissingPolicyFail
)

// ParsePolicy controls behavior when parser fails for a source.
type ParsePolicy int

const (
	// ParsePolicyDefault fails load on parse errors.
	ParsePolicyDefault ParsePolicy = iota
	// ParsePolicyIgnore treats parse failure as empty tree.
	ParsePolicyIgnore
	// ParsePolicyFail fails load on parse errors.
	ParsePolicyFail
)

// SourceMeta holds per-source metadata for merge order and error handling.
// Higher Priority is merged later (wins). Required false means read failures are treated as empty tree.
type SourceMeta struct {
	Priority      int           // Higher = merged later (default 0, preserve registration order when equal)
	Required      bool          // If true, Read failure fails Load; if false, use empty tree and continue (default true)
	MissingPolicy MissingPolicy // Fine-grained missing-input behavior for file-like sources.
	ParsePolicy   ParsePolicy   // Fine-grained parse failure behavior.
}
