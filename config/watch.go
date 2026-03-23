package config

import (
	"context"
	"sync"

	"github.com/ArmanAvanesyan/go-config/runtime/diff"
)

// ReloadTrigger signals when the loader should re-run the pipeline.
// Start begins watching and calls onReload whenever a reload should occur;
// it must return quickly (e.g. start watching in a goroutine).
// Stop releases resources and is idempotent.
// Implementations: runtime/watch/fsnotify, runtime/watch/polling.
type ReloadTrigger interface {
	Start(ctx context.Context, onReload func()) error
	Stop() error
}

// WatchTyped runs the pipeline once, calls callback with (nil, initial, nil),
// then starts the trigger. On each reload it decodes into a new T, computes
// diff.Changes(prevTree, newTree), and calls callback(prev, new, changes).
// Snapshots are immutable; the callback receives distinct old/new values.
// Blocks until ctx is canceled, then stops the trigger and returns.
func WatchTyped[T any](ctx context.Context, l *Loader, trigger ReloadTrigger, callback func(old, new *T, changes []diff.Change)) error {
	if trigger == nil || l == nil {
		return ErrNilTarget
	}
	if len(l.sources) == 0 {
		return ErrNoSources
	}
	if l.options.Decoder == nil {
		return ErrDecoderRequired
	}

	var initial T
	tree, err := l.loadTreeAndDecode(ctx, &initial)
	if err != nil {
		return err
	}
	callback(nil, &initial, nil)

	var prevSnapshot T
	prevSnapshot = initial
	prevTree := tree
	var mu sync.Mutex

	onReload := func() {
		mu.Lock()
		defer mu.Unlock()
		var next T
		newTree, err := l.loadTreeAndDecode(ctx, &next)
		if err != nil {
			return
		}
		changes := diff.Changes(prevTree, newTree)
		callback(&prevSnapshot, &next, changes)
		prevSnapshot = next
		prevTree = newTree
	}

	if err := trigger.Start(ctx, onReload); err != nil {
		return err
	}
	<-ctx.Done()
	_ = trigger.Stop()
	return ctx.Err()
}
