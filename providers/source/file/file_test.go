package file_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/providers/source/file"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

// testdataPath returns an absolute path to the repo-level testdata directory.
func testdataPath(name string) string {
	_, thisFile, _, _ := runtime.Caller(0)
	// thisFile: .../providers/source/file/file_test.go
	// Dir:      .../providers/source/file
	// up 3:     .../go-config (repo root)
	root := filepath.Join(filepath.Dir(thisFile), "..", "..", "..")
	return filepath.Join(root, "testdata", name)
}

func TestFileSource_Read(t *testing.T) {
	t.Parallel()

	path := testdataPath("basic.json")
	src := file.New(path)

	v, err := src.Read(context.Background())
	testutil.RequireNoError(t, err)

	doc, ok := v.(*config.Document)
	if !ok {
		t.Fatalf("expected *config.Document, got %T", v)
	}

	testutil.RequireEqual(t, path, doc.Name)
	testutil.RequireEqual(t, "json", doc.Format)

	if len(doc.Raw) == 0 {
		t.Fatal("expected non-empty raw bytes")
	}
}

func TestFileSource_NotFound(t *testing.T) {
	t.Parallel()

	src := file.New("/does/not/exist/config.json")
	_, err := src.Read(context.Background())
	testutil.RequireError(t, err)
}
