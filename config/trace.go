package config

import (
	"fmt"
	"sort"
)

// Trace captures optional explain/provenance details for a load run.
type Trace struct {
	Keys      map[string]KeyTrace
	HookOrder []string
}

// KeyTrace describes provenance decisions for a single flattened key.
type KeyTrace struct {
	FinalSource string
	FinalValue  any
	Candidates  []TraceCandidate
}

// TraceCandidate records one source contribution for a key.
type TraceCandidate struct {
	Source   string
	Value    any
	Selected bool
}

type traceCollector struct {
	keys  map[string]KeyTrace
	hooks []string
}

func newTraceCollector(t *Trace) *traceCollector {
	if t == nil {
		return nil
	}
	t.Keys = map[string]KeyTrace{}
	t.HookOrder = nil
	return &traceCollector{
		keys: map[string]KeyTrace{},
	}
}

func (c *traceCollector) recordTree(source string, tree map[string]any) {
	if c == nil || tree == nil {
		return
	}
	flat := flattenTree(tree, "")
	keys := make([]string, 0, len(flat))
	for k := range flat {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := flat[k]
		kt := c.keys[k]
		for i := range kt.Candidates {
			kt.Candidates[i].Selected = false
		}
		kt.Candidates = append(kt.Candidates, TraceCandidate{
			Source:   source,
			Value:    v,
			Selected: true,
		})
		kt.FinalSource = source
		kt.FinalValue = v
		c.keys[k] = kt
	}
}

func (c *traceCollector) recordHook(name string) {
	if c == nil {
		return
	}
	c.hooks = append(c.hooks, name)
}

func (c *traceCollector) flush(t *Trace) {
	if c == nil || t == nil {
		return
	}
	t.Keys = c.keys
	t.HookOrder = append([]string{}, c.hooks...)
}

func flattenTree(tree map[string]any, prefix string) map[string]any {
	out := map[string]any{}
	for k, v := range tree {
		path := k
		if prefix != "" {
			path = fmt.Sprintf("%s.%s", prefix, k)
		}
		if nested, ok := v.(map[string]any); ok {
			for nk, nv := range flattenTree(nested, path) {
				out[nk] = nv
			}
			continue
		}
		out[path] = v
	}
	return out
}
