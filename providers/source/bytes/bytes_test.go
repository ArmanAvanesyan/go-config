package bytes_test

import (
	"context"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	sourcebytes "github.com/ArmanAvanesyan/go-config/providers/source/bytes"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

func TestBytesSource(t *testing.T) {
	t.Parallel()

	raw := []byte(`{"key":"value"}`)
	src := sourcebytes.New("inline", "json", raw)

	v, err := src.Read(context.Background())
	testutil.RequireNoError(t, err)

	doc, ok := v.(*config.Document)
	if !ok {
		t.Fatalf("expected *config.Document, got %T", v)
	}

	testutil.RequireEqual(t, "inline", doc.Name)
	testutil.RequireEqual(t, "json", doc.Format)
	testutil.RequireEqual(t, raw, doc.Raw)
}
