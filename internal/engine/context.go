package engine

import (
	"context"

	"github.com/ArmanAvanesyan/go-config/providers/merge"
)

type pipelineContext struct {
	ctx      context.Context
	strategy merge.Strategy
}

func newPipelineContext(ctx context.Context, strategy merge.Strategy) pipelineContext {
	return pipelineContext{
		ctx:      ctx,
		strategy: strategy,
	}
}
