package config

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/ArmanAvanesyan/go-config/runtime/diff"
)

// safeMemSource is a thread-safe in-memory source for watch tests; Read returns a copy of the tree.
type safeMemSource struct {
	mu   sync.RWMutex
	tree map[string]any
}

func (s *safeMemSource) Read(context.Context) (any, error) {
	s.mu.RLock()
	tree := cloneTree(s.tree)
	s.mu.RUnlock()
	return &TreeDocument{Name: "safe", Tree: tree}, nil
}

func (s *safeMemSource) setTree(tree map[string]any) {
	s.mu.Lock()
	s.tree = tree
	s.mu.Unlock()
}

func cloneTree(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		if nested, ok := v.(map[string]any); ok {
			out[k] = cloneTree(nested)
		} else {
			out[k] = v
		}
	}
	return out
}

// mockTrigger calls onReload once after Start, then blocks until ctx is done.
type mockTrigger struct {
	onReload func()
	stop     func()
	startErr error
}

func (m *mockTrigger) Start(ctx context.Context, onReload func()) error {
	m.onReload = onReload
	if m.startErr != nil {
		return m.startErr
	}
	go func() {
		select {
		case <-time.After(10 * time.Millisecond):
			onReload()
		case <-ctx.Done():
			return
		}
	}()
	return nil
}

func (m *mockTrigger) Stop() error {
	if m.stop != nil {
		m.stop()
	}
	return nil
}

func TestWatchTyped_InitialAndReload(t *testing.T) {
	t.Parallel()

	type Cfg struct {
		Name string `json:"name"`
	}

	source := &safeMemSource{tree: map[string]any{"name": "v1"}}
	loader := New().AddSource(source)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var calls int
	var mu sync.Mutex

	errCh := make(chan error, 1)
	go func() {
		errCh <- WatchTyped[Cfg](ctx, loader, &mockTrigger{}, func(old, new *Cfg, changes []diff.Change) {
			mu.Lock()
			calls++
			mu.Unlock()
			if old == nil && new != nil {
				if new.Name != "v1" {
					t.Errorf("initial name: got %q", new.Name)
				}
				return
			}
			if old != nil && new != nil {
				if old.Name != "v1" || new.Name != "v2" {
					t.Errorf("reload: old.Name=%q new.Name=%q", old.Name, new.Name)
				}
			}
		})
	}()

	// Change source before mock fires onReload at 10ms
	time.Sleep(5 * time.Millisecond)
	source.setTree(map[string]any{"name": "v2"})
	time.Sleep(50 * time.Millisecond)
	cancel()

	err := <-errCh
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	mu.Lock()
	n := calls
	mu.Unlock()
	if n < 2 {
		t.Errorf("expected at least 2 callbacks (initial + reload), got %d", n)
	}
}

func TestWatchTyped_NilTrigger(t *testing.T) {
	t.Parallel()
	loader := New().AddSource(&testMemSource{tree: map[string]any{}})
	err := WatchTyped[struct{ X int }](context.Background(), loader, nil, func(_, _ *struct{ X int }, _ []diff.Change) {})
	if err != ErrNilTarget {
		t.Fatalf("expected ErrNilTarget, got %v", err)
	}
}

func TestWatchTyped_NilLoader(t *testing.T) {
	t.Parallel()
	err := WatchTyped[struct{ X int }](context.Background(), nil, &mockTrigger{}, func(_, _ *struct{ X int }, _ []diff.Change) {})
	if err != ErrNilTarget {
		t.Fatalf("expected ErrNilTarget, got %v", err)
	}
}

func TestWatchTyped_NoSources(t *testing.T) {
	t.Parallel()
	loader := New()
	err := WatchTyped[struct{ X int }](context.Background(), loader, &mockTrigger{}, func(_, _ *struct{ X int }, _ []diff.Change) {})
	if err != ErrNoSources {
		t.Fatalf("expected ErrNoSources, got %v", err)
	}
}

func TestWatchTyped_BranchErrorPaths(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	loaderNoDecoder := New(WithDecoder(nil)).AddSource(&testMemSource{tree: map[string]any{"v": "x"}})
	if err := WatchTyped[struct {
		V string `json:"v"`
	}](ctx, loaderNoDecoder, &mockTrigger{}, func(_, _ *struct {
		V string `json:"v"`
	}, _ []diff.Change) {
	}); !errors.Is(err, ErrDecoderRequired) {
		t.Fatalf("expected ErrDecoderRequired, got %v", err)
	}

	startErr := errors.New("start failed")
	loader := New().AddSource(&testMemSource{tree: map[string]any{"v": "x"}})
	if err := WatchTyped[struct {
		V string `json:"v"`
	}](ctx, loader, &mockTrigger{startErr: startErr}, func(_, _ *struct {
		V string `json:"v"`
	}, _ []diff.Change) {
	}); !errors.Is(err, startErr) {
		t.Fatalf("expected start error, got %v", err)
	}

	loaderBadInitial := New().AddSource(&testMemSource{err: errors.New("boom")})
	if err := WatchTyped[struct {
		V string `json:"v"`
	}](ctx, loaderBadInitial, &mockTrigger{}, func(_, _ *struct {
		V string `json:"v"`
	}, _ []diff.Change) {
	}); !errors.Is(err, ErrSourceReadFailed) {
		t.Fatalf("expected initial load source read error, got %v", err)
	}
}

func TestWatchTyped_ReloadDecodeErrorIsSwallowed(t *testing.T) {
	t.Parallel()
	type C struct {
		V string `json:"v"`
	}
	src := &safeMemSource{tree: map[string]any{"v": "ok"}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	calls := 0
	done := make(chan error, 1)
	tr := &mockTrigger{
		stop: cancel,
	}
	go func() {
		done <- WatchTyped[C](ctx, New().AddSource(src), tr, func(old, new *C, _ []diff.Change) {
			_ = old
			_ = new
			calls++
		})
	}()

	time.Sleep(20 * time.Millisecond)
	src.setTree(map[string]any{"v": map[string]any{"bad": 1}})
	if tr.onReload != nil {
		tr.onReload() // should fail decode and be swallowed
	}
	time.Sleep(20 * time.Millisecond)
	src.setTree(map[string]any{"v": "ok2"})
	if tr.onReload != nil {
		tr.onReload()
	}
	time.Sleep(20 * time.Millisecond)
	cancel()
	<-done
	if calls < 2 {
		t.Fatalf("expected initial + successful reload callback, got %d", calls)
	}
}
