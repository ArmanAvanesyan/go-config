package config

import "context"

// Parser parses a raw document into a tree representation.
type Parser interface {
	Parse(ctx context.Context, doc *Document) (map[string]any, error)
}

// TypedParser is an optional parser extension that can decode directly into
// a typed target, bypassing generic map tree materialization.
type TypedParser interface {
	ParseTyped(ctx context.Context, doc *Document, out any) error
}
