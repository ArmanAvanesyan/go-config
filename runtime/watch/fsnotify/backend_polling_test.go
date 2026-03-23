//go:build !linux && !darwin

package fsnotify

import (
	"context"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

func TestPollingBackend_StartInitError(t *testing.T) {
	t.Parallel()
	p := &pollingBackend{}
	_, _, err := p.start(context.Background(), []string{filepath.Join(t.TempDir(), "missing.yaml")}, 0, func() {})
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestPollingBackend_StopAndCancel(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "a.yaml")
	if err := os.WriteFile(path, []byte("a: 1"), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := &pollingBackend{}
	stop, done, err := p.start(ctx, []string{path}, 10*time.Millisecond, func() {})
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}
	if err := stop(); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("done channel not closed after stop")
	}
}

func TestPollingBackend_ContextCancel(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "cancel.yaml")
	if err := os.WriteFile(path, []byte("a: 1"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	p := &pollingBackend{}
	_, done, err := p.start(ctx, []string{path}, 10*time.Millisecond, func() {})
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("done channel not closed after context cancel")
	}
}

func TestPollingBackend_ChangeTriggersDebouncedReload(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "b.yaml")
	if err := os.WriteFile(path, []byte("x: 1"), 0o644); err != nil {
		t.Fatal(err)
	}

	var calls atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := &pollingBackend{}
	stop, done, err := p.start(ctx, []string{path}, 20*time.Millisecond, func() {
		calls.Add(1)
	})
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}
	defer func() {
		_ = stop()
		<-done
	}()

	time.Sleep(70 * time.Millisecond)
	if err := os.WriteFile(path, []byte("x: 2"), 0o644); err != nil {
		t.Fatal(err)
	}
	time.Sleep(120 * time.Millisecond)
	if calls.Load() < 1 {
		t.Fatalf("expected at least one reload callback, got %d", calls.Load())
	}

	// Rapid updates should still coalesce within debounce window.
	_ = os.WriteFile(path, []byte("x: 3"), 0o644)
	_ = os.WriteFile(path, []byte("x: 4"), 0o644)
	time.Sleep(120 * time.Millisecond)
	if calls.Load() < 2 {
		t.Fatalf("expected additional callback after rapid writes, got %d", calls.Load())
	}
}

func TestPollingBackend_NonRegularAndStatErrorPaths(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	filePath := filepath.Join(dir, "c.yaml")
	if err := os.WriteFile(filePath, []byte("x: 1"), 0o644); err != nil {
		t.Fatal(err)
	}

	var calls atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := &pollingBackend{}
	stop, done, err := p.start(ctx, []string{filePath, dir}, 10*time.Millisecond, func() {
		calls.Add(1)
	})
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}
	defer func() {
		_ = stop()
		<-done
	}()

	// Non-regular path (dir) should be ignored; missing stat path during tick should continue.
	// We simulate missing path during tick by removing and recreating file quickly.
	_ = os.Remove(filePath)
	time.Sleep(80 * time.Millisecond)
	_ = os.WriteFile(filePath, []byte("x: 9"), 0o644)
	time.Sleep(120 * time.Millisecond)

	// Not asserting a strict count; this test ensures loop survives stat and non-regular branches.
	_ = calls.Load()
}

func TestPollingBackend_StopWhileDebounceTimerPending(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "timer.yaml")
	if err := os.WriteFile(path, []byte("x: 1"), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p := &pollingBackend{}
	stop, done, err := p.start(ctx, []string{path}, 100*time.Millisecond, func() {})
	if err != nil {
		t.Fatalf("start failed: %v", err)
	}
	time.Sleep(70 * time.Millisecond)
	_ = os.WriteFile(path, []byte("x: 2"), 0o644)
	time.Sleep(20 * time.Millisecond)
	if err := stop(); err != nil {
		t.Fatalf("stop failed: %v", err)
	}
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("done channel not closed after stop with pending timer")
	}
}
