package memory_test

import (
	"context"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/providers/source/memory"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

func TestMemorySource(t *testing.T) {
	t.Parallel()

	tree := map[string]any{"host": "localhost", "port": float64(8080)}
	src := memory.New(tree)

	v, err := src.Read(context.Background())
	testutil.RequireNoError(t, err)

	doc, ok := v.(*config.TreeDocument)
	if !ok {
		t.Fatalf("expected *config.TreeDocument, got %T", v)
	}

	testutil.RequireEqual(t, "memory", doc.Name)
	testutil.RequireEqual(t, tree, doc.Tree)
}

func TestMemorySource_Named(t *testing.T) {
	t.Parallel()

	tree := map[string]any{"key": "val"}
	src := memory.Named("defaults", tree)

	v, err := src.Read(context.Background())
	testutil.RequireNoError(t, err)

	doc := v.(*config.TreeDocument)
	testutil.RequireEqual(t, "defaults", doc.Name)
}
