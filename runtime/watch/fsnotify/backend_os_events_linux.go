//go:build linux

package fsnotify

import (
	"context"
	"encoding/binary"
	"errors"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

type osEventsBackend struct{}

func newBackend() backend {
	return &osEventsBackend{}
}

func (b *osEventsBackend) start(ctx context.Context, paths []string, debounce time.Duration, onReload func()) (func() error, <-chan struct{}, error) {
	// Child context is cancelled from stop() so the read loop can exit even when
	// the caller passed context.Background() (common in tests) and would otherwise
	// spin on EAGAIN until the fd closes.
	watchCtx, cancel := context.WithCancel(ctx)

	fd, err := unix.InotifyInit1(unix.IN_NONBLOCK | unix.IN_CLOEXEC)
	if err != nil {
		cancel()
		return nil, nil, err
	}
	for _, p := range paths {
		_, err := unix.InotifyAddWatch(fd, p, unix.IN_MODIFY|unix.IN_ATTRIB|unix.IN_CLOSE_WRITE|unix.IN_CREATE)
		if err != nil {
			_ = unix.Close(fd)
			cancel()
			return nil, nil, err
		}
	}
	var closeOnce sync.Once
	closeFD := func() error {
		var closeErr error
		closeOnce.Do(func() {
			closeErr = unix.Close(fd)
		})
		return closeErr
	}

	done := make(chan struct{})
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
		defer cancel()
		defer close(done)
		defer func() { _ = closeFD() }()
		buf := make([]byte, unix.SizeofInotifyEvent*128+unix.NAME_MAX+1)
		for {
			n, err := unix.Read(fd, buf)
			if err != nil {
				if errors.Is(err, unix.EAGAIN) {
					select {
					case <-watchCtx.Done():
						stopTimer()
						return
					default:
						time.Sleep(10 * time.Millisecond)
						continue
					}
				}
				stopTimer()
				return
			}
			for i := 0; i < n; {
				if i+unix.SizeofInotifyEvent > n {
					break
				}
				// struct inotify_event: wd, mask, cookie, len (name bytes follow; len includes padding).
				mask := binary.LittleEndian.Uint32(buf[i+4 : i+8])
				if mask&(unix.IN_MODIFY|unix.IN_ATTRIB|unix.IN_CLOSE_WRITE|unix.IN_CREATE) != 0 {
					scheduleReload()
				}
				nameLen := int(binary.LittleEndian.Uint32(buf[i+12 : i+16]))
				next := i + unix.SizeofInotifyEvent + nameLen
				if next > n || next < i {
					break
				}
				i = next
			}
		}
	}()

	stop := func() error {
		stopTimer()
		cancel()
		return closeFD()
	}
	return stop, done, nil
}
