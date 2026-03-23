package generate

import "encoding/json"

// MarshalSchema serializes a schema document with indentation.
func MarshalSchema(s *Schema) ([]byte, error) {
	return json.MarshalIndent(s, "", "  ")
}
