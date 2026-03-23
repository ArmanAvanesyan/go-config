package config

// Metadata holds optional key-value metadata attached to a document (e.g. source name, format).
type Metadata map[string]string

// SourceMeta holds per-source metadata for merge order and error handling.
// Higher Priority is merged later (wins). Required false means read failures are treated as empty tree.
type SourceMeta struct {
	Priority int  // Higher = merged later (default 0, preserve registration order when equal)
	Required bool // If true, Read failure fails Load; if false, use empty tree and continue (default true)
}
