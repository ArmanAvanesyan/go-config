package deep

import (
	"github.com/ArmanAvanesyan/go-config/internal/tree"
	"github.com/ArmanAvanesyan/go-config/providers/merge"
)

type strategy struct{}

// New returns a merge strategy that performs a deep merge of nested maps,
// with values from src overriding those in dst.
func New() merge.Strategy {
	return strategy{}
}

func (strategy) Merge(dst, src map[string]any) (map[string]any, error) {
	if dst == nil {
		dst = map[string]any{}
	}
	out := tree.CloneMap(dst)
	if src == nil {
		return out, nil
	}

	deepMerge(out, src)
	return out, nil
}

func deepMerge(dst, src map[string]any) {
	for k, v := range src {
		if srcMap, ok := v.(map[string]any); ok {
			if dstMap, ok := dst[k].(map[string]any); ok {
				deepMerge(dstMap, srcMap)
				dst[k] = dstMap
				continue
			}
			// Avoid aliasing src subtrees in the output map.
			dst[k] = tree.CloneMap(srcMap)
			continue
		}
		dst[k] = v
	}
}
