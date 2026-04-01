package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/ArmanAvanesyan/go-config/providers/decoder/mapstructure"
)

func TestLoader_ErrorCases(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name    string
		setup   func() *Loader
		target  any
		wantErr error
	}{
		{
			name:    "nil target",
			setup:   func() *Loader { return New().AddSource(&testMemSource{tree: map[string]any{}}) },
			target:  nil,
			wantErr: ErrNilTarget,
		},
		{
			name:    "no sources",
			setup:   func() *Loader { return New() },
			target:  &struct{}{},
			wantErr: ErrNoSources,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			err := tc.setup().Load(context.Background(), tc.target)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected %v, got %v", tc.wantErr, err)
			}
		})
	}
}

type testMemSource struct {
	tree map[string]any
	err  error
}

func (s *testMemSource) Read(context.Context) (any, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &TreeDocument{
		Name: "test",
		Tree: s.tree,
	}, nil
}

func TestLoader_SingleMemorySource(t *testing.T) {
	t.Parallel()

	type AppConfig struct {
		Name string `json:"name"`
	}

	tree := map[string]any{"name": "demo"}

	var cfg AppConfig
	err := New().AddSource(&testMemSource{tree: tree}).Load(context.Background(), &cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Name != "demo" {
		t.Fatalf("expected name demo, got %q", cfg.Name)
	}
}

func TestLoader_LoadTyped(t *testing.T) {
	t.Parallel()

	type AppConfig struct {
		Name string `json:"name"`
	}

	tree := map[string]any{"name": "typed-demo"}
	loader := New().AddSource(&testMemSource{tree: tree})
	cfg, err := LoadTyped[AppConfig](context.Background(), loader)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Name != "typed-demo" {
		t.Fatalf("expected name typed-demo, got %q", cfg.Name)
	}
}

func TestLoader_MultipleMemorySources_DeepMerge(t *testing.T) {
	t.Parallel()

	type Server struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}
	type Cfg struct {
		Server Server `json:"server"`
	}

	cases := []struct {
		name    string
		sources []map[string]any
		want    Cfg
	}{
		{
			name: "second source adds key",
			sources: []map[string]any{
				{"server": map[string]any{"host": "localhost"}},
				{"server": map[string]any{"port": float64(9090)}},
			},
			want: Cfg{Server: Server{Host: "localhost", Port: 9090}},
		},
		{
			name: "second source overrides key",
			sources: []map[string]any{
				{"server": map[string]any{"host": "old"}},
				{"server": map[string]any{"host": "new"}},
			},
			want: Cfg{Server: Server{Host: "new"}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			loader := New()
			for _, tree := range tc.sources {
				loader = loader.AddSource(&testMemSource{tree: tree})
			}

			var cfg Cfg
			if err := loader.Load(context.Background(), &cfg); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg != tc.want {
				t.Fatalf("want %+v, got %+v", tc.want, cfg)
			}
		})
	}
}

