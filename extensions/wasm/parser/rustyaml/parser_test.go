//go:build integration

package rustyaml

import (
	"context"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

func TestParserSharedLifecycle_Integration(t *testing.T) {
	ctx := context.Background()
	raw := []byte("app:\n  name: demo\n")

	p1, err := NewShared(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "invalid magic number") || strings.Contains(err.Error(), "compile module") {
			t.Skip("WASM binary is a placeholder — run 'cd rust && make all' first")
		}
		testutil.RequireNoError(t, err)
	}

	p2, err := NewShared(ctx)
	testutil.RequireNoError(t, err)

	_, err = p1.Parse(ctx, &config.Document{Name: "a.yaml", Format: "yaml", Raw: raw})
	testutil.RequireNoError(t, err)
	_, err = p2.Parse(ctx, &config.Document{Name: "b.yaml", Format: "yaml", Raw: raw})
	testutil.RequireNoError(t, err)

	testutil.RequireNoError(t, ReleaseShared(ctx))
	testutil.RequireNoError(t, ReleaseShared(ctx))
}

func TestParserSharedParallelParse_Integration(t *testing.T) {
	ctx := context.Background()
	raw := []byte("app:\n  name: demo\nserver:\n  host: localhost\n  port: 8080\n")

	p, err := NewShared(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "invalid magic number") || strings.Contains(err.Error(), "compile module") {
			t.Skip("WASM binary is a placeholder — run 'cd rust && make all' first")
		}
		testutil.RequireNoError(t, err)
	}
	defer func() { _ = ReleaseShared(ctx) }()

	var wg sync.WaitGroup
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_, e := p.Parse(ctx, &config.Document{Name: "parallel.yaml", Format: "yaml", Raw: raw})
				if e != nil {
					t.Errorf("parse failed: %v", e)
					return
				}
			}
		}()
	}
	wg.Wait()
}
