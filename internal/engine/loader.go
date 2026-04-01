package engine

import (
	"context"

	"github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/providers/merge"
)

type sourceBinding struct {
	source config.Source
	parser config.Parser
}

// Binding is the typed contract for engine orchestration callers.
type Binding struct {
	Source config.Source
	Parser config.Parser
}

// LoadAndMerge reads all bindings, parses documents as needed, and merges
// them into a single configuration tree using the provided merge strategy.
func LoadAndMerge(ctx context.Context, bindings any, strategy merge.Strategy) (map[string]any, error) {
	// Supported typed binding shapes.
	switch typed := bindings.(type) {
	case []sourceBinding:
		out := make([]config.PipelineBinding, 0, len(typed))
		for _, b := range typed {
			out = append(out, config.PipelineBinding{Source: b.source, Parser: b.parser})
		}
		return config.LoadAndMergeBindings(ctx, out, strategy)
	case []Binding:
		out := make([]config.PipelineBinding, 0, len(typed))
		for _, b := range typed {
			out = append(out, config.PipelineBinding{Source: b.Source, Parser: b.Parser})
		}
		return config.LoadAndMergeBindings(ctx, out, strategy)
	default:
		// Compatibility wrapper path removed in Phase 4:
		// callers must provide engine.Binding explicitly.
		return nil, errBindingsMustBeBindingSlice()
	}
}
