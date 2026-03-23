package env

import (
	"context"
	"os"
	"strings"

	"github.com/ArmanAvanesyan/go-config/config"
)

// Resolver replaces ${ENV:VAR} placeholders in config values with environment variables.
type Resolver struct{}

// New returns a new env Resolver.
func New() *Resolver {
	return &Resolver{}
}

// Resolve walks the tree and expands ${ENV:VAR} strings using os.Getenv.
func (r *Resolver) Resolve(_ context.Context, tree map[string]any) (map[string]any, error) {
	return resolveMap(tree), nil
}

func resolveMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		switch t := v.(type) {
		case string:
			out[k] = resolveString(t)
		case map[string]any:
			out[k] = resolveMap(t)
		case []any:
			out[k] = resolveSlice(t)
		default:
			out[k] = v
		}
	}
	return out
}

func resolveSlice(in []any) []any {
	out := make([]any, len(in))
	for i, v := range in {
		switch t := v.(type) {
		case string:
			out[i] = resolveString(t)
		case map[string]any:
			out[i] = resolveMap(t)
		default:
			out[i] = v
		}
	}
	return out
}

func resolveString(s string) string {
	if strings.HasPrefix(s, "${ENV:") && strings.HasSuffix(s, "}") {
		key := strings.TrimSuffix(strings.TrimPrefix(s, "${ENV:"), "}")
		return os.Getenv(key)
	}
	return s
}

var _ config.Resolver = (*Resolver)(nil)
