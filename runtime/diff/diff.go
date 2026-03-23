// Package diff compares two configuration trees (map[string]any) and returns
// a list of path-level changes (add, remove, change).
package diff

import "reflect"

// Kind is the type of change at a path.
type Kind int

// Kinds of change reported by Changes().
const (
	KindAdd    Kind = iota // key present only in new tree
	KindRemove             // key present only in old tree
	KindChange             // key in both; value changed
)

// Change describes a single difference between old and new tree at a path.
type Change struct {
	Path     string // Dot-separated path, e.g. "server.port"
	Kind     Kind
	OldValue any
	NewValue any
}

// Changes recursively compares old and new trees and returns a list of changes.
// Paths are dot-separated. For nested maps, only leaf changes are reported
// (e.g. a changed value at "a.b.c", not intermediate keys).
// Nil maps are treated as empty.
func Changes(old, new map[string]any) []Change {
	if old == nil {
		old = map[string]any{}
	}
	if new == nil {
		new = map[string]any{}
	}
	var out []Change
	walk("", old, new, &out)
	return out
}

func walk(path string, old, new map[string]any, out *[]Change) {
	seen := make(map[string]bool)

	for k, oldVal := range old {
		seen[k] = true
		p := join(path, k)
		newVal, inNew := new[k]
		if !inNew {
			*out = append(*out, Change{Path: p, Kind: KindRemove, OldValue: oldVal})
			continue
		}
		oldMap, oldIsMap := oldVal.(map[string]any)
		newMap, newIsMap := newVal.(map[string]any)
		if oldIsMap && newIsMap {
			walk(p, oldMap, newMap, out)
			continue
		}
		if !valueEqual(oldVal, newVal) {
			*out = append(*out, Change{Path: p, Kind: KindChange, OldValue: oldVal, NewValue: newVal})
		}
	}

	for k, newVal := range new {
		if seen[k] {
			continue
		}
		p := join(path, k)
		*out = append(*out, Change{Path: p, Kind: KindAdd, NewValue: newVal})
	}
}

func join(prefix, seg string) string {
	if prefix == "" {
		return seg
	}
	return prefix + "." + seg
}

func valueEqual(a, b any) bool {
	return reflect.DeepEqual(a, b)
}
