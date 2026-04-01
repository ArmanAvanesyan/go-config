package config

import (
	"context"
	"strings"
	"testing"
)

func TestLoadWithSpec_ParityWithManualLoader(t *testing.T) {
	t.Parallel()
	type cfg struct {
		Name string `json:"name"`
	}

	manual := New().
		AddSource(&testMemSource{tree: map[string]any{"name": "manual"}})
	var manualCfg cfg
	if err := manual.Load(context.Background(), &manualCfg); err != nil {
		t.Fatalf("manual load failed: %v", err)
	}

	var specCfg cfg
	err := LoadWithSpec(context.Background(), &specCfg, Spec{
		Trees: []TreeSpec{{Tree: map[string]any{"name": "manual"}, Priority: 0}},
	})
	if err != nil {
		t.Fatalf("spec load failed: %v", err)
	}
	if manualCfg != specCfg {
		t.Fatalf("expected parity, manual=%+v spec=%+v", manualCfg, specCfg)
	}
}

func TestLoadTypedWithSpec(t *testing.T) {
	t.Parallel()
	type cfg struct {
		Name string `json:"name"`
	}
	got, err := LoadTypedWithSpec[cfg](context.Background(), Spec{
		Trees: []TreeSpec{{Tree: map[string]any{"name": "typed"}}},
	})
	if err != nil {
		t.Fatalf("typed spec load failed: %v", err)
	}
	if got.Name != "typed" {
		t.Fatalf("expected typed, got %q", got.Name)
	}
}

func TestLoadWithSpec_TraceCapturesSourceAndHooks(t *testing.T) {
	t.Parallel()
	type cfg struct {
		Name string `json:"name"`
	}
	tr := &Trace{}
	var c cfg
	err := LoadWithSpec(context.Background(), &c, Spec{
		Trace: tr,
		Trees: []TreeSpec{
			{Tree: map[string]any{"name": "first"}, Priority: 0},
			{Tree: map[string]any{"name": "second"}, Priority: 10},
		},
		DefaultsFn: func(context.Context, any) error { return nil },
		ValidateFn: func(context.Context, any) error { return nil },
	})
	if err != nil {
		t.Fatalf("spec trace load failed: %v", err)
	}
	kt, ok := tr.Keys["name"]
	if !ok {
		t.Fatalf("expected trace key for name, got %#v", tr.Keys)
	}
	if kt.FinalValue != "second" {
		t.Fatalf("expected second winner, got %#v", kt.FinalValue)
	}
	if len(kt.Candidates) < 2 {
		t.Fatalf("expected override candidates, got %#v", kt.Candidates)
	}
	if len(tr.HookOrder) < 2 || tr.HookOrder[0] != "defaults-callback" || tr.HookOrder[1] != "validate-callback" {
		t.Fatalf("unexpected hook order: %#v", tr.HookOrder)
	}
}

func TestLoadWithSpec_ValidationErrorContract(t *testing.T) {
	t.Parallel()
	type cfg struct {
		Name string `json:"name"`
	}
	var c cfg
	err := LoadWithSpec(context.Background(), &c, Spec{
		Trees: []TreeSpec{{Tree: map[string]any{"name": "x"}}},
		ValidateFn: func(context.Context, any) error {
			return errString("bad callback")
		},
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "validate-callback") {
		t.Fatalf("expected callback stage marker, got: %v", err)
	}
}

type errString string

func (e errString) Error() string { return string(e) }
