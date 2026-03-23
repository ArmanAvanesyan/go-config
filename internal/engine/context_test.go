package engine

import (
	"context"
	"testing"

	"github.com/ArmanAvanesyan/go-config/providers/merge/deep"
)

func TestPipelineContext(t *testing.T) {
	t.Parallel()
	pc := newPipelineContext(context.Background(), deep.New())
	if pc.ctx == nil || pc.strategy == nil {
		t.Fatalf("pipeline context not initialized: %+v", pc)
	}
}
