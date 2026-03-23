// Package polling provides a polling-based watcher for go-config that
// periodically triggers a reload on a fixed interval. The loader re-runs
// the pipeline and the callback receives (old, new) snapshots; no change
// detection is done inside the watcher.
package polling

import (
	"context"
	"sync"
	"time"
)

// Watcher calls onReload on a fixed interval. Implements config.ReloadTrigger.
type Watcher struct {
	interval time.Duration

	mu      sync.Mutex
	ticker  *time.Ticker
	stopped bool
	done    chan struct{}
}

// New creates a watcher that calls onReload every interval.
func New(interval time.Duration) *Watcher {
	return &Watcher{interval: interval}
}

// Start implements config.ReloadTrigger. It starts a goroutine that calls
// onReload every interval. Returns immediately.
func (w *Watcher) Start(ctx context.Context, onReload func()) error {
	w.mu.Lock()
	if w.ticker != nil {
		w.mu.Unlock()
		return nil
	}
	interval := w.interval
	if interval <= 0 {
		interval = time.Second
	}
	w.ticker = time.NewTicker(interval)
	w.stopped = false
	w.done = make(chan struct{})
	w.mu.Unlock()

	go w.run(ctx, onReload)
	return nil
}

// Stop implements config.ReloadTrigger. It stops the ticker and releases
// resources. Idempotent. Call after context is canceled so the run goroutine exits.
func (w *Watcher) Stop() error {
	w.mu.Lock()
	if w.stopped || w.ticker == nil {
		w.mu.Unlock()
		return nil
	}
	w.stopped = true
	ticker := w.ticker
	w.ticker = nil
	w.mu.Unlock()
	ticker.Stop()
	<-w.done
	return nil
}

func (w *Watcher) run(ctx context.Context, onReload func()) {
	defer close(w.done)
	w.mu.Lock()
	ticker := w.ticker
	w.mu.Unlock()
	if ticker == nil {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			onReload()
		}
	}
}
