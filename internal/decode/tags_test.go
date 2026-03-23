package decode

import (
	"reflect"
	"testing"
)

func TestFieldKeyAndCachedFields(t *testing.T) {
	t.Parallel()
	type T struct {
		Skip string `mapstructure:"-"`
		MS   string `mapstructure:"ms"`
		JS   string `json:"js"`
		JS2  string `json:",omitempty"`
		Def  string
	}
	f0 := reflect.TypeOf(T{}).Field(0)
	if fieldKey(f0) != "-" {
		t.Fatalf("expected - field key")
	}
	f1 := reflect.TypeOf(T{}).Field(1)
	if fieldKey(f1) != "ms" {
		t.Fatalf("expected mapstructure key")
	}
	f2 := reflect.TypeOf(T{}).Field(2)
	if fieldKey(f2) != "js" {
		t.Fatalf("expected json key")
	}
	f3 := reflect.TypeOf(T{}).Field(3)
	if fieldKey(f3) != "JS2" {
		t.Fatalf("expected field name fallback from empty json tag")
	}
	f4 := reflect.TypeOf(T{}).Field(4)
	if fieldKey(f4) != "Def" {
		t.Fatalf("expected field name fallback")
	}
	fields := cachedFields(reflect.TypeOf(T{}))
	if len(fields) != 4 {
		t.Fatalf("cachedFields len=%d want 4", len(fields))
	}
	_ = cachedFields(reflect.TypeOf(T{}))
}
