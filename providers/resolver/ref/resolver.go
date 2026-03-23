package ref

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/internal/tree"
)

const prefix = "${REF:"
const suffix = "}"

// ErrRefNotFound is returned when a ${REF:path} placeholder references a missing key.
var ErrRefNotFound = errors.New("ref: path not found in config tree")

// Resolver replaces ${REF:key.path} placeholders with values from the config tree.
type Resolver struct{}

// New returns a new REF Resolver.
func New() *Resolver {
	return &Resolver{}
}

// Resolve walks the tree and expands ${REF:key.path} strings using the current tree.
// Missing paths return an error. Non-string values are converted with fmt.Sprint.
func (r *Resolver) Resolve(ctx context.Context, tr map[string]any) (map[string]any, error) {
	out, err := resolveMap(tr, tr)
	if err != nil {
		return nil, err
	}
	return out, nil
}

var _ config.Resolver = (*Resolver)(nil)

func resolveMap(in, root map[string]any) (map[string]any, error) {
	out := make(map[string]any, len(in))
	for k, v := range in {
		switch t := v.(type) {
		case string:
			resolved, err := resolveString(t, root)
			if err != nil {
				return nil, err
			}
			out[k] = resolved
		case map[string]any:
			nested, err := resolveMap(t, root)
			if err != nil {
				return nil, err
			}
			out[k] = nested
		case []any:
			nested, err := resolveSlice(t, root)
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

func resolveSlice(in []any, root map[string]any) ([]any, error) {
	out := make([]any, len(in))
	for i, v := range in {
		switch t := v.(type) {
		case string:
			resolved, err := resolveString(t, root)
			if err != nil {
				return nil, err
			}
			out[i] = resolved
		case map[string]any:
			nested, err := resolveMap(t, root)
			if err != nil {
				return nil, err
			}
			out[i] = nested
		case []any:
			nested, err := resolveSlice(t, root)
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

func resolveString(s string, root map[string]any) (string, error) {
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
			return "", fmt.Errorf("%w: empty path", ErrRefNotFound)
		}
		val, ok := tree.Get(root, path)
		if !ok {
			return "", fmt.Errorf("%w: %q", ErrRefNotFound, path)
		}
		switch v := val.(type) {
		case string:
			out.WriteString(v)
		default:
			fmt.Fprint(&out, v)
		}
	}
	return out.String(), nil
}
