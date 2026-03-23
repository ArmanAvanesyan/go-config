package generate

// Schema represents a JSON Schema document or subschema node.
//
// It is intentionally small and focused on the keywords needed for
// describing typical configuration structs:
//   - type: object / array / string / integer / number / boolean
//   - properties / required for structs
//   - items for slices and arrays
//   - additionalProperties for map values
//   - format (e.g. time.RFC3339 as date-time)
//   - $schema / $id for the root
//   - $defs / $ref for shared nested definitions
type Schema struct {
	// Meta
	Schema      string `json:"$schema,omitempty"`
	ID          string `json:"$id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`

	// Core type information
	Type   string `json:"type,omitempty"`
	Format string `json:"format,omitempty"`
	Ref    string `json:"$ref,omitempty"`

	// Object members
	Properties           map[string]*Schema `json:"properties,omitempty"`
	Required             []string           `json:"required,omitempty"`
	AdditionalProperties *Schema            `json:"additionalProperties,omitempty"`

	// Array items
	Items *Schema `json:"items,omitempty"`

	// Shared definitions for reuse / recursion
	Defs map[string]*Schema `json:"$defs,omitempty"`
}
