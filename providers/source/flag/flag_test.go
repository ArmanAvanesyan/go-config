package flag_test

import (
	"context"
	goflag "flag"
	"testing"

	"github.com/ArmanAvanesyan/go-config/config"
	"github.com/ArmanAvanesyan/go-config/providers/source/flag"
	"github.com/ArmanAvanesyan/go-config/testutil"
)

func TestFlagSource_ExplicitlySetFlags(t *testing.T) {
	t.Parallel()

	fs := goflag.NewFlagSet("test", goflag.ContinueOnError)
	fs.String("host", "default-host", "server host")
	fs.Int("port", 8080, "server port")

	// Only parse --host; port keeps its default and should NOT appear.
	if err := fs.Parse([]string{"--host", "myhost"}); err != nil {
		t.Fatalf("flag parse failed: %v", err)
	}

	src := flag.NewFromFlagSet(fs)
	v, err := src.Read(context.Background())
	testutil.RequireNoError(t, err)

	doc, ok := v.(*config.TreeDocument)
	if !ok {
		t.Fatalf("expected *config.TreeDocument, got %T", v)
	}

	testutil.RequireEqual(t, "flags", doc.Name)

	if doc.Tree["host"] != "myhost" {
		t.Fatalf("expected host=myhost, got %v", doc.Tree["host"])
	}
	if _, ok := doc.Tree["port"]; ok {
		t.Fatal("port was not explicitly set; should not appear in tree")
	}
}

func TestFlagSource_NoFlagsSet(t *testing.T) {
	t.Parallel()

	fs := goflag.NewFlagSet("empty", goflag.ContinueOnError)
	fs.String("host", "default", "host")

	// Parse with no args — nothing is explicitly set.
	_ = fs.Parse(nil)

	src := flag.NewFromFlagSet(fs)
	v, err := src.Read(context.Background())
	testutil.RequireNoError(t, err)

	doc := v.(*config.TreeDocument)
	if len(doc.Tree) != 0 {
		t.Fatalf("expected empty tree, got %#v", doc.Tree)
	}
}
