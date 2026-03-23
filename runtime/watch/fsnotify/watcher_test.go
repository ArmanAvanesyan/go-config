package fsnotify

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestWatcher_StartStop(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("x: 1"), 0644); err != nil {
		t.Fatal(err)
	}

	w := New(path)
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

	time.Sleep(20 * time.Millisecond)
	if err := os.WriteFile(path, []byte("x: 2"), 0644); err != nil {
		t.Fatal(err)
	}
	// Poll for reload; on polling backend (e.g. Windows) may need poll interval + debounce.
	// Use a generous timeout to avoid flakes under load or slow filesystems.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := reloads
		mu.Unlock()
		if n >= 1 {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	mu.Lock()
	n := reloads
	mu.Unlock()
	if n < 1 {
		t.Errorf("expected at least 1 reload within 2s after write, got %d", n)
	}

	cancel()
	_ = w.Stop()
	_ = w.Stop()
}

func TestWatcher_MissingPath(t *testing.T) {
	t.Parallel()
	w := New(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	ctx := context.Background()
	err := w.Start(ctx, func() {})
	if err == nil {
		t.Fatal("expected error when path does not exist")
	}
}

func TestWatcher_WithDebounce(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "c.yaml")
	if err := os.WriteFile(path, []byte("a: 1"), 0644); err != nil {
		t.Fatal(err)
	}

	w := New(path).WithDebounce(50 * time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var reloads int
	var mu sync.Mutex
	if err := w.Start(ctx, func() {
		mu.Lock()
		reloads++
		mu.Unlock()
	}); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte("a: 2"), 0644); err != nil {
		t.Fatal(err)
	}
	// Poll for reload; polling backend can take poll interval + debounce (e.g. 50ms+50ms on Windows).
	// Use a generous timeout to reduce flakes in CI under load.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := reloads
		mu.Unlock()
		if n >= 1 {
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	mu.Lock()
	n := reloads
	mu.Unlock()
	if n < 1 {
		t.Errorf("expected at least 1 reload within timeout, got %d", n)
	}
	cancel()
	_ = w.Stop()
}
