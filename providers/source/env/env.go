package env

import (
	"context"
	"os"
	"reflect"
	"sort"
	"strings"

	"github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/internal/normalize"
)

// BindingPrecedence controls resolution order between explicit bindings and inferred names.
type BindingPrecedence int

const (
	// ExplicitFirst resolves explicit key bindings before inferred env names.
	ExplicitFirst BindingPrecedence = iota
	// InferredFirst resolves inferred env names before explicit bindings.
	InferredFirst
	// ExplicitOnly resolves only explicit bindings and skips inferred names.
	ExplicitOnly
)

// NamingPolicy controls key/env normalization and inference behavior.
type NamingPolicy struct {
	DotToUnderscore    bool
	HyphenToUnderscore bool
	UppercaseInferred  bool
}

// Options configures environment source binding behavior.
type Options struct {
	Prefix             string
	Bindings           map[string][]string
	Infer              bool
	Precedence         BindingPrecedence
	Naming             NamingPolicy
	UseStructTagEnvFor any
}

// Source provides config from environment variables, optionally filtered by prefix.
type Source struct {
	opts Options
}

// New creates an environment source. When prefix is non-empty, only
// variables starting with PREFIX_ are considered, and the prefix is
// stripped before building the config tree.
func New(prefix string) *Source {
	return NewWithOptions(Options{Prefix: prefix})
}

// NewWithOptions creates an environment source with explicit binding controls.
func NewWithOptions(opts Options) *Source {
	if !opts.Infer && len(opts.Bindings) == 0 && opts.UseStructTagEnvFor == nil {
		opts.Infer = true
	}
	if !opts.Naming.DotToUnderscore && !opts.Naming.HyphenToUnderscore && !opts.Naming.UppercaseInferred {
		opts.Naming = NamingPolicy{
			DotToUnderscore:    true,
			HyphenToUnderscore: true,
			UppercaseInferred:  true,
		}
	}
	if opts.UseStructTagEnvFor != nil {
		tagBindings := BindingsFromStruct(opts.UseStructTagEnvFor)
		if opts.Bindings == nil {
			opts.Bindings = map[string][]string{}
		}
		for k, v := range tagBindings {
			opts.Bindings[k] = append(opts.Bindings[k], v...)
		}
	}
	return &Source{opts: opts}
}

func (s *Source) Read(_ context.Context) (any, error) {
	tree := map[string]any{}
	envMap := snapshotEnv()

	if s.opts.Precedence == InferredFirst && s.opts.Infer {
		s.applyInferred(tree, envMap)
	}
	s.applyExplicit(tree, envMap)
	if s.opts.Precedence == ExplicitFirst && s.opts.Infer {
		s.applyInferred(tree, envMap)
	}
	if s.opts.Precedence != InferredFirst && s.opts.Precedence != ExplicitFirst && s.opts.Infer {
		s.applyInferred(tree, envMap)
	}

	return &config.TreeDocument{
		Name: "env",
		Tree: tree,
	}, nil
}

func snapshotEnv() map[string]string {
	out := map[string]string{}
	for _, entry := range os.Environ() {
		parts := strings.SplitN(entry, "=", 2)
		key := parts[0]
		val := ""
		if len(parts) == 2 {
			val = parts[1]
		}
		out[key] = val
	}
	return out
}

func (s *Source) applyExplicit(tree map[string]any, envMap map[string]string) {
	keys := make([]string, 0, len(s.opts.Bindings))
	for k := range s.opts.Bindings {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, key := range keys {
		aliases := append([]string{}, s.opts.Bindings[key]...)
		if s.opts.Infer {
			aliases = append(aliases, inferredNameForKey(key, s.opts.Naming))
		}
		for _, alias := range aliases {
			full := s.withPrefix(alias)
			if val, ok := envMap[full]; ok {
				insert(tree, normalizePathFromConfigKey(key), val, s.opts.Precedence == ExplicitFirst)
				break
			}
		}
	}
}

func (s *Source) applyInferred(tree map[string]any, envMap map[string]string) {
	for key, val := range envMap {
		if s.opts.Prefix != "" {
			prefix := s.opts.Prefix + "_"
			if !strings.HasPrefix(key, prefix) {
				continue
			}
			key = strings.TrimPrefix(key, prefix)
		}
		path := strings.Split(strings.ToLower(key), "__")
		for i := range path {
			path[i] = normalize.Key(path[i])
		}
		insert(tree, path, val, s.opts.Precedence != InferredFirst)
	}
}

func (s *Source) withPrefix(envName string) string {
	if s.opts.Prefix == "" || strings.HasPrefix(envName, s.opts.Prefix+"_") {
		return envName
	}
	return s.opts.Prefix + "_" + envName
}

func normalizePathFromConfigKey(key string) []string {
	key = strings.TrimSpace(key)
	key = strings.ReplaceAll(key, "-", "_")
	segs := strings.Split(key, ".")
	out := make([]string, 0, len(segs))
	for _, s := range segs {
		if t := normalize.Key(s); t != "" {
			out = append(out, t)
		}
	}
	return out
}

func inferredNameForKey(key string, naming NamingPolicy) string {
	name := key
	if naming.DotToUnderscore {
		name = strings.ReplaceAll(name, ".", "_")
	}
	if naming.HyphenToUnderscore {
		name = strings.ReplaceAll(name, "-", "_")
	}
	if naming.UppercaseInferred {
		name = strings.ToUpper(name)
	}
	return name
}

// BindingsFromStruct derives key->env aliases from `env:"A,B"` tags.
func BindingsFromStruct(v any) map[string][]string {
	out := map[string][]string{}
	t := reflect.TypeOf(v)
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t == nil || t.Kind() != reflect.Struct {
		return out
	}
	var walk func(reflect.Type, []string)
	walk = func(rt reflect.Type, prefix []string) {
		for i := 0; i < rt.NumField(); i++ {
			f := rt.Field(i)
			if !f.IsExported() {
				continue
			}
			key := fieldKey(f)
			if key == "" || key == "-" {
				continue
			}
			path := append(append([]string{}, prefix...), normalize.Key(key))
			if tag, ok := f.Tag.Lookup("env"); ok {
				parts := strings.Split(tag, ",")
				aliases := make([]string, 0, len(parts))
				for _, p := range parts {
					p = strings.TrimSpace(p)
					if p != "" {
						aliases = append(aliases, p)
					}
				}
				if len(aliases) > 0 {
					out[strings.Join(path, ".")] = aliases
				}
			}
			ft := f.Type
			for ft.Kind() == reflect.Pointer {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				walk(ft, path)
			}
		}
	}
	walk(t, nil)
	return out
}

func fieldKey(f reflect.StructField) string {
	if tag, ok := f.Tag.Lookup("mapstructure"); ok {
		name := strings.Split(tag, ",")[0]
		if name != "" {
			return name
		}
	}
	if tag, ok := f.Tag.Lookup("json"); ok {
		name := strings.Split(tag, ",")[0]
		if name != "" {
			return name
		}
	}
	return f.Name
}

func insert(tree map[string]any, path []string, value string, preserveExisting bool) {
	if len(path) == 0 {
		return
	}
	current := tree
	for i, p := range path {
		if i == len(path)-1 {
			if preserveExisting {
				if _, exists := current[p]; exists {
					return
				}
			}
			current[p] = value
			return
		}
		next, ok := current[p].(map[string]any)
		if !ok {
			next = map[string]any{}
			current[p] = next
		}
		current = next
	}
}
