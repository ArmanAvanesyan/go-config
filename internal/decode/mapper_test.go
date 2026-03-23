package decode

import (
	"reflect"
	"testing"
)

func TestDecodeValue_StructMapSliceArrayAndMapErrors(t *testing.T) {
	t.Parallel()
	type child struct {
		V int `json:"v"`
	}
	type top struct {
		Name  string           `json:"name"`
		Child *child           `json:"child"`
		M     map[int]string   `json:"m"`
		S     []string         `json:"s"`
		A     [2]int           `json:"a"`
		Iface any              `json:"iface"`
		MS    map[string]child `json:"ms"`
	}
	in := map[string]any{
		"name":  "n",
		"child": map[string]any{"v": 8},
		"m":     map[string]any{"1": "x"},
		"s":     []string{"a", "b"},
		"a":     []any{1.0, 2.0, 3.0},
		"iface": "raw",
		"ms":    map[string]any{"k": map[string]any{"v": 1.0}},
	}
	var out top
	if err := Decode(in, &out, Options{WeaklyTypedInput: true}); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if out.Child == nil || out.Child.V != 8 {
		t.Fatalf("pointer child decode failed: %+v", out.Child)
	}
	if out.M[1] != "x" || len(out.S) != 2 || out.A[1] != 2 || out.Iface != "raw" || out.MS["k"].V != 1 {
		t.Fatalf("decoded values unexpected: %+v", out)
	}

	var badMap struct {
		M map[string]int `json:"m"`
	}
	if err := Decode(map[string]any{"m": 1}, &badMap, Options{}); err == nil {
		t.Fatal("expected map type error")
	}
	var badSlice struct {
		S []int `json:"s"`
	}
	if err := Decode(map[string]any{"s": 1}, &badSlice, Options{}); err == nil {
		t.Fatal("expected slice type error")
	}
	var badArray struct {
		A [1]int `json:"a"`
	}
	if err := Decode(map[string]any{"a": 1}, &badArray, Options{}); err == nil {
		t.Fatal("expected array type error")
	}
}

func TestDecodeStruct_ErrorUnusedAndCaseInsensitive(t *testing.T) {
	t.Parallel()
	type cfg struct {
		Foo string `json:"foo"`
	}
	var out cfg
	if err := Decode(map[string]any{"FOO": "x"}, &out, Options{}); err != nil || out.Foo != "x" {
		t.Fatalf("case-insensitive decode failed err=%v out=%+v", err, out)
	}
	if err := Decode(map[string]any{"foo": "x", "extra": 1}, &out, Options{ErrorUnused: true}); err == nil {
		t.Fatal("expected unknown field error")
	}
}

func TestDecodeValue_CannotSet(t *testing.T) {
	t.Parallel()
	v := reflect.ValueOf(struct{ A int }{A: 1}).Field(0)
	if err := decodeValue(v, 2, Options{}, "root.a"); err == nil {
		t.Fatal("expected cannot set error")
	}
}

func TestDecodeValue_NilAndInterfaceBranches(t *testing.T) {
	t.Parallel()
	var p *int
	rvPtr := reflect.ValueOf(&p).Elem()
	if err := decodeValue(rvPtr, nil, Options{}, "root.p"); err != nil {
		t.Fatalf("decode nil pointer failed: %v", err)
	}
	if p != nil {
		t.Fatal("nil source should zero pointer destination")
	}

	var iface any
	rvIface := reflect.ValueOf(&iface).Elem()
	if err := decodeValue(rvIface, "x", Options{}, "root.i"); err != nil {
		t.Fatalf("decode interface failed: %v", err)
	}
	if iface != "x" {
		t.Fatalf("interface value mismatch: %#v", iface)
	}
}

func TestDecodeStruct_NonMapFallsBackToScalarError(t *testing.T) {
	t.Parallel()
	type C struct{ A int }
	var out C
	if err := Decode(map[string]any{"A": "not-int"}, &out, Options{WeaklyTypedInput: false}); err == nil {
		t.Fatal("expected decode error for non-map scalar assignment into struct field")
	}
}

func TestDecodeMap_KeyConversionError(t *testing.T) {
	t.Parallel()
	type cfg struct {
		M map[bool]int `json:"m"`
	}
	var out cfg
	if err := Decode(map[string]any{"m": map[string]any{"not-bool": 1}}, &out, Options{}); err == nil {
		t.Fatal("expected key conversion error")
	}
}

func TestDecodeSliceAndArray_ElementErrorPaths(t *testing.T) {
	t.Parallel()
	type cfg struct {
		S []int  `json:"s"`
		A [2]int `json:"a"`
	}
	var out cfg
	if err := Decode(map[string]any{"s": []any{"x"}}, &out, Options{}); err == nil {
		t.Fatal("expected slice element decode error")
	}
	if err := Decode(map[string]any{"a": []any{"x"}}, &out, Options{}); err == nil {
		t.Fatal("expected array element decode error")
	}
}
