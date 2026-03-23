package config

import "context"

// Source reads configuration input.
//
// A source may return either:
//   - a raw Document, which requires a Parser
//   - a TreeDocument, which can skip parsing
type Source interface {
	Read(ctx context.Context) (any, error)
}
