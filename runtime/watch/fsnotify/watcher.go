// Package fsnotify provides a file-system watcher for go-config that
// triggers a config reload whenever a watched file changes. On Linux and
// macOS it uses OS-level file events (inotify/kqueue) via golang.org/x/sys;
// on Windows and other platforms it falls back to stdlib-only polling.
package fsnotify

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"
)

// DefaultDebounce is the delay before firing onReload after a file event.
const DefaultDebounce = 150 * time.Millisecond

// backend starts watching paths and invokes onReload (debounced) on change.
// start returns a stop function and a channel that closes when the watcher has exited.
type backend interface {
	start(ctx context.Context, paths []string, debounce time.Duration, onReload func()) (stop func() error, done <-chan struct{}, err error)
}

// Watcher watches one or more files for changes and calls the registered
// callback when a change is detected. Implements config.ReloadTrigger.
type Watcher struct {
	paths    []string
	debounce time.Duration

	mu      sync.Mutex
	stop    func() error
	done    <-chan struct{}
	stopped bool
}

// New creates a watcher for the given file paths. Paths should match those
// used for file sources (e.g. the same path passed to providers/source/file.New).
// Start will return an error if any path cannot be watched (e.g. missing).
func New(paths ...string) *Watcher {
	return &Watcher{
		paths:    append([]string(nil), paths...),
		debounce: DefaultDebounce,
	}
}

// WithDebounce sets the delay between the last file event and calling onReload.
// Default is DefaultDebounce.
func (w *Watcher) WithDebounce(d time.Duration) *Watcher {
	w.debounce = d
	return w
}

// Start implements config.ReloadTrigger. It validates paths, starts the
// platform-appropriate backend (OS events or polling), and calls onReload
// after file changes (debounced). Returns immediately; use ctx to stop via cancellation.
func (w *Watcher) Start(ctx context.Context, onReload func()) error {
	w.mu.Lock()
	if w.stop != nil {
		w.mu.Unlock()
		return nil
	}
	if err := validatePaths(w.paths); err != nil {
		w.mu.Unlock()
		return err
	}
	debounce := w.debounce
	if debounce <= 0 {
		debounce = DefaultDebounce
	}
	b := newBackend()
	stop, done, err := b.start(ctx, w.paths, debounce, onReload)
	if err != nil {
		w.mu.Unlock()
		return err
	}
	w.stop = stop
	w.done = done
	w.stopped = false
	w.mu.Unlock()
	return nil
}

// Stop implements config.ReloadTrigger. It stops the watcher and releases
// resources. Idempotent.
func (w *Watcher) Stop() error {
	w.mu.Lock()
	if w.stopped || w.stop == nil {
		w.mu.Unlock()
		return nil
	}
	w.stopped = true
	stop := w.stop
	done := w.done
	w.stop = nil
	w.done = nil
	w.mu.Unlock()
	err := stop()
	if done != nil {
		<-done
	}
	return err
}

// validatePaths ensures each path exists and is a regular file.
func validatePaths(paths []string) error {
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return &os.PathError{Op: "watch", Path: p, Err: errors.New("not a regular file")}
		}
	}
	return nil
}
