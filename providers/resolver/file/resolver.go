package file

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/ArmanAvanesyan/go-config/config"
)

var errEmptyFilePath = errors.New("file resolver: empty path")

const prefix = "${FILE:"
const suffix = "}"

// Resolver replaces ${FILE:path} placeholders with file contents.
// Paths are absolute or relative to the process working directory.
// File contents are trimmed of a single trailing newline for convenience.
type Resolver struct{}

// New returns a new FILE Resolver.
func New() *Resolver {
	return &Resolver{}
}

// Resolve walks the tree and expands ${FILE:path} strings with file contents.
func (r *Resolver) Resolve(ctx context.Context, tree map[string]any) (map[string]any, error) {
	out, err := resolveMap(tree)
	if err != nil {
		return nil, err
	}
	return out, nil
}

var _ config.Resolver = (*Resolver)(nil)

func resolveMap(in map[string]any) (map[string]any, error) {
	out := make(map[string]any, len(in))
	for k, v := range in {
		switch t := v.(type) {
		case string:
			resolved, err := resolveString(t)
			if err != nil {
				return nil, err
			}
			out[k] = resolved
		case map[string]any:
			nested, err := resolveMap(t)
			if err != nil {
				return nil, err
			}
			out[k] = nested
		case []any:
			nested, err := resolveSlice(t)
			if err != nil {
				return nil, err
			}
			out[k] = nested
		default:
			out[k] = v
		}
	}
	return out, nil
}

func resolveSlice(in []any) ([]any, error) {
	out := make([]any, len(in))
	for i, v := range in {
		switch t := v.(type) {
		case string:
			resolved, err := resolveString(t)
			if err != nil {
				return nil, err
			}
			out[i] = resolved
		case map[string]any:
			nested, err := resolveMap(t)
			if err != nil {
				return nil, err
			}
			out[i] = nested
		case []any:
			nested, err := resolveSlice(t)
			if err != nil {
				return nil, err
			}
			out[i] = nested
		default:
			out[i] = v
		}
	}
	return out, nil
}

func resolveString(s string) (string, error) {
	var out strings.Builder
	remaining := s
	for {
		i := strings.Index(remaining, prefix)
		if i < 0 {
			out.WriteString(remaining)
			break
		}
		out.WriteString(remaining[:i])
		remaining = remaining[i+len(prefix):]
		j := strings.Index(remaining, suffix)
		if j < 0 {
			out.WriteString(prefix)
			out.WriteString(remaining)
			break
		}
		path := strings.TrimSpace(remaining[:j])
		remaining = remaining[j+len(suffix):]
		if path == "" {
			return "", errEmptyFilePath
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		content := strings.TrimSuffix(string(body), "\n")
		out.WriteString(content)
	}
	return out.String(), nil
}
