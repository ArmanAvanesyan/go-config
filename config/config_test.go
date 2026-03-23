package config

import (
	"reflect"
	"testing"
)

func TestNew_StrictDecoderSelection(t *testing.T) {
	t.Parallel()
	l1 := New(WithStrict(true))
	if reflect.TypeOf(l1.options.Decoder).String() == reflect.TypeOf((&noOpDecoder{})).String() {
		t.Fatal("strict should select strict decoder when not explicitly set")
	}

	custom := &noOpDecoder{}
	l2 := New(WithStrict(true), WithDecoder(custom))
	if l2.options.Decoder != custom {
		t.Fatal("custom decoder should not be overridden by strict")
	}
}
