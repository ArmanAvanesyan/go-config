package polling

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestWatcher_StartStop(t *testing.T) {
	t.Parallel()
	w := New(20 * time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var reloads int
	var mu sync.Mutex
	err := w.Start(ctx, func() {
		mu.Lock()
		reloads++
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("Start: %v", err)
	}

	time.Sleep(80 * time.Millisecond)
	cancel()
	_ = w.Stop()

	mu.Lock()
	n := reloads
	mu.Unlock()
	if n < 2 {
		t.Errorf("expected at least 2 reloads in 80ms with 20ms interval, got %d", n)
	}
}

func TestWatcher_StopIdempotent(t *testing.T) {
	t.Parallel()
	w := New(time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	_ = w.Start(ctx, func() {})
	cancel()
	_ = w.Stop()
	_ = w.Stop()
	_ = w.Stop()
}

func TestWatcher_ContextCancelStopsTicker(t *testing.T) {
	t.Parallel()
	w := New(10 * time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())

	var reloads int
	var mu sync.Mutex
	_ = w.Start(ctx, func() {
		mu.Lock()
		reloads++
		mu.Unlock()
	})

	time.Sleep(25 * time.Millisecond)
	cancel()
	time.Sleep(30 * time.Millisecond)

	mu.Lock()
	beforeStop := reloads
	mu.Unlock()

	_ = w.Stop()
	time.Sleep(30 * time.Millisecond)

	mu.Lock()
	afterStop := reloads
	mu.Unlock()

	if afterStop != beforeStop {
		t.Errorf("reloads should not increase after cancel: before stop %d, after %d", beforeStop, afterStop)
	}
}
