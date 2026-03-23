//go:build linux

package fsnotify

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestOSEventsBackend_Linux_StartStop(t *testing.T) {
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
	case <-time.After(5 * time.Second):
		t.Fatal("done channel not closed after stop")
	}
}

func TestOSEventsBackend_Linux_DetectsChange(t *testing.T) {
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
		t.Errorf("expected at least 1 reload after write, got %d", n)
	}
}

func TestOSEventsBackend_Linux_InvalidPath(t *testing.T) {
	t.Parallel()
	b := newBackend()
	ctx := context.Background()
	dir, err := os.MkdirTemp("", "os_events_invalid")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(dir) }()

	_, _, err = b.start(ctx, []string{filepath.Join(dir, "nonexistent")}, time.Second, func() {})
	if err == nil {
		t.Fatal("expected error for missing path")
	}
}
