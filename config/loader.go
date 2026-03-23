package config

import (
	"context"
	"fmt"
	"sort"

	"github.com/ArmanAvanesyan/go-config/providers/merge"
)

type sourceBinding struct {
	source Source
	parser Parser
	meta   *SourceMeta
}

// PipelineBinding is an explicit input contract for shared pipeline orchestration.
// It is intentionally public to allow internal wrappers (e.g. internal/engine)
// to delegate merge orchestration without duplicating read/parse logic.
type PipelineBinding struct {
	Source Source
	Parser Parser
	Meta   *SourceMeta
}

// Loader builds and runs the config loading pipeline (sources, parse, merge, resolve, validate, decode).
type Loader struct {
	sources []sourceBinding
	options Options
}

// AddSource registers a source and optional parser; if no parser is given, format is inferred from the source.
func (l *Loader) AddSource(src Source, parser ...Parser) *Loader {
	return l.AddSourceWithMeta(src, parser, nil)
}

// AddSourceWithMeta registers a source with optional parser and metadata (priority, required).
// meta may be nil for default behavior (required, order-based precedence).
func (l *Loader) AddSourceWithMeta(src Source, parser []Parser, meta *SourceMeta) *Loader {
	var p Parser
	if len(parser) > 0 {
		p = parser[0]
	}
	l.sources = append(l.sources, sourceBinding{
		source: src,
		parser: p,
		meta:   meta,
	})
	return l
}

// LoadTyped runs the pipeline and decodes the result into a new instance of T.
// It is equivalent to allocating var t T and calling l.Load(ctx, &t), returning (t, err).
func LoadTyped[T any](ctx context.Context, l *Loader) (T, error) {
	var t T
	err := l.Load(ctx, &t)
	return t, err
}

// Load runs the pipeline and decodes the result into out.
func (l *Loader) Load(ctx context.Context, out any) error {
	if out == nil {
		return ErrNilTarget
	}

	if len(l.sources) == 0 {
		return ErrNoSources
	}

	if l.options.Decoder == nil {
		return ErrDecoderRequired
	}

	if ok, err := l.tryDirectDecode(ctx, out); ok {
		if err != nil {
			return err
		}
		if l.options.Validator != nil {
			if err := l.options.Validator.Validate(ctx, out); err != nil {
				return fmt.Errorf("%w: %v", ErrValidationFailed, err)
			}
		}
		return nil
	}

	tree, err := loadAndMerge(ctx, l.sources, l.options.MergeStrategy)
	if err != nil {
		return err
	}

	if l.options.Resolver != nil {
		tree, err = l.options.Resolver.Resolve(ctx, tree)
		if err != nil {
			return fmt.Errorf("%w: %v", ErrResolutionFailed, err)
		}
	}

	if err := l.options.Decoder.Decode(tree, out); err != nil {
		return fmt.Errorf("%w: %v", ErrDecodeFailed, err)
	}

	if l.options.Validator != nil {
		if err := l.options.Validator.Validate(ctx, out); err != nil {
			return fmt.Errorf("%w: %v", ErrValidationFailed, err)
		}
	}

	return nil
}

// loadTreeAndDecode runs the pipeline and decodes into out; it returns the
// merged-and-resolved tree (for diffing) or an error.
func (l *Loader) loadTreeAndDecode(ctx context.Context, out any) (map[string]any, error) {
	if len(l.sources) == 0 {
		return nil, ErrNoSources
	}
	if l.options.Decoder == nil {
		return nil, ErrDecoderRequired
	}
	tree, err := loadAndMerge(ctx, l.sources, l.options.MergeStrategy)
	if err != nil {
		return nil, err
	}
	if l.options.Resolver != nil {
		tree, err = l.options.Resolver.Resolve(ctx, tree)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrResolutionFailed, err)
		}
	}
	if err := l.options.Decoder.Decode(tree, out); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecodeFailed, err)
	}
	if l.options.Validator != nil {
		if err := l.options.Validator.Validate(ctx, out); err != nil {
			return nil, fmt.Errorf("%w: %v", ErrValidationFailed, err)
		}
	}
	return tree, nil
}

