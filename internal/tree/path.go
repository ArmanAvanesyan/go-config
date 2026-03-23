// Package tree provides helpers for traversing and manipulating the
// generic map[string]any configuration tree used throughout go-config.
package tree

import "strings"

// Get returns the value at the dot-separated path (e.g. "server.port").
// It traverses nested maps only; array indices are not supported.
// Returns (nil, false) for missing keys or if a path segment is not a map.
func Get(tree map[string]any, path string) (any, bool) {
	if tree == nil || path == "" {
		return nil, false
	}
	segments := strings.Split(path, ".")
	current := any(tree)
	for i, seg := range segments {
		if seg == "" {
			return nil, false
		}
		m, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		v, ok := m[seg]
		if !ok {
			return nil, false
		}
		if i == len(segments)-1 {
			return v, true
		}
		current = v
	}
	return current, true
}