func TestLoader_SourceMeta_Priority(t *testing.T) {
	t.Parallel()

	type Cfg struct {
		Key string `json:"key"`
	}

	// Add high priority first, then low; after sort by priority ascending, low is merged first, high wins.
	loader := New().
		AddSourceWithMeta(&testMemSource{tree: map[string]any{"key": "high"}}, nil, &SourceMeta{Priority: 10}).
		AddSourceWithMeta(&testMemSource{tree: map[string]any{"key": "low"}}, nil, &SourceMeta{Priority: 0})

	var cfg Cfg
	err := loader.Load(context.Background(), &cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Key != "high" {
		t.Fatalf("expected key=high (priority wins), got %q", cfg.Key)
	}
}

func TestLoader_SourceMeta_OptionalSkipsFailure(t *testing.T) {
	t.Parallel()

	type Cfg struct {
		Key string `json:"key"`
	}

	failingSource := &testMemSource{err: errors.New("read failed")}

	loader := New().
		AddSource(&testMemSource{tree: map[string]any{"key": "ok"}}).
		AddSourceWithMeta(failingSource, nil, &SourceMeta{Required: false})

	var cfg Cfg
	err := loader.Load(context.Background(), &cfg)
	if err != nil {
		t.Fatalf("optional source failure should be skipped, got: %v", err)
	}
	if cfg.Key != "ok" {
		t.Fatalf("expected key=ok, got %q", cfg.Key)
	}
}

func TestLoader_WithStrict_DoesNotOverrideCustomDecoder(t *testing.T) {
	t.Parallel()

	type Cfg struct {
		Name string `json:"name"`
	}

	var cfg Cfg
	err := New(
		WithStrict(true),
		WithDecoder(mapstructure.New()),
	).AddSource(&testMemSource{
		tree: map[string]any{
			"name":  "demo",
			"extra": "ignored-by-mapstructure",
		},
	}).Load(context.Background(), &cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Name != "demo" {
		t.Fatalf("expected name demo, got %q", cfg.Name)
	}
}

type testDocSource struct {
	doc *Document
	err error
}

func (s *testDocSource) Read(context.Context) (any, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.doc, nil
}

type testTypedParser struct{}

func (p *testTypedParser) Parse(context.Context, *Document) (map[string]any, error) {
	return nil, fmt.Errorf("unexpected Parse call")
}

type orderingValidator struct {
	order *[]string
}

func (v *orderingValidator) Validate(context.Context, any) error {
	*v.order = append(*v.order, "validator-interface")
	return nil
}

type orderingConfig struct {
	Name  string `json:"name"`
	steps *[]string
}

func (c *orderingConfig) ApplyDefaults() {
	*c.steps = append(*c.steps, "defaults-interface")
}

func (p *testTypedParser) ParseTyped(_ context.Context, _ *Document, out any) error {
	cfg, ok := out.(*struct {
		Name string `json:"name"`
	})
	if !ok {
		return fmt.Errorf("unexpected output type %T", out)
	}
	cfg.Name = "typed-fast-path"
	return nil
}

func TestLoader_WithDirectDecode_FastPath(t *testing.T) {
	t.Parallel()
	var cfg struct {
		Name string `json:"name"`
	}
	loader := New(WithDirectDecode(true)).
		AddSource(&testDocSource{
			doc: &Document{Name: "test.yaml", Format: "yaml", Raw: []byte("name: demo")},
		}, &testTypedParser{})
	if err := loader.Load(context.Background(), &cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Name != "typed-fast-path" {
		t.Fatalf("expected typed-fast-path, got %q", cfg.Name)
	}
}

func TestLoader_DirectDecodeConstraints(t *testing.T) {
	t.Parallel()
	tp := &typedParserStub{}
	l := New(WithDirectDecode(true)).AddSource(&docSrc{v: &Document{Name: "a", Raw: []byte("v: 1")}}, tp)
	var out struct {
		V string `json:"v"`
	}
	if err := l.Load(context.Background(), &out); err != nil {
		t.Fatalf("fast path should succeed: %v", err)
	}
	if !tp.parseTypedCalled || out.V != "typed" {
		t.Fatalf("typed parser should be used")
	}

	// resolver disables direct decode and falls back to Parse path.
	tp2 := &typedParserStub{}
	l2 := New(WithDirectDecode(true), WithResolver(&noOpResolver{})).AddSource(&docSrc{v: &Document{Raw: []byte("x")}}, tp2)
	var out2 struct {
		V string `json:"v"`
	}
	if err := l2.Load(context.Background(), &out2); err == nil {
		t.Fatal("expected fallback path parse error when resolver present")
	}
}

func TestLoader_LifecycleHookOrdering(t *testing.T) {
	t.Parallel()
	steps := []string{}
	loader := New(
		WithDefaultsFunc(func(context.Context, any) error {
			steps = append(steps, "defaults-callback")
			return nil
		}),
		WithValidateFunc(func(context.Context, any) error {
			steps = append(steps, "validate-callback")
			return nil
		}),
		WithValidator(&orderingValidator{order: &steps}),
	).AddSource(&testMemSource{tree: map[string]any{"name": "demo"}})

	cfg := &orderingConfig{steps: &steps}
	if err := loader.Load(context.Background(), cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"defaults-interface", "defaults-callback", "validate-callback", "validator-interface"}
	if !reflect.DeepEqual(want, steps) {
		t.Fatalf("unexpected order, want=%v got=%v", want, steps)
	}
}

type parseErrorParser struct{}

func (p *parseErrorParser) Parse(context.Context, *Document) (map[string]any, error) {
	return nil, fmt.Errorf("parse boom")
}

type notFoundSource struct{}

func (s *notFoundSource) Read(context.Context) (any, error) {
	return nil, &os.PathError{Op: "open", Path: "missing.yaml", Err: os.ErrNotExist}
}

func TestLoader_SourceMeta_ParsePolicyIgnore(t *testing.T) {
	t.Parallel()
	type Cfg struct {
		Key string `json:"key"`
	}

	loader := New().
		AddSource(&testMemSource{tree: map[string]any{"key": "ok"}}).
		AddSourceWithMeta(
			&testDocSource{doc: &Document{Name: "broken.yaml", Format: "yaml", Raw: []byte(":::")}},
			[]Parser{&parseErrorParser{}},
			&SourceMeta{Required: true, ParsePolicy: ParsePolicyIgnore},
		)

	var cfg Cfg
	if err := loader.Load(context.Background(), &cfg); err != nil {
		t.Fatalf("expected parse policy ignore to continue, got %v", err)
	}
	if cfg.Key != "ok" {
		t.Fatalf("expected key=ok, got %q", cfg.Key)
	}
}

func TestLoader_SourceMeta_MissingPolicyIgnore(t *testing.T) {
	t.Parallel()
	type Cfg struct {
		Key string `json:"key"`
	}
	loader := New().
		AddSource(&testMemSource{tree: map[string]any{"key": "ok"}}).
		AddSourceWithMeta(&notFoundSource{}, nil, &SourceMeta{Required: true, MissingPolicy: MissingPolicyIgnore})

	var cfg Cfg
	if err := loader.Load(context.Background(), &cfg); err != nil {
		t.Fatalf("expected missing policy ignore to continue, got %v", err)
	}
	if cfg.Key != "ok" {
		t.Fatalf("expected key=ok, got %q", cfg.Key)
	}
}
