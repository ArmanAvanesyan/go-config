//go:build !linux && !darwin

package fsnotify

import (
	"context"
	"os"
	"sync"
	"time"
)

// DefaultPollInterval is how often the polling backend checks for file changes.
const DefaultPollInterval = 50 * time.Millisecond

type pollingBackend struct{}

func newBackend() backend {
	return &pollingBackend{}
}

func (p *pollingBackend) start(ctx context.Context, paths []string, debounce time.Duration, onReload func()) (func() error, <-chan struct{}, error) {
	interval := DefaultPollInterval
	if interval <= 0 {
		interval = 50 * time.Millisecond
	}
	if debounce <= 0 {
		debounce = DefaultDebounce
	}

	state := make([]fileState, len(paths))
	for i, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, nil, err
		}
		state[i] = fileState{path: path, mtime: info.ModTime(), size: info.Size()}
	}

	ticker := time.NewTicker(interval)
	done := make(chan struct{})
	quit := make(chan struct{})
	var timerMu sync.Mutex
	var timer *time.Timer
	stopTimer := func() {
		timerMu.Lock()
		if timer != nil {
			timer.Stop()
			timer = nil
		}
		timerMu.Unlock()
	}
	scheduleReload := func() {
		timerMu.Lock()
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(debounce, func() {
			timerMu.Lock()
			timer = nil
			timerMu.Unlock()
			onReload()
		})
		timerMu.Unlock()
	}

	go func() {
		defer close(done)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				stopTimer()
				return
			case <-quit:
				stopTimer()
				return
			case <-ticker.C:
				changed := false
				for i, path := range paths {
					info, err := os.Stat(path)
					if err != nil {
						continue
					}
					if !info.Mode().IsRegular() {
						continue
					}
					if state[i].mtime != info.ModTime() || state[i].size != info.Size() {
						state[i].mtime = info.ModTime()
						state[i].size = info.Size()
						changed = true
					}
				}
				if changed {
					scheduleReload()
				}
			}
		}
	}()

	stop := func() error {
		ticker.Stop()
		close(quit)
		stopTimer()
		return nil
	}
	return stop, done, nil
}

type fileState struct {
	path  string
	mtime time.Time
	size  int64
}
