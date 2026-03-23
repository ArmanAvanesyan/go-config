package schemautil

import (
	"reflect"
	"testing"
)

func TestClassifyScalar(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		in   reflect.Type
		want string
	}{
		{name: "nil", in: nil, want: ""},
		{name: "bool", in: reflect.TypeOf(true), want: "boolean"},
		{name: "int", in: reflect.TypeOf(int64(1)), want: "integer"},
		{name: "uint", in: reflect.TypeOf(uint32(1)), want: "integer"},
		{name: "float", in: reflect.TypeOf(float64(1)), want: "number"},
		{name: "string", in: reflect.TypeOf("x"), want: "string"},
		{name: "ptr scalar", in: reflect.TypeOf(new(int)), want: "integer"},
		{name: "slice unsupported", in: reflect.TypeOf([]int{}), want: ""},
		{name: "map unsupported", in: reflect.TypeOf(map[string]int{}), want: ""},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := ClassifyScalar(tc.in); got != tc.want {
				t.Fatalf("ClassifyScalar(%v)=%q want %q", tc.in, got, tc.want)
			}
		})
	}
}
