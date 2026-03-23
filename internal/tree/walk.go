package tree

// Walk visits every node in a tree in depth-first order.
// The callback receives dot-separated paths for nested values.
func Walk(tree map[string]any, visit func(path string, v any)) {
	if tree == nil || visit == nil {
		return
	}
	walkMap("", tree, visit)
}

func walkMap(prefix string, m map[string]any, visit func(path string, v any)) {
	for k, v := range m {
		path := join(prefix, k)
		visit(path, v)
		if nested, ok := v.(map[string]any); ok {
			walkMap(path, nested, visit)
		}
	}
}

func join(prefix, seg string) string {
	if prefix == "" {
		return seg
	}
	return prefix + "." + seg
}
