package tree

// CloneMap deep-clones a map[string]any tree.
func CloneMap(in map[string]any) map[string]any {
	if in == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(in))
	for k, v := range in {
		if nested, ok := v.(map[string]any); ok {
			out[k] = CloneMap(nested)
			continue
		}
		out[k] = v
	}
	return out
}
