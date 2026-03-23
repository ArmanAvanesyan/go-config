package bytes

import (
	"context"

	"github.com/ArmanAvanesyan/go-config/config"
)

// Source provides config from an in-memory byte slice (name, format, and raw bytes).
type Source struct {
	name   string
	format string
	raw    []byte
}

// New creates a bytes-backed source with the given document name and format.
func New(name, format string, raw []byte) *Source {
	return &Source{
		name:   name,
		format: format,
		raw:    raw,
	}
}

func (s *Source) Read(_ context.Context) (any, error) {
	return &config.Document{
		Name:   s.name,
		Format: s.format,
		Raw:    s.raw,
	}, nil
}
