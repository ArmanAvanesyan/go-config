package config

import "testing"

func TestErrorSentinelsDistinct(t *testing.T) {
	t.Parallel()
	errs := []error{
		ErrNilTarget, ErrNoSources, ErrUnknownFormat, ErrParserRequired, ErrDecoderRequired,
		ErrInvalidDocument, ErrValidationFailed, ErrResolutionFailed, ErrDecodeFailed,
		ErrSourceReadFailed, ErrParseFailed, ErrMergeFailed,
	}
	for i := range errs {
		if errs[i] == nil {
			t.Fatalf("nil sentinel at %d", i)
		}
	}
}
