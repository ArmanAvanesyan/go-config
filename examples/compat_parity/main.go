// compat_parity demonstrates a fixture-driven compatibility check using Spec + Trace.
// It mirrors the shape of Phase 4 contract fixtures in a runnable example.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/ArmanAvanesyan/go-config/config"
	envsource "github.com/ArmanAvanesyan/go-config/providers/source/env"
)

type fixture struct {
	Name          string            `json:"name"`
	Strict        bool              `json:"strict"`
	Trees         []fixtureTree     `json:"trees"`
	Env           fixtureEnv        `json:"env"`
	Expected      map[string]any    `json:"expected"`
	ExpectedTrace map[string]string `json:"expected_trace"`
}

type fixtureTree struct {
	Priority int            `json:"priority"`
	Tree     map[string]any `json:"tree"`
}

type fixtureEnv struct {
	Priority   int                 `json:"priority"`
	Prefix     string              `json:"prefix"`
	Infer      bool                `json:"infer"`
	Precedence string              `json:"precedence"`
	Bindings   map[string][]string `json:"bindings"`
	Vars       map[string]string   `json:"vars"`
}

type appConfig struct {
	Server struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	} `json:"server"`
	Feature struct {
		Enabled bool   `json:"enabled"`
		Mode    string `json:"mode"`
	} `json:"feature"`
}

func main() {
	ctx := context.Background()

	raw, err := os.ReadFile("examples/compat_parity/fixture.json")
	if err != nil {
		log.Fatal(err)
	}

	var fx fixture
	if err := json.Unmarshal(raw, &fx); err != nil {
		log.Fatal(err)
	}

	for k, v := range fx.Env.Vars {
		if err := os.Setenv(k, v); err != nil {
			log.Fatal(err)
		}
	}

	trace := &config.Trace{}
	spec := config.Spec{
		Strict: fx.Strict,
		Trace:  trace,
	}
	for _, t := range fx.Trees {
		spec.Trees = append(spec.Trees, config.TreeSpec{
			Tree:     t.Tree,
			Priority: t.Priority,
		})
	}
	spec.Sources = append(spec.Sources, config.SourceSpec{
		Source: envsource.NewWithOptions(envsource.Options{
			Prefix:     fx.Env.Prefix,
			Infer:      fx.Env.Infer,
			Precedence: parsePrecedence(fx.Env.Precedence),
			Bindings:   fx.Env.Bindings,
		}),
		Meta: &config.SourceMeta{Priority: fx.Env.Priority, Required: true},
	})

	got, err := config.LoadTypedWithSpec[appConfig](ctx, spec)
	if err != nil {
		log.Fatal(err)
	}

	snapshot := map[string]any{
		"server": map[string]any{
			"host": got.Server.Host,
			"port": got.Server.Port,
		},
		"feature": map[string]any{
			"enabled": got.Feature.Enabled,
			"mode":    got.Feature.Mode,
		},
	}

	wantJSON, _ := json.MarshalIndent(fx.Expected, "", "  ")
	gotJSON, _ := json.MarshalIndent(snapshot, "", "  ")
	if string(wantJSON) != string(gotJSON) {
		log.Fatalf("parity mismatch\nwant:\n%s\n\ngot:\n%s", wantJSON, gotJSON)
	}

	fmt.Printf("compat parity passed for fixture: %s\n", fx.Name)
	for key, wantSource := range fx.ExpectedTrace {
		kt, ok := trace.Keys[key]
		if !ok {
			log.Fatalf("missing trace key: %s", key)
		}
		if kt.FinalSource != wantSource {
			log.Fatalf("trace mismatch for %s: want %s got %s", key, wantSource, kt.FinalSource)
		}
		fmt.Printf("trace %s -> %s\n", key, kt.FinalSource)
	}
}

func parsePrecedence(v string) envsource.BindingPrecedence {
	if v == "inferred_first" {
		return envsource.InferredFirst
	}
	return envsource.ExplicitFirst
}
