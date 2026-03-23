//go:build linux

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
	fd, err := unix.InotifyInit1(unix.IN_NONBLOCK | unix.IN_CLOEXEC)
	if err != nil {
		return nil, nil, err
	}
	for _, p := range paths {
		_, err := unix.InotifyAddWatch(fd, p, unix.IN_MODIFY|unix.IN_ATTRIB|unix.IN_CLOSE_WRITE|unix.IN_CREATE)
		if err != nil {
			_ = unix.Close(fd)
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
		defer close(done)
		defer func() { _ = closeFD() }()
		buf := make([]byte, unix.SizeofInotifyEvent*128+unix.NAME_MAX+1)
		for {
			n, err := unix.Read(fd, buf)
			if err != nil {
				if err == unix.EAGAIN {
					select {
					case <-ctx.Done():
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
				raw := buf[i : i+unix.SizeofInotifyEvent]
				wd := int(raw[0]) | int(raw[1])<<8 | int(raw[2])<<16 | int(raw[3])<<24
				mask := uint32(raw[4]) | uint32(raw[5])<<8 | uint32(raw[6])<<16 | uint32(raw[7])<<24
				_ = wd
				if mask&(unix.IN_MODIFY|unix.IN_ATTRIB|unix.IN_CLOSE_WRITE|unix.IN_CREATE) != 0 {
					scheduleReload()
				}
				evLen := int(raw[8]) | int(raw[9])<<8 | int(raw[10])<<16 | int(raw[11])<<24
				if evLen <= 0 {
					evLen = unix.SizeofInotifyEvent
				}
				i += unix.SizeofInotifyEvent + evLen
			}
		}
	}()

	stop := func() error {
		stopTimer()
		return closeFD()
	}
	return stop, done, nil
}
