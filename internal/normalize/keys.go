package normalize

import "strings"

// Key canonicalizes a single config key for stable matching.
func Key(key string) string {
	return strings.ToLower(strings.TrimSpace(key))
}

// Keys returns a new tree with recursively normalized map keys.
func Keys(tree map[string]any) map[string]any {
	if tree == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(tree))
	for k, v := range tree {
		nk := Key(k)
		if nested, ok := v.(map[string]any); ok {
			out[nk] = Keys(nested)
			continue
		}
		out[nk] = v
	}
	return out
}
