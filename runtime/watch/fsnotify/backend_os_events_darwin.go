//go:build darwin

package fsnotify

import (
	"context"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

type osEventsBackend struct{}

func newBackend() backend {
	return &osEventsBackend{}
}

func (b *osEventsBackend) start(ctx context.Context, paths []string, debounce time.Duration, onReload func()) (func() error, <-chan struct{}, error) {
	kq, err := unix.Kqueue()
	if err != nil {
		return nil, nil, err
	}
	var fds []int
	for _, p := range paths {
		fd, err := unix.Open(p, unix.O_RDONLY, 0)
		if err != nil {
			for _, c := range fds {
				_ = unix.Close(c)
			}
			_ = unix.Close(kq)
			return nil, nil, err
		}
		fds = append(fds, fd)
		ev := unix.Kevent_t{
			Ident:  uint64(fd),
			Filter: unix.EVFILT_VNODE,
			Flags:  unix.EV_ADD | unix.EV_CLEAR,
			Fflags: unix.NOTE_WRITE | unix.NOTE_ATTRIB,
		}
		_, err = unix.Kevent(kq, []unix.Kevent_t{ev}, nil, nil)
		if err != nil {
			for _, c := range fds {
				_ = unix.Close(c)
			}
			_ = unix.Close(kq)
			return nil, nil, err
		}
	}
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
		defer func() {
			for _, fd := range fds {
				_ = unix.Close(fd)
			}
			_ = unix.Close(kq)
		}()
		events := make([]unix.Kevent_t, len(paths)+8)
		for {
			timeout := unix.Timespec{Sec: 1, Nsec: 0}
			n, err := unix.Kevent(kq, nil, events, &timeout)
			if err != nil {
				stopTimer()
				return
			}
			for i := 0; i < n; i++ {
				if events[i].Filter == unix.EVFILT_VNODE && (events[i].Fflags&(unix.NOTE_WRITE|unix.NOTE_ATTRIB) != 0) {
					scheduleReload()
					break
				}
			}
			select {
			case <-ctx.Done():
				stopTimer()
				return
			case <-quit:
				stopTimer()
				return
			default:
			}
		}
	}()

	stop := func() error {
		stopTimer()
		close(quit)
		return nil
	}
	return stop, done, nil
}
