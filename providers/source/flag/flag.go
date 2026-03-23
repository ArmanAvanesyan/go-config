package flag

import (
	"context"
	goflag "flag"

	"github.com/ArmanAvanesyan/go-config/config"
)

// Source reads configuration values from registered command-line flags
// (flag.CommandLine). Only flags that have been explicitly set on the
// command line are included; unset flags with default values are skipped.
//
// Flag names are used as-is as config keys (dots are not interpreted as
// path separators). If you need nested keys, combine this source with a
// custom key-mapping layer.
type Source struct {
	fs *goflag.FlagSet
}

// New returns a Source that reads from flag.CommandLine.
func New() *Source {
	return &Source{fs: goflag.CommandLine}
}

// NewFromFlagSet returns a Source that reads from the provided FlagSet.
// Useful for testing or when you manage your own FlagSet.
func NewFromFlagSet(fs *goflag.FlagSet) *Source {
	return &Source{fs: fs}
}

// Read returns a TreeDocument containing all explicitly set flags.
func (s *Source) Read(_ context.Context) (any, error) {
	tree := make(map[string]any)

	s.fs.Visit(func(f *goflag.Flag) {
		tree[f.Name] = f.Value.String()
	})

	return &config.TreeDocument{
		Name: "flags",
		Tree: tree,
	}, nil
}