func (l *Loader) tryDirectDecode(ctx context.Context, out any) (bool, error) {
	if !l.options.DirectDecode || len(l.sources) != 1 || l.options.Resolver != nil {
		return false, nil
	}
	b := l.sources[0]
	tp, ok := b.parser.(TypedParser)
	if !ok {
		return false, nil
	}

	v, err := b.source.Read(ctx)
	if err != nil {
		if b.meta != nil && !b.meta.Required {
			return true, nil
		}
		return true, fmt.Errorf("%w: %v", ErrSourceReadFailed, err)
	}
	doc, ok := v.(*Document)
	if !ok || doc == nil {
		return false, nil
	}
	if err := tp.ParseTyped(ctx, doc, out); err != nil {
		return true, fmt.Errorf("%w: %v", ErrDecodeFailed, err)
	}
	return true, nil
}

// loadAndMerge is the core orchestration that reads all sources, parses
// documents as needed, and merges them into a single configuration tree
// using the provided merge strategy.
// Bindings are sorted by Priority ascending (higher priority merged last).
// If a source has meta.Required == false and Read fails, it is treated as empty tree.
func loadAndMerge(ctx context.Context, bindings []sourceBinding, strategy merge.Strategy) (map[string]any, error) {
	pipelineBindings := make([]PipelineBinding, 0, len(bindings))
	for _, b := range bindings {
		pipelineBindings = append(pipelineBindings, PipelineBinding{
			Source: b.source,
			Parser: b.parser,
			Meta:   b.meta,
		})
	}
	return LoadAndMergeBindings(ctx, pipelineBindings, strategy)
}

// LoadAndMergeBindings reads all bindings, parses documents as needed, and merges
// them into a single tree.
//
// Contracts:
//   - bindings are merged by ascending Meta.Priority (higher priority merged later)
//   - read failures for optional sources (Meta.Required=false) are treated as empty trees
//   - merge failures are wrapped with ErrMergeFailed
func LoadAndMergeBindings(ctx context.Context, bindings []PipelineBinding, strategy merge.Strategy) (map[string]any, error) {
	merged := map[string]any{}

	// Stable sort by Priority ascending so higher priority is merged last.
	sorted := make([]PipelineBinding, len(bindings))
	copy(sorted, bindings)
	sort.SliceStable(sorted, func(i, j int) bool {
		pi, pj := 0, 0
		if sorted[i].Meta != nil {
			pi = sorted[i].Meta.Priority
		}
		if sorted[j].Meta != nil {
			pj = sorted[j].Meta.Priority
		}
		return pi < pj
	})

	for _, b := range sorted {
		tree, err := readBinding(ctx, b.Source, b.Parser)
		if err != nil {
			if b.Meta != nil && !b.Meta.Required {
				tree = map[string]any{}
			} else {
				return nil, err
			}
		}

		merged, err = strategy.Merge(merged, tree)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrMergeFailed, err)
		}
	}

	return merged, nil
}

func readBinding(ctx context.Context, src Source, parser Parser) (map[string]any, error) {
	v, err := src.Read(ctx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSourceReadFailed, err)
	}

	switch doc := v.(type) {
	case *TreeDocument:
		if doc == nil || doc.Tree == nil {
			return nil, ErrInvalidDocument
		}
		return doc.Tree, nil

	case *Document:
		if doc == nil {
			return nil, ErrInvalidDocument
		}
		if parser == nil {
			return nil, ErrParserRequired
		}
		tree, err := parser.Parse(ctx, doc)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrParseFailed, err)
		}
		return tree, nil

	default:
		return nil, ErrInvalidDocument
	}
}
