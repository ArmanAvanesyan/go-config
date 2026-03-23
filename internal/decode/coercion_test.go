package decode

import (
	"reflect"
	"testing"
)

func TestAssignScalar_Table(t *testing.T) {
	t.Parallel()
	type sample struct {
		S string
		B bool
		I int8
		U uint8
		F float32
	}
	makeDst := func(name string) (reflect.Value, *sample) {
		x := &sample{}
		rv := reflect.ValueOf(x).Elem().FieldByName(name)
		return rv, x
	}
	cases := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "assignable and convertible",
			fn: func(t *testing.T) {
				dst, x := makeDst("S")
				if err := assignScalar(dst, "ok", Options{}, "p"); err != nil || x.S != "ok" {
					t.Fatalf("assign failed err=%v val=%q", err, x.S)
				}
			},
		},
		{
			name: "weak string formatting",
			fn: func(t *testing.T) {
				dst, x := makeDst("S")
				if err := assignScalar(dst, 12, Options{WeaklyTypedInput: true}, "p"); err != nil || x.S != "12" {
					t.Fatalf("weak string failed err=%v val=%q", err, x.S)
				}
			},
		},
		{
			name: "bool parse weak",
			fn: func(t *testing.T) {
				dst, x := makeDst("B")
				if err := assignScalar(dst, " true ", Options{WeaklyTypedInput: true}, "p"); err != nil || !x.B {
					t.Fatalf("bool parse failed err=%v val=%v", err, x.B)
				}
			},
		},
		{
			name: "int overflow",
			fn: func(t *testing.T) {
				dst, _ := makeDst("I")
				if err := assignScalar(dst, "200", Options{WeaklyTypedInput: true}, "p"); err == nil {
					t.Fatal("expected int overflow error")
				}
			},
		},
		{
			name: "uint negative rejected",
			fn: func(t *testing.T) {
				dst, _ := makeDst("U")
				if err := assignScalar(dst, "-1", Options{WeaklyTypedInput: true}, "p"); err == nil {
					t.Fatal("expected negative uint assignment error")
				}
			},
		},
		{
			name: "float overflow",
			fn: func(t *testing.T) {
				dst, _ := makeDst("F")
				if err := assignScalar(dst, "1e100", Options{WeaklyTypedInput: true}, "p"); err == nil {
					t.Fatal("expected float overflow error")
				}
			},
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, tc.fn)
	}
}

func TestToConverters_Table(t *testing.T) {
	t.Parallel()
	if v, ok := toBool(true, false); !ok || !v {
		t.Fatal("bool direct conversion failed")
	}
	if _, ok := toBool("x", true); ok {
		t.Fatal("invalid bool should fail")
	}
	if _, ok := toBool("true", false); ok {
		t.Fatal("strict mode should not parse bool strings")
	}

	intCases := []any{int(1), int8(1), int16(1), int32(1), int64(1), uint(1), uint8(1), uint16(1), uint32(1), float32(1), float64(1)}
	for _, in := range intCases {
		if v, ok := toInt64(in, false); !ok || v != 1 {
			t.Fatalf("toInt64 failed for %T", in)
		}
	}
	if v, ok := toInt64("10", true); !ok || v != 10 {
		t.Fatalf("toInt64 weak parse failed: %v %v", v, ok)
	}
	if _, ok := toInt64("bad", true); ok {
		t.Fatal("invalid int parse should fail")
	}
	if _, ok := toInt64("10", false); ok {
		t.Fatal("strict mode should not parse int strings")
	}
	if v, ok := toInt64(uint64(1), true); !ok || v != 1 {
		t.Fatal("uint64 small value should convert to int64")
	}
	if _, ok := toInt64(uint64(^uint64(0)), true); ok {
		t.Fatal("uint64 > max int64 should fail")
	}
	uintCases := []any{uint(1), uint8(1), uint16(1), uint32(1), uint64(1), int(1), int8(1), int16(1), int32(1), int64(1), float32(1), float64(1)}
	for _, in := range uintCases {
		if v, ok := toUint64(in, false); !ok || v != 1 {
			t.Fatalf("toUint64 failed for %T", in)
		}
	}
	if _, ok := toUint64(-1, true); ok {
		t.Fatal("negative uint should fail")
	}
	if v, ok := toUint64("7", true); !ok || v != 7 {
		t.Fatalf("toUint64 weak parse failed: %v %v", v, ok)
	}
	if _, ok := toUint64("bad", true); ok {
		t.Fatal("invalid uint parse should fail")
	}
	if _, ok := toUint64("7", false); ok {
		t.Fatal("strict mode should not parse uint strings")
	}
	floatCases := []any{float32(1), float64(1), int(1), int8(1), int16(1), int32(1), int64(1), uint(1), uint8(1), uint16(1), uint32(1), uint64(1)}
	for _, in := range floatCases {
		if v, ok := toFloat64(in, false); !ok || v != 1 {
			t.Fatalf("toFloat64 failed for %T", in)
		}
	}
	if v, ok := toFloat64("3.14", true); !ok || v <= 3.1 || v >= 3.2 {
		t.Fatalf("toFloat64 weak parse failed: %v %v", v, ok)
	}
	if _, ok := toFloat64("bad", true); ok {
		t.Fatal("invalid float parse should fail")
	}
	if _, ok := toFloat64("3.14", false); ok {
		t.Fatal("strict mode should not parse float strings")
	}
}

func TestAssignScalar_StrictStringFailsForNonStringInput(t *testing.T) {
	t.Parallel()
	type sample struct{ S string }
	x := &sample{}
	dst := reflect.ValueOf(x).Elem().FieldByName("S")
	if err := assignScalar(dst, 1, Options{WeaklyTypedInput: false}, "root.s"); err == nil {
		t.Fatal("expected strict mode string assignment failure")
	}
}

func TestAssignScalar_KindSuccessCases(t *testing.T) {
	t.Parallel()
	type sample struct {
		B bool
		I int
		U uint
		F float64
	}
	x := &sample{}
	rv := reflect.ValueOf(x).Elem()
	if err := assignScalar(rv.FieldByName("B"), true, Options{}, "b"); err != nil {
		t.Fatalf("bool assign failed: %v", err)
	}
	if err := assignScalar(rv.FieldByName("I"), 1.0, Options{}, "i"); err != nil {
		t.Fatalf("int assign failed: %v", err)
	}
	if err := assignScalar(rv.FieldByName("U"), 1.0, Options{}, "u"); err != nil {
		t.Fatalf("uint assign failed: %v", err)
	}
	if err := assignScalar(rv.FieldByName("F"), 1, Options{}, "f"); err != nil {
		t.Fatalf("float assign failed: %v", err)
	}
}

func TestToUint64_NegativeVariants(t *testing.T) {
	t.Parallel()
	negatives := []any{int(-1), int8(-1), int16(-1), int32(-1), int64(-1), float32(-1), float64(-1)}
	for _, n := range negatives {
		if _, ok := toUint64(n, false); ok {
			t.Fatalf("expected negative rejection for %T", n)
		}
	}
}
