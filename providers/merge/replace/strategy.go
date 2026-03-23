package replace

import (
	"github.com/ArmanAvanesyan/go-config/internal/tree"
	"github.com/ArmanAvanesyan/go-config/providers/merge"
)

type strategy struct{}

// New returns a strategy that replaces the destination with the source
// at the top level (shallow replace — no recursive merging).
func New() merge.Strategy {
	return strategy{}
}

func (strategy) Merge(dst, src map[string]any) (map[string]any, error) {
	return tree.CloneMap(src), nil
}
