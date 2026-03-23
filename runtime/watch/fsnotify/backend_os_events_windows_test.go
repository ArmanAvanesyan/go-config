//go:build windows && ignore

// Tests for the Windows OS-events backend. Only compiled with -tags ignore
// (same as the backend, which is currently disabled due to sync ReadDirectoryChangesW deadlock).

package fsnotify

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestOSEventsBackend_Windows_StartStop(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("x: 1"), 0644); err != nil {
		t.Fatal(err)
	}

	b := newBackend()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop, done, err := b.start(ctx, []string{path}, 50*time.Millisecond, func() {})
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if stop == nil || done == nil {
		t.Fatal("start returned nil stop or done")
	}

	if err := stop(); err != nil {
		t.Errorf("stop: %v", err)
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("done channel not closed after stop")
	}
}

func TestOSEventsBackend_Windows_DetectsChange(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("x: 1"), 0644); err != nil {
		t.Fatal(err)
	}

	b := newBackend()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var reloads int
	var mu sync.Mutex
	stop, done, err := b.start(ctx, []string{path}, 50*time.Millisecond, func() {
		mu.Lock()
		reloads++
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	defer func() { _ = stop(); <-done }()

	time.Sleep(20 * time.Millisecond)
	if err := os.WriteFile(path, []byte("x: 2"), 0644); err != nil {
		t.Fatal(err)
	}
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	n := reloads
	mu.Unlock()
	if n < 1 {
		t.Errorf("expected at least 1 reload after write, got %d", n)
	}
}

func TestOSEventsBackend_Windows_InvalidPath(t *testing.T) {
	t.Parallel()
	b := newBackend()
	ctx := context.Background()
	_, _, err := b.start(ctx, []string{filepath.Join(t.TempDir(), "nonexistent")}, time.Second, func() {})
	if err == nil {
		t.Fatal("expected error for missing path")
	}
}
