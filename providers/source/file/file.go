package file

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/ArmanAvanesyan/go-config/config"
)

// Source provides config from a file (path and optional format override).
type Source struct {
	path   string
	format string
}

// New creates a file-backed source and infers the format from the file
// extension.
func New(path string) *Source {
	return &Source{
		path:   path,
		format: inferFormat(path),
	}
}

// WithFormat creates a file-backed source with an explicit format override.
func WithFormat(path, format string) *Source {
	return &Source{
		path:   path,
		format: strings.ToLower(format),
	}
}

func (s *Source) Read(_ context.Context) (any, error) {
	b, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}

	return &config.Document{
		Name:   s.path,
		Format: s.format,
		Raw:    b,
	}, nil
}

func inferFormat(path string) string {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")
	return strings.ToLower(ext)
}
