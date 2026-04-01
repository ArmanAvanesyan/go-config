package config_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/ArmanAvanesyan/go-config/config"
	envsource "github.com/ArmanAvanesyan/go-config/providers/source/env"
)

type compatFixture struct {
	Name          string              `json:"name"`
	Description   string              `json:"description"`
	Strict        bool                `json:"strict"`
	Trees         []compatTreeFixture `json:"trees"`
	Env           compatEnvFixture    `json:"env"`
	Expected      map[string]any      `json:"expected"`
	ExpectedTrace map[string]string   `json:"expected_trace"`
}

type compatTreeFixture struct {
	Priority int            `json:"priority"`
	Tree     map[string]any `json:"tree"`
}

type compatEnvFixture struct {
	Priority   int                 `json:"priority"`
	Prefix     string              `json:"prefix"`
	Infer      bool                `json:"infer"`
	Precedence string              `json:"precedence"`
	Bindings   map[string][]string `json:"bindings"`
	Vars       map[string]string   `json:"vars"`
}

type compatOutput struct {
	Server struct {
		Host string `json:"host"`
		Port int    `json:"port"`
		TLS  bool   `json:"tls"`
	} `json:"server"`
	Feature struct {
		Enabled bool   `json:"enabled"`
		Mode    string `json:"mode"`
	} `json:"feature"`
	Timeouts struct {
		Request time.Duration `json:"request"`
	} `json:"timeouts"`
	Labels map[string]string `json:"labels"`
}

func TestContract_CompatParity_Fixtures(t *testing.T) {
	fixtures := mustLoadCompatFixtures(t)
	for _, fx := range fixtures {
		fx := fx
		t.Run(fx.Name, func(t *testing.T) {
			applyCompatEnv(t, fx.Env.Vars)

			trace := &config.Trace{}
			spec := config.Spec{
				Strict: fx.Strict,
				Trace:  trace,
			}
			for _, tree := range fx.Trees {
				spec.Trees = append(spec.Trees, config.TreeSpec{
					Tree:     tree.Tree,
					Priority: tree.Priority,
				})
			}
			if len(fx.Env.Vars) > 0 {
				spec.Sources = append(spec.Sources, config.SourceSpec{
					Source: envsource.NewWithOptions(envsource.Options{
						Prefix:     fx.Env.Prefix,
						Infer:      fx.Env.Infer,
						Precedence: mustParsePrecedence(t, fx.Env.Precedence),
						Bindings:   fx.Env.Bindings,
					}),
					Meta: &config.SourceMeta{
						Priority: fx.Env.Priority,
						Required: true,
					},
				})
			}

			got, err := config.LoadTypedWithSpec[compatOutput](context.Background(), spec)
			if err != nil {
				t.Fatalf("compat fixture %q failed to load: %v", fx.Name, err)
			}

			gotSnapshot := compatSnapshot(got)
			assertCompatNoDiff(t, fx.Name, fx.Expected, gotSnapshot)
			assertCompatTrace(t, fx.Name, fx.ExpectedTrace, trace)
		})
	}
}

func mustLoadCompatFixtures(t *testing.T) []compatFixture {
	t.Helper()
	dir := filepath.Join("..", "testdata", "compat")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read compat fixture dir %q: %v", dir, err)
	}
	fixtures := make([]compatFixture, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		b, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read fixture %q: %v", path, err)
		}
		var fx compatFixture
		if err := json.Unmarshal(b, &fx); err != nil {
			t.Fatalf("unmarshal fixture %q: %v", path, err)
		}
		if fx.Name == "" {
			t.Fatalf("fixture %q must include name", path)
		}
		fixtures = append(fixtures, fx)
	}
	sort.Slice(fixtures, func(i, j int) bool { return fixtures[i].Name < fixtures[j].Name })
	if len(fixtures) == 0 {
		t.Fatal("expected at least one compatibility fixture")
	}
	return fixtures
}

func applyCompatEnv(t *testing.T, vars map[string]string) {
	t.Helper()
	for k, v := range vars {
		t.Setenv(k, v)
	}
}

func mustParsePrecedence(t *testing.T, raw string) envsource.BindingPrecedence {
	t.Helper()
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "explicit_first":
		return envsource.ExplicitFirst
	case "inferred_first":
		return envsource.InferredFirst
	default:
		t.Fatalf("unknown env precedence %q", raw)
		return envsource.ExplicitFirst
	}
}

func compatSnapshot(out compatOutput) map[string]any {
	labels := map[string]any{}
	for k, v := range out.Labels {
		labels[k] = v
	}
	return map[string]any{
		"server": map[string]any{
			"host": out.Server.Host,
			"port": out.Server.Port,
			"tls":  out.Server.TLS,
		},
		"feature": map[string]any{
			"enabled": out.Feature.Enabled,
			"mode":    out.Feature.Mode,
		},
		"timeouts": map[string]any{
			"request": out.Timeouts.Request.String(),
		},
		"labels": labels,
	}
}

func assertCompatNoDiff(t *testing.T, name string, want, got map[string]any) {
	t.Helper()
	wantJSON := canonicalJSON(want)
	gotJSON := canonicalJSON(got)
	if wantJSON != gotJSON {
		t.Fatalf("compat parity mismatch for %s\nwant:\n%s\n\ngot:\n%s", name, wantJSON, gotJSON)
	}
}

func assertCompatTrace(t *testing.T, name string, expected map[string]string, trace *config.Trace) {
	t.Helper()
	for key, wantSource := range expected {
		got, ok := trace.Keys[key]
		if !ok {
			t.Fatalf("compat trace missing key %q for %s", key, name)
		}
		if got.FinalSource != wantSource {
			t.Fatalf("compat trace source mismatch for %s key %q: want %q got %q", name, key, wantSource, got.FinalSource)
		}
	}
}

func canonicalJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("marshal-error: %v", err)
	}
	return string(b)
}
