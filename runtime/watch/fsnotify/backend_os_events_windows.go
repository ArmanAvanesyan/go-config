//go:build windows && ignore

// Synchronous ReadDirectoryChangesW blocks; CloseHandle from another goroutine
// can deadlock. Use polling backend on Windows (see backend_polling.go).

package fsnotify

import (
	"context"
	"path/filepath"
	"sync"
	"time"
	"unicode/utf16"

	"golang.org/x/sys/windows"
)

const (
	fileListDirectory         = 0x00000001
	fileShareRead             = 0x00000001
	openExisting              = 3
	fileFlagBackupSemantics   = 0x02000000
	fileNotifyChangeFileName  = 0x00000001
	fileNotifyChangeLastWrite = 0x00000010
	fileNotifyChangeSize      = 0x00000008
)

type osEventsBackend struct{}

func newBackend() backend {
	return &osEventsBackend{}
}

func (b *osEventsBackend) start(ctx context.Context, paths []string, debounce time.Duration, onReload func()) (func() error, <-chan struct{}, error) {
	dirToNames := make(map[string]map[string]struct{})
	for _, p := range paths {
		dir := filepath.Dir(p)
		name := filepath.Base(p)
		if dirToNames[dir] == nil {
			dirToNames[dir] = make(map[string]struct{})
		}
		dirToNames[dir][name] = struct{}{}
	}

	var handles []windows.Handle
	var dirs []string
	for dir := range dirToNames {
		pathPtr, err := windows.UTF16PtrFromString(dir)
		if err != nil {
			for _, h := range handles {
				windows.CloseHandle(h)
			}
			return nil, nil, err
		}
		h, err := windows.CreateFile(
			pathPtr,
			fileListDirectory,
			fileShareRead,
			nil,
			openExisting,
			fileFlagBackupSemantics,
			0,
		)
		if err != nil {
			for _, c := range handles {
				windows.CloseHandle(c)
			}
			return nil, nil, err
		}
		handles = append(handles, h)
		dirs = append(dirs, dir)
	}

	reloadCh := make(chan struct{}, 1)
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

	namesByDir := dirToNames
	var wg sync.WaitGroup
	wg.Add(len(handles))

	for i, h := range handles {
		watchNames := namesByDir[dirs[i]]
		go func(handle windows.Handle, watchNames map[string]struct{}) {
			defer wg.Done()
			buf := make([]byte, 64*1024)
			for {
				var n uint32
				err := windows.ReadDirectoryChanges(
					handle,
					&buf[0],
					uint32(len(buf)),
					false,
					fileNotifyChangeFileName|fileNotifyChangeLastWrite|fileNotifyChangeSize,
					&n,
					nil,
					0,
				)
				if err != nil {
					return
				}
				select {
				case <-ctx.Done():
					return
				case <-quit:
					return
				default:
				}
				offset := uint32(0)
				for {
					if offset >= n {
						break
					}
					nextOffset := uint32(buf[offset]) | uint32(buf[offset+1])<<8 | uint32(buf[offset+2])<<16 | uint32(buf[offset+3])<<24
					fileNameLength := uint32(buf[offset+8]) | uint32(buf[offset+9])<<8 | uint32(buf[offset+10])<<16 | uint32(buf[offset+11])<<24
					fileNameOffset := offset + 12
					if fileNameLength > 0 && fileNameOffset+fileNameLength <= n {
						utf16Name := make([]uint16, fileNameLength/2)
						for j := uint32(0); j < fileNameLength/2 && fileNameOffset+j*2+1 < n; j++ {
							utf16Name[j] = uint16(buf[fileNameOffset+j*2]) | uint16(buf[fileNameOffset+j*2+1])<<8
						}
						name := string(utf16.Decode(utf16Name))
						if _, ok := watchNames[name]; ok {
							select {
							case reloadCh <- struct{}{}:
							default:
							}
							break
						}
					}
					if nextOffset == 0 {
						break
					}
					offset = nextOffset
				}
			}
		}(h, watchNames)
	}

	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				stopTimer()
				go func() {
					for _, h := range handles {
						windows.CloseHandle(h)
					}
				}()
				wg.Wait()
				return
			case <-quit:
				stopTimer()
				go func() {
					for _, h := range handles {
						windows.CloseHandle(h)
					}
				}()
				wg.Wait()
				return
			case <-reloadCh:
				scheduleReload()
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
