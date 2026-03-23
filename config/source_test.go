package config

import (
	"context"
	"testing"
)

type docSrc struct {
	v any
	e error
}

func (s *docSrc) Read(context.Context) (any, error) { return s.v, s.e }

func TestSourceInterface(t *testing.T) {
	t.Parallel()
	var _ Source = (*docSrc)(nil)
}
